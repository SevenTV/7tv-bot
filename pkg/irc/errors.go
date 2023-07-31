package irc

import "errors"

var (
	// ErrClientDisconnected is returned when the client has been disconnected
	ErrClientDisconnected = errors.New("client disconnected")
	// ErrPartialMessage is returned when the message doesn't contain all expected data
	ErrPartialMessage = errors.New("partial message")
)
