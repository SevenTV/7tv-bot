package irc

import "errors"

var (
	// ErrClientDisconnected is returned when the client has been disconnected
	ErrClientDisconnected = errors.New("client disconnected")
)
