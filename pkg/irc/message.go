package irc

// MessageType is a type with the underlying type int, it is used with enum constants defined below to help classify incoming IRC messages
type MessageType int

const (
	// Unknown means the message type is currently not supported
	Unknown MessageType = iota - 1
	// Unset means the message type has not yet been parsed
	Unset
	// Ping is a PING message coming from the IRC and must be replied to with a PONG message
	Ping
	// Reconnect when this is sent by the IRC, the server is about to disconnect the client, usually means there's a server update
	Reconnect
	// Part can mean the bot got banned from a channel, usually means a user disconnected
	Part
	// PrivMessage is a chat message sent in a Twitch channel
	PrivMessage
)

// Message contains some methods to help with the correct handling of incoming IRC messages
type Message interface {
	// GetType returns the message type, as defined with the MessageType type and its constants
	GetType() MessageType
	// Raw returns the raw IRC message as a slice of bytes
	Raw() []byte
	// String returns the raw IRC message as a string
	String() string
}

type message struct {
	raw         []byte
	messageType MessageType
}

// GetType returns message type, if it has not yet been set, it parses the message to find its type and sets m.messageType for future reference
func (m message) GetType() MessageType {
	// return stored type if it has already been parsed
	if m.messageType != Unset {
		return m.messageType
	}

	// TODO: high priority! implement, needed for PING

	return m.messageType
}

func (m message) Raw() []byte {
	return m.raw
}

func (m message) String() string {
	return string(m.raw)
}
