package irc

import (
	"strings"
)

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
	// Join means the IRC added the given channel to the current active subscriptions
	Join
	// Part means the bot has been removed from the given channel, could be a timeout/ban, or could be because we asked the IRC to disconnect us
	Part
	// PrivMessage is a chat message sent in a Twitch channel
	PrivMessage
	// Cap is a message pertaining to the twitch-specific capabilities requested on connect
	Cap
	// Notice is a message about a command issues by the client, indicating whether it succeeded or failed
	Notice
)

type Message struct {
	raw         string
	messageType MessageType
}

// ParseMessage returns a new message pointer containing the raw data & the message type.
// returns an error if something went wrong, but will still contain the message object, so you can access the raw data.
func ParseMessage(data string) (*Message, error) {
	m := &Message{raw: data}

	var err error
	m.messageType, err = parseMessageType(data)

	return m, err
}

// String returns the raw IRC message as a string
func (m *Message) String() string {
	return m.raw
}

// GetType returns message type, if it has not yet been set, it parses the message to find its type and sets m.messageType for future reference
func (m *Message) GetType() MessageType {
	// return stored type if it has already been parsed
	if m.messageType != Unset {
		return m.messageType
	}

	// parse type if not set yet, with normal use this shouldn't happen, but you never know
	m.messageType, _ = parseMessageType(m.String())

	return m.messageType
}

func parseMessageType(rawMessage string) (MessageType, error) {
	var i int
	split := strings.Split(rawMessage, " ")

	if strings.HasPrefix(rawMessage, "@") {
		i++
	}

	if i >= len(split) {
		return Unknown, ErrPartialMessage
	}

	if strings.HasPrefix(split[i], ":") {
		i++
	}

	if i >= len(split) {
		return Unknown, ErrPartialMessage
	}

	return parseType(split[i]), nil
}

func parseType(str string) MessageType {
	switch str {
	case "PING":
		return Ping
	case "RECONNECT":
		return Reconnect
	case "JOIN":
		return Join
	case "PART":
		return Part
	case "PRIVMSG":
		return PrivMessage
	case "CAP":
		return Cap
	case "NOTICE":
		return Notice
	default:
		return Unknown
	}
}
