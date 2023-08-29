package manager

import (
	"context"
	"strings"
	"sync"

	"github.com/seventv/twitch-irc-reader/pkg/irc"
	"github.com/seventv/twitch-irc-reader/pkg/util"
)

// IRCManager manages multiple IRC connections & keeps track of their connected channels
type IRCManager struct {
	user, oauth string

	connectionCounter uint

	// map with connections mapped to incremental uint, we don't use a slice here since removing an element would change the keys after it
	connections map[uint]*connection
	// map with channel names mapped to IRCChannel struct, we use this for executing commands such as Part quickly
	channels map[string]*IRCChannel

	wg *sync.WaitGroup
	mx *sync.Mutex

	isClosing bool

	// OrphanedChannels is a channel that sends a queue of IRC channels have lost their parent connection,
	// without explicitly having called Part, Shutdown or Disconnect.
	//
	// This channel MUST be received!
	OrphanedChannels chan *IRCChannel
	partedChannels   chan *IRCChannel

	onMessage func(*irc.Message, error)

	rateLimiter RateLimiter
}

// New returns a new IRC manager, set up with the passed authentication.
//
// Requires you to set OnMessage, and listen to the OrphanedChannels channel, before you call Init
func New(user, oauth string) *IRCManager {
	return &IRCManager{
		user:  user,
		oauth: oauth,

		connections: make(map[uint]*connection),
		channels:    make(map[string]*IRCChannel),

		OrphanedChannels: make(chan *IRCChannel),
		partedChannels:   make(chan *IRCChannel),

		wg: &sync.WaitGroup{},
		mx: &sync.Mutex{},

		rateLimiter: &NoLimit{},
	}
}

// WithLimit adds a rate limiter to an IRCManager
func (m *IRCManager) WithLimit(limiter RateLimiter) *IRCManager {
	m.rateLimiter = limiter
	return m
}

// Init does some basic checks to make sure the IRCManager is ready to use.
//
// It also starts a goroutine for reading channels to keep the service running properly
func (m *IRCManager) Init() error {
	if m.onMessage == nil {
		return ErrOnMessageUnset
	}
	if m.isClosing {
		return ErrManagerClosing
	}

	// Start first connection, so we're ready for the first Join call.
	// Also, so we have at least 1 worker in m.wg before we call Wait()
	m.mx.Lock()
	m.addNewConnection()
	m.mx.Unlock()

	done := &util.Closer{}
	done.Reset()

	// wait for all connections to close, then send a signal to the worker to stop
	go func() {
		m.wg.Wait()
		done.Close()
	}()

	go m.startWorker(done)

	return nil
}

func (m *IRCManager) startWorker(done *util.Closer) {
	for {
		select {
		// stop worker when Shutdown is completed
		case <-done.C:
			close(m.partedChannels)
			return
		// delete channels we Parted from memory
		case channel := <-m.partedChannels:
			m.mx.Lock()
			m.deleteChannel(channel.Name)
			m.mx.Unlock()
		}
	}
}

// Shutdown stops & disconnects ALL connections in the manager, returns a WaitGroup to wait for graceful shutdown
// Calling Shutdown will not send back any previously joined channels to OrphanedChannels.
func (m *IRCManager) Shutdown() *sync.WaitGroup {
	m.isClosing = true
	m.mx.Lock()
	defer m.mx.Unlock()
	for _, conn := range m.connections {
		conn.disconnect()
	}

	return m.wg
}

// Part sends a PART message for the channel you want to leave to the connection that it is connected to
func (m *IRCManager) Part(channelName string) error {
	if m.isClosing {
		return ErrManagerClosing
	}

	m.mx.Lock()
	defer m.mx.Unlock()
	channel := m.findChannel(strings.ToLower(channelName))
	if channel == nil {
		return ErrChanNotFound
	}
	conn, ok := m.connections[channel.connectionKey]
	if !ok {
		return ErrConnNotFound
	}
	conn.part(strings.ToLower(channelName))
	return nil
}

// OnMessage sets a callback, executed on all incoming IRC messages from every connection.
// Must be set before you try to Join channels, not setting this will result in nil pointer!
// The callback will be agnostic of which underlying connection it came from.
func (m *IRCManager) OnMessage(cb func(*irc.Message, error)) {
	m.onMessage = cb
}

// Join finds an IRC connection suitable for the passed channel, if none is found, it starts a new IRC connection.
// Requires the name of the channel you want to JOIN & a weight, so we can avoid putting too many busy channels on the same connection.
// Default max capacity of a connection is 50, any weight value equal to or higher than the max capacity will assign the channel its own unique connection.
func (m *IRCManager) Join(channelName string, weight int) error {
	if m.isClosing {
		return ErrManagerClosing
	}

	err := m.rateLimiter.WaitToJoin(context.TODO())
	if err != nil {
		return err
	}

	m.mx.Lock()
	// if channel is already joined, return error
	if _, found := m.channels[strings.ToLower(channelName)]; found {
		m.mx.Unlock()
		return ErrChanAlreadyJoined
	}

	connectionKey := m.findConnectionWithCapacity(weight)
	// 0 means no suitable connection is available
	if connectionKey == 0 {
		err = m.rateLimiter.WaitToAuth(context.TODO())
		if err != nil {
			return err
		}
		connectionKey = m.addNewConnection()
	}

	channel := newIrcChannel(channelName, weight)
	channel.connectionKey = connectionKey

	m.channels[channel.Name] = channel
	// mutex unlock, so we can call Join() again, without having to wait for new connections
	m.mx.Unlock()

	return m.connections[connectionKey].join(channel)
}

// findConnectionWithCapacity returns the key of a suitable connection, given the weight passed to it.
// returns 0 if none is found
func (m *IRCManager) findConnectionWithCapacity(weight int) uint {
	for k, conn := range m.connections {
		// skip if this connection is not ready for new channels (usually means the connection is closing)
		if !conn.isReady {
			continue
		}
		if conn.hasCapacity(weight) {
			return k
		}
	}

	return 0
}

// addNewConnection starts a new connection & adds it to the manager.
// returns the key for the connection in m.connections
func (m *IRCManager) addNewConnection() uint {
	// connectionCounter is incremented before its value is read, so 0 can be used in findConnectionWithCapacity
	m.connectionCounter++

	// this is for the very unlikely event we overflow the uint type
	if m.connectionCounter == 0 {
		m.connectionCounter++
	}

	con := newConnection(m.user, m.oauth)
	con.Parted = m.partedChannels
	con.setOnMessage(m.onMessage)

	m.connections[m.connectionCounter] = con

	// create worker
	m.wg.Add(1)
	go m.startConnection(con, m.connectionCounter)

	return m.connectionCounter
}

func (m *IRCManager) startConnection(con *connection, connectionKey uint) {
	defer m.wg.Done()
	err := con.connect()
	if err == irc.ErrServerDisconnect {
		// if we were disconnected by the server, flush all connected channels to the OrphanedChannels channel
		con.flushChannels(m.OrphanedChannels)
	}
	m.deleteConnection(connectionKey)
}

// deleteConnection deletes a connection and all channels related to it
func (m *IRCManager) deleteConnection(key uint) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	conn, ok := m.connections[key]
	if !ok {
		return ErrConnNotFound
	}

	for _, channel := range conn.channels {
		m.deleteChannel(channel.Name)
	}

	delete(m.connections, key)

	return nil
}

// findChannel looks for the given channel name in the list of connected irc channels.
// Please call strings.ToLower() before passing along the channelName argument
func (m *IRCManager) findChannel(channelName string) *IRCChannel {
	channel, found := m.channels[channelName]
	if !found {
		return nil
	}

	return channel
}

// findChannel looks for the given channel name in the list of connected irc channels and deletes it.
// Please call strings.ToLower() before passing along the channelName argument
func (m *IRCManager) deleteChannel(channelName string) error {
	channel, found := m.channels[channelName]
	if !found {
		return ErrChanNotFound
	}
	delete(m.channels, channel.Name)

	return nil
}
