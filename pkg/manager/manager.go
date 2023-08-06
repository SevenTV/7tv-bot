package manager

import (
	"github.com/seventv/twitch-irc-reader/pkg/irc"
	"strings"
	"sync"
)

// IRCManager manages multiple IRC connections & keeps track of their connected channels
type IRCManager struct {
	user, oauth string

	connectionCounter uint

	// map with connections mapped to incremental uint, we don't use a slice here since removing an element would change the keys after it
	connections map[uint]*connection
	channels    map[string]*ircChannel

	wg *sync.WaitGroup
	mx *sync.Mutex

	isClosing bool

	// OrphanedChannels is a channel that sends a queue of IRC channels have lost their parent connection,
	// without explicitly having called Part or Disconnect.
	//
	// This channel MUST be received!
	OrphanedChannels chan *ircChannel
	partedChannels   chan *ircChannel

	onMessage func(*irc.Message, error)
}

// New returns a new IRC manager, set up with the passed authentication.
//
// Requires you to call .OnMessage(), and listen to the OrphanedChannels channel, before you .Join() a channel
func New(user, oauth string) *IRCManager {
	return &IRCManager{
		user:  user,
		oauth: oauth,

		connections: make(map[uint]*connection),
		channels:    make(map[string]*ircChannel),

		OrphanedChannels: make(chan *ircChannel),
		partedChannels:   make(chan *ircChannel),

		wg: &sync.WaitGroup{},
		mx: &sync.Mutex{},
	}
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
	go func() {
		for channel := range m.partedChannels {
			m.deleteChannel(channel.name)
		}
	}()
	return nil
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

	m.mx.Lock()
	// if channel is already joined, return error
	if _, found := m.channels[strings.ToLower(channelName)]; found {
		m.mx.Unlock()
		return ErrChanAlreadyJoined
	}

	connectionKey := m.findConnectionWithCapacity(weight)
	// keys only start at 1, so 0 means no suitable connection is available
	if connectionKey == 0 {
		connectionKey = m.addNewConnection()
	}

	channel := newIrcChannel(channelName, weight)
	channel.connectionKey = connectionKey

	m.channels[channel.name] = channel
	// mutex unlock, so we can call Join() again, without having to wait for new connections
	m.mx.Unlock()

	return m.connections[connectionKey].join(channel)
}

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

func (m *IRCManager) addNewConnection() uint {
	// connectionCounter is incremented before its value is read, so 0 can be used in findConnectionWithCapacity
	m.connectionCounter++

	con := newConnection(m.user, m.oauth)
	con.Parted = m.partedChannels
	con.setOnMessage(m.onMessage)

	m.connections[m.connectionCounter] = con

	// create worker
	m.wg.Add(1)
	connectionKey := m.connectionCounter
	go func() {
		defer m.wg.Done()
		err := con.connect()
		if err == irc.ErrServerDisconnect {
			// if we were disconnected by the server, flush all connected channels to the OrphanedChannels channel
			con.flushChannels(m.OrphanedChannels)
		}
		m.deleteConnection(connectionKey)
	}()

	return m.connectionCounter
}

func (m *IRCManager) deleteConnection(key uint) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	conn, ok := m.connections[key]
	if !ok {
		return ErrConnNotFound
	}

	for _, channel := range conn.channels {
		m.deleteChannel(channel.name)
	}

	delete(m.connections, key)

	return nil
}

// findChannel looks for the given channel name in the list of connected irc channels.
// Please call strings.ToLower() before passing along the channelName argument
func (m *IRCManager) findChannel(channelName string) *ircChannel {
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
	delete(m.channels, channel.name)

	return nil
}
