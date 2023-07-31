package irc

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"net/textproto"
	"strings"
	"time"
)

var (
	// ReadBuffer assigns a buffer size to the read channel
	ReadBuffer = 32
	// WriteBuffer assigns a buffer size to the write channel
	WriteBuffer = 0
	// Address sets the address that the client will connect to
	Address = "irc.chat.twitch.tv:6667"
	// AddressTLS sets the address that the client will connect to in TLS mode
	AddressTLS = "irc.chat.twitch.tv:6697"
)

// Client handles the IRC connection and incoming & outgoing messages.
// The client requires you to respond to PING messages manually
// as well as keep track of which channels you're connected to using the incoming JOIN & PART messages
type Client struct {
	user  string
	oauth string

	capabilities []string

	// UseTLS determines whether the IRC connects with or without TLS, needs to be set before you call Connect, default = true
	UseTLS bool

	read  chan string
	write chan []byte

	// TODO: convert to closer
	clientDisconnect chan struct{}

	onConnect func()
	onMessage func(msg *Message, err error)
}

// New returns a new client
func New(user, oauth string) *Client {
	return &Client{
		user:             user,
		oauth:            oauth,
		UseTLS:           true,
		read:             make(chan string, ReadBuffer),
		write:            make(chan []byte, WriteBuffer),
		clientDisconnect: make(chan struct{}, 0),
	}
}

// NewAnon returns an anonymous client, useful for testing, or small read-only bots
func NewAnon() *Client {
	return &Client{
		user:             "justinfan4321",
		oauth:            "oauth",
		UseTLS:           true,
		read:             make(chan string, ReadBuffer),
		write:            make(chan []byte, WriteBuffer),
		clientDisconnect: make(chan struct{}, 0),
	}
}

// WithCapabilities adds twitch-irc specific capabilities to a New client, use the constants defined in capabilities.go
func (c *Client) WithCapabilities(caps ...string) *Client {
	c.capabilities = caps
	return c
}

// Connect starts the IRC connection
func (c *Client) Connect() (err error) {
	dialer := &net.Dialer{KeepAlive: time.Second * 10}
	var conn net.Conn
	if c.UseTLS {
		conn, err = tls.DialWithDialer(dialer, "tcp", AddressTLS, &tls.Config{MinVersion: tls.VersionTLS12})
	} else {
		conn, err = dialer.Dial("tcp", Address)
	}

	if err != nil {
		return err
	}

	if len(c.capabilities) > 0 {
		err = c.requestCapabilities(conn)
		if err != nil {
			return err
		}
	}

	err = c.login(conn)
	if err != nil {
		return err
	}

	go c.startReader(conn)
	go c.startWriter(conn)

	return c.startHandler()
}

// Join makes the client join the passed channels
func (c *Client) Join(channels ...string) {
	c.sendString("JOIN " + appendChannels(channels...))
}

// Depart makes the client leave the passed channels
func (c *Client) Depart(channels ...string) {
	c.sendString("PART " + appendChannels(channels...))
}

func (c *Client) requestCapabilities(conn net.Conn) error {
	_, err := conn.Write([]byte("CAP REQ :" + strings.Join(c.capabilities, " ") + "\r\n"))
	return err
}

func (c *Client) login(conn net.Conn) error {
	_, err := conn.Write([]byte("PASS " + c.oauth + "\r\n"))
	if err != nil {
		return err
	}
	_, err = conn.Write([]byte("NICK " + c.user + "\r\n"))
	return err
}

func (c *Client) startReader(reader io.Reader) {
	defer func() {
		c.clientDisconnect <- struct{}{}
	}()

	lineReader := textproto.NewReader(bufio.NewReader(reader))

	for {
		line, err := lineReader.ReadLine()
		if err != nil {
			// return will send a signal through c.clientDisconnect
			return
		}

		for _, msg := range strings.Split(line, "\r\n") {
			c.read <- msg
		}
	}
}

func (c *Client) startWriter(conn net.Conn) {
	// TODO: implement the (as of yet non-existent) connection closer
	for {
		select {
		case msg := <-c.write:
			c.writeMessage(conn, msg)
		}
	}
}

var newLine = []byte("\r\n")

func (c *Client) writeMessage(writer io.WriteCloser, data []byte) {
	_, err := writer.Write(append(data, newLine...))
	if err != nil {
		// closes underlying connection, making the reader send disconnect signal (not ideal, will be fixed)
		writer.Close()
	}
}

func (c *Client) startHandler() error {
	for {
		select {
		case line := <-c.read:
			c.onMessage(parseMessage(line))
		case <-c.clientDisconnect:
			return ErrClientDisconnected
		}
	}
}

// OnConnect sets a callback that runs when the client has established a connection
func (c *Client) OnConnect(cb func()) {
	c.onConnect = cb
}

// OnMessage sets a callback to handle the raw incoming IRC messages
func (c *Client) OnMessage(cb func(msg *Message, err error)) {
	c.onMessage = cb
}

func (c *Client) send(line []byte) {
	c.write <- line
}

func (c *Client) sendString(line string) {
	c.send([]byte(line))
}
