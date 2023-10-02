package aggregator

import "errors"

var (
	ErrIncompleteResponse = errors.New("incomplete response")
	ErrUnexpectedStatus   = errors.New("unexpected status code")
	ErrEmotesNotEnabled   = errors.New("user does not have an active emote set")
)
