package database

import "errors"

var (
	// ErrChannelAlreadyExists is returned when trying to add a channel that already exists in the database
	ErrChannelAlreadyExists = errors.New("channel already exists")
)
