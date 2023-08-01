package manager

import "errors"

var (
	// ErrConnNotFound means the connection you're trying to perform an operation on does not exist
	ErrConnNotFound = errors.New("connection not found")
	// ErrChanAlreadyJoined means the channel you're trying to join is already active
	ErrChanAlreadyJoined = errors.New("channel already joined")
	// ErrChanNotFound means the channel you're trying to perform an operation on wasn't found on the current connection
	ErrChanNotFound = errors.New("channel not found")
)
