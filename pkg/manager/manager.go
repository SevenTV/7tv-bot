package manager

import (
	"github.com/seventv/twitch-irc-reader/pkg/irc"
	"strings"
	"sync"
	"time"
)

// IRCManager manages multiple IRC connections & keeps track of their connected channels
type IRCManager struct {
	user, oauth string

	connectionCounter uint

	// map with connections mapped to incremental uint, we don't use a slice here since removing an element would change the keys after it
	connections map[uint]*connection
	channels    map[string]*ircChannel

	mx *sync.Mutex

	onMessage func(*irc.Message, error)
}

// New returns a new IRC manager, set up with the passed authentication.
// Requires you to call .OnMessage() before you .Join() a channel
func New(user, oauth string) *IRCManager {
	return &IRCManager{
		user:  user,
		oauth: oauth,

		connections: make(map[uint]*connection),
		channels:    make(map[string]*ircChannel),

		mx: &sync.Mutex{},
	}
}

// TODO: implement Disconnect()

// TODO: implement Part()

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
	m.mx.Lock()
	defer m.mx.Unlock()
	// if channel is already joined, return error
	if _, found := m.channels[strings.ToLower(channelName)]; found {
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
	// TODO: mutex unlock here? So we can call Join() again without having to wait for new connections

	return m.connections[connectionKey].join(channel)
}

func (m *IRCManager) findConnectionWithCapacity(weight int) uint {
	for k, v := range m.connections {
		if v.hasCapacity(weight) {
			return k
		}
	}

	return 0
}

func (m *IRCManager) addNewConnection() uint {
	// connectionCounter is incremented before its value is read, so 0 can be used in findConnectionWithCapacity
	m.connectionCounter++

	con := newConnection(m.user, m.oauth)
	con.setOnMessage(m.onMessage)

	m.connections[m.connectionCounter] = con

	// create worker to keep connection active
	// TODO: this needs to be revised, just a quick solution to start testing
	go func() {
		for {
			err := con.connect()
			// TODO: feed back channels to user on disconnect & delete connection in m.connections, instead of reusing the dead connection, so we don't exceed JOIN ratelimit
			if err == irc.ErrServerDisconnect {
				// if we were disconnected by the server, wait 5 seconds and try again
				<-time.NewTimer(5 * time.Second).C
				continue
			}
			// end worker, if we reach this, it means the connection has been ended deliberately
			break
		}
	}()

	return m.connectionCounter
}
