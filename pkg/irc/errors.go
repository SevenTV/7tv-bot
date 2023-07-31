package irc

import "errors"

var (
	// ErrClientDisconnected is returned when the client has disconnected
	ErrClientDisconnected = errors.New("client disconnected")
	// ErrServerDisconnect is returned when the server disconnected us
	ErrServerDisconnect = errors.New("server disconnectd")
	// ErrPartialMessage is returned when the message doesn't contain all expected data
	ErrPartialMessage = errors.New("partial message")
)
