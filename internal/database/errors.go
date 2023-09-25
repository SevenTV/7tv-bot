package database

import "errors"

var (
	// ErrChannelAlreadyExists is returned when trying to add a channel that already exists in the database
	ErrChannelAlreadyExists = errors.New("channel already exists")
	// ErrChannelNotFound is returned when the channel you're trying to perform an operation on doesn't exist
	ErrChannelNotFound = errors.New("channel not found")
)
