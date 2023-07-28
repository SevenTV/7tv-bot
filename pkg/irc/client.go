package irc

import (
	"bufio"
	"io"
	"net"
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

// Client handles the IRC connection and incoming messages
type Client struct {
	user  string
	oauth string

	UseTLS bool

	read  chan []byte
	write chan []byte

	// TODO: convert to closer
	clientDisconnect chan struct{}

	onConnect func()
	onMessage func(msg Message)
}

// New returns a new client
func New(user, oauth string) *Client {
	return &Client{
		user:             user,
		oauth:            oauth,
		read:             make(chan []byte, ReadBuffer),
		write:            make(chan []byte, WriteBuffer),
		clientDisconnect: make(chan struct{}, 0),
	}
}

// NewAnon returns an anonymous client, useful for testing, or small read-only bots
func NewAnon() *Client {
	return &Client{
		user:             "justinfan4321",
		oauth:            "oauth",
		read:             make(chan []byte, ReadBuffer),
		write:            make(chan []byte, WriteBuffer),
		clientDisconnect: make(chan struct{}, 0),
	}
}

// Connect starts the IRC connection
func (c *Client) Connect() error {
	// TODO: TLS support
	dialer := &net.Dialer{KeepAlive: time.Second * 10}

	conn, err := dialer.Dial("tcp", Address)
	if err != nil {
		return err
	}

	c.login(conn)

	go c.startReader(conn)
	go c.startWriter(conn)

	return c.startHandler()
}

// Join makes the client join the passed channels
func (c *Client) Join(channels ...string) {
	joinString := "JOIN "
	for i, channel := range channels {
		joinString += "#" + channel

		if i < len(channels)-1 {
			joinString += ","
		}
	}

	c.sendString(joinString)
}

// Depart makes the client leave the passed channels
func (c *Client) Depart(channels ...string) {
	// TODO: implement
}

func (c *Client) login(conn net.Conn) {
	conn.Write([]byte("PASS " + c.oauth + "\r\n"))
	conn.Write([]byte("NICK " + c.user + "\r\n"))
}

func (c *Client) startReader(reader io.Reader) {
	defer func() {
		c.clientDisconnect <- struct{}{}
	}()

	lineReader := bufio.NewReader(reader)

	for {
		// TODO: check to make sure there's no line breaks in the incoming data
		line, _, err := lineReader.ReadLine()
		if err != nil {
			// return will send a signal through c.clientDisconnect
			return
		}
		c.read <- line
	}
}

func (c *Client) startWriter(conn net.Conn) {
	// TODO: implement the (as of yet non existent) connection closer
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
			c.onMessage(message{raw: line})
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
func (c *Client) OnMessage(cb func(msg Message)) {
	c.onMessage = cb
}

func (c *Client) send(line []byte) {
	c.write <- line
}

func (c *Client) sendString(line string) {
	c.send([]byte(line))
}
