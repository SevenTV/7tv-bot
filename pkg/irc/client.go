package irc

import (
	"bufio"
	"crypto/tls"
	"github.com/seventv/twitch-irc-reader/pkg/util"
	"io"
	"net"
	"net/textproto"
	"strings"
	"sync"
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

	serverDisconnect util.Closer
	clientDisconnect util.Closer

	// Connected gets closed when the connection to the IRC has been established.
	// Meaning reading from its C channel is no longer blocking.
	Connected util.Closer

	onMessage func(msg *Message, err error)
}

// New returns a new client
func New(user, oauth string) *Client {
	return &Client{
		user:   user,
		oauth:  oauth,
		UseTLS: true,
		read:   make(chan string, ReadBuffer),
		write:  make(chan []byte, WriteBuffer),
	}
}

// NewAnon returns an anonymous client, useful for testing, or small read-only bots
func NewAnon() *Client {
	return &Client{
		user:   "justinfan77777",
		oauth:  "oauth",
		UseTLS: true,
		read:   make(chan string, ReadBuffer),
		write:  make(chan []byte, WriteBuffer),
	}
}

// WithCapabilities adds twitch-irc specific capabilities to a New client, use the constants defined in capabilities.go
func (c *Client) WithCapabilities(caps ...string) *Client {
	c.capabilities = caps
	return c
}

// Connect starts the IRC connection
func (c *Client) Connect() (err error) {
	// reset all closers
	c.Connected.Reset()
	c.clientDisconnect.Reset()
	c.serverDisconnect.Reset()

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

	err = c.requestCapabilities(conn)
	if err != nil {
		return err
	}

	err = c.login(conn)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go c.startReader(conn, &wg)
	go c.startWriter(conn, &wg)

	// Send signal when the client has connected
	c.Connected.Close()

	// blocks here, until server disconnects or c.Disconnect() is called,
	//the error we get from here tells us whether the client disconnected, or the server disconnected us
	err = c.startHandler()

	// close all channels & connections to make sure there's no memory leaks
	conn.Close()
	c.serverDisconnect.Close()
	c.clientDisconnect.Close()

	// Wait until both reader & writer are closed
	wg.Wait()

	return
}

// Disconnect closes the IRC connection
func (c *Client) Disconnect() {
	c.clientDisconnect.Close()
}

// Join makes the client join the passed channels
func (c *Client) Join(channels ...string) {
	c.SendString("JOIN " + appendChannels(channels...))
}

// Part makes the client leave the passed channels
func (c *Client) Part(channels ...string) {
	c.SendString("PART " + appendChannels(channels...))
}

func (c *Client) requestCapabilities(conn net.Conn) error {
	if len(c.capabilities) == 0 {
		return nil
	}
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

func (c *Client) startReader(reader io.Reader, wg *sync.WaitGroup) {
	defer func() {
		c.serverDisconnect.Close()
		wg.Done()
	}()

	lineReader := textproto.NewReader(bufio.NewReader(reader))

	for {
		line, err := lineReader.ReadLine()
		if err != nil {
			// return will close c.serverDisconnect
			return
		}

		for _, msg := range strings.Split(line, "\r\n") {
			c.read <- msg
		}
	}
}

func (c *Client) startWriter(conn net.Conn, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()
	for {
		select {
		case <-c.clientDisconnect.C:
			return
		case <-c.serverDisconnect.C:
			return
		case msg := <-c.write:
			c.writeMessage(conn, msg)
		}
	}
}

var newLine = []byte("\r\n")

func (c *Client) writeMessage(writer io.WriteCloser, data []byte) {
	_, err := writer.Write(append(data, newLine...))
	if err != nil {
		writer.Close()
		c.serverDisconnect.Close()
	}
}

func (c *Client) startHandler() error {
	for {
		select {
		case line := <-c.read:
			c.onMessage(ParseMessage(line))
		case <-c.serverDisconnect.C:
			return ErrServerDisconnect
		case <-c.clientDisconnect.C:
			return ErrClientDisconnected
		}
	}
}

// OnMessage sets a callback to handle the raw incoming IRC messages
func (c *Client) OnMessage(cb func(msg *Message, err error)) {
	c.onMessage = cb
}

// Send a []byte message to the server (does not need \r\n at the end of the line)
func (c *Client) Send(line []byte) {
	c.write <- line
}

// SendString sends a string message to the server (does not need \r\n at the end of the line)
func (c *Client) SendString(line string) {
	c.Send([]byte(line))
}
