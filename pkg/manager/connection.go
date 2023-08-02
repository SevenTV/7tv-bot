package manager

import (
	"github.com/seventv/twitch-irc-reader/pkg/irc"
	"strings"
)

var (
	// ConnectionCapacity determines how many channels you can JOIN on a single connection,
	// must be set before creating any connections
	ConnectionCapacity = 50

	// PartedBuffer determines the buffer size for the OnPart channel
	PartedBuffer = 10
)

// connection helps you manage a single IRC connection using some middleware,
// to enable the IRCManager to manage lots of connections
type connection struct {
	client   *irc.Client
	channels []*ircChannel
	// capacity determines how many channels can be joined on a connection
	capacity  int
	onMessage func(msg *irc.Message, err error)
	OnPart    chan *ircChannel
}

// newConnection sets up a new connection with capacity as set in ConnectionCapacity
func newConnection(user, oauth string) *connection {
	return &connection{
		client:   irc.New(user, oauth).WithCapabilities(irc.CapTags),
		channels: []*ircChannel{},
		capacity: ConnectionCapacity,
		OnPart:   make(chan *ircChannel, PartedBuffer),
	}
}

func (c *connection) connect() error {
	c.client.OnMessage(c.handleMessages)
	return c.client.Connect()
}

func (c *connection) disconnect() {
	// TODO: implement, should make a way to pass parted channels back to the manager after
}

func (c *connection) join(channel *ircChannel) error {
	if c.capacity < channel.weight {
		return ErrNoCapacity
	}
	c.capacity -= channel.weight
	c.channels = append(c.channels, channel)

	c.client.Join(channel.name)

	return nil
}

func (c *connection) part() {
	// TODO: implement, should make a way to pass parted channels back to the manager after
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
	c.setChannelsJoined(msg.String(), true)
}

func (c *connection) onPart(msg *irc.Message) {
	// TODO: implement logic for properly removing channel or attempt reconnect (signal back to manager?)
	// flag parted channels as isJoined = false
	c.setChannelsJoined(msg.String(), false)
}

func (c *connection) setChannelsJoined(data string, isJoined bool) {
	for _, joined := range parseChannels(data) {
		for _, ch := range c.channels {
			if joined == ch.name {
				ch.isJoined = isJoined
				break
			}
		}
	}
}

func parseChannels(data string) []string {
	splits := strings.Split(data, " ")
	if len(splits) == 0 {
		return nil
	}
	// list of users is contained in the last split
	last := splits[len(splits)-1]

	result := []string{}
	for _, user := range strings.Split(last, ",") {
		// user should always start with #, if this is not the case, something is wrong
		if !strings.HasPrefix(user, "#") {
			continue
		}

		result = append(result, strings.TrimLeft(user, "#"))
	}
	return result
}

func parsePingPayload(data string) string {
	return strings.TrimPrefix(data, "PING")
}
