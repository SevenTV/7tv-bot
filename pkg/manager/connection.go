package manager

import (
	"strings"
	"sync"

	"github.com/seventv/7tv-bot/pkg/irc"
)

var (
	// ConnectionCapacity determines how many channels you can JOIN on a single connection,
	// must be set before creating any connections
	ConnectionCapacity = 50
)

// connection helps you manage a single IRC connection using some middleware,
// to enable the IRCManager to manage lots of connections
type connection struct {
	client   *irc.Client
	channels []*IRCChannel
	// avoids a lot of headaches
	channelsMx *sync.Mutex

	// capacity determines how many channels can be joined on a connection
	capacity  int
	onMessage func(msg *irc.Message, err error)

	// isReady means the connection is ready to accept new channels.
	// If this is false, it means the connection is closing or closed
	isReady bool

	// Parted is used to feed back channels we left to the manager, must be set before calling connect
	Parted chan *IRCChannel
}

// newConnection sets up a new connection with capacity as set in ConnectionCapacity
func newConnection(user, oauth string) *connection {
	return &connection{
		client:     irc.New(user, oauth).WithCapabilities(irc.CapTags),
		channels:   []*IRCChannel{},
		channelsMx: &sync.Mutex{},
		capacity:   ConnectionCapacity,
		isReady:    true,
	}
}

func (c *connection) connect() error {
	c.client.OnMessage(c.handleMessages)

	defer func() {
		// set isReady to false after the connection is closed
		c.isReady = false
	}()

	return c.client.Connect()
}

func (c *connection) disconnect() {
	c.client.Disconnect()
}

func (c *connection) join(channel *IRCChannel) error {
	err := c.addChannel(channel)
	if err != nil {
		return err
	}

	// make sure the client is connected
	<-c.client.Connected.C

	c.client.Join(channel.Name)

	return nil
}

func (c *connection) hasCapacity(weight int) bool {
	return c.capacity >= weight
}

func (c *connection) part(channels ...string) {
	c.client.Part(channels...)
}

func (c *connection) setOnMessage(cb func(msg *irc.Message, err error)) {
	c.onMessage = cb
}

// this is middleware, needed to properly handle important incoming system messages like PING, JOIN & PART
func (c *connection) handleMessages(msg *irc.Message, err error) {
	// don't bother running the middleware if there's an error for the message
	if err != nil {
		c.onMessage(msg, err)
		return
	}
	switch msg.GetType() {
	case irc.Ping:
		c.pong(msg)
	case irc.Join:
		c.onJoin(msg)
	case irc.Part:
		c.onPart(msg)
	}
	c.onMessage(msg, err)
}

func (c *connection) pong(msg *irc.Message) {
	c.client.SendString("PONG" + parsePingPayload(msg.String()))
}

func (c *connection) onJoin(msg *irc.Message) {
	// flag joined channels as isJoined = true
	for _, joined := range parseChannels(msg.String()) {
		c.setChannelIsJoined(joined, true)
	}
}

func (c *connection) onPart(msg *irc.Message) {
	// flag parted channels as isJoined = false
	for _, parted := range parseChannels(msg.String()) {
		c.setChannelIsJoined(parted, false)
		c.partChannel(parted)
	}
}

func (c *connection) setChannelIsJoined(joined string, isJoined bool) {
	c.channelsMx.Lock()
	defer c.channelsMx.Unlock()

	for _, ch := range c.channels {
		if joined == ch.Name {
			ch.isJoined = isJoined
			break
		}
	}
}

func (c *connection) addChannel(channel *IRCChannel) error {
	c.channelsMx.Lock()
	defer c.channelsMx.Unlock()

	if !c.hasCapacity(channel.Weight) {
		return ErrNoCapacity
	}
	c.capacity -= channel.Weight
	c.channels = append(c.channels, channel)

	return nil
}

// partChannel removes the given channel from this connection, and sends a signal to Parted
func (c *connection) partChannel(channelName string) {
	c.channelsMx.Lock()

	for i, channel := range c.channels {
		if channel.Name == channelName {
			c.channels[i] = c.channels[len(c.channels)-1]
			c.channels = c.channels[:len(c.channels)-1]

			// give capacity back to connection
			c.capacity += channel.Weight

			// unlock mutex, we are done manipulating the slice
			c.channelsMx.Unlock()

			c.Parted <- channel
			return
		}
	}
	c.channelsMx.Unlock()
}

// flushChannels flushes all channels related to this connection to the passed channel
func (c *connection) flushChannels(ch chan *IRCChannel) {
	for _, channel := range c.channels {
		ch <- channel
	}
}

func parseChannels(data string) []string {
	splits := strings.Split(data, " ")
	if len(splits) == 0 {
		return nil
	}
	// list of users is contained in the last split
	last := splits[len(splits)-1]

	var result []string
	for _, user := range strings.Split(last, ",") {
		// user should always start with #, if this is not the case, something is wrong
		if !strings.HasPrefix(user, "#") {
			continue
		}

		result = append(result, strings.ToLower(strings.TrimLeft(user, "#")))
	}
	return result
}

func parsePingPayload(data string) string {
	return strings.TrimPrefix(data, "PING")
}
