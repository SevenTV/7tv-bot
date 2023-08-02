package manager

import (
	"strings"
)

type ircChannel struct {
	name string
	// weight can be increased for busy channels, so they'll count towards more capacity being taken from the connection
	// this is important to keep track of for when we PART this channel
	weight   int
	isJoined bool

	// connectionKey contains the key of the assigned connection
	connectionKey uint
}

// newIrcChannel returns a new ircChannel with given name & weight.
// Weight determines how much capacity the channel takes on the connection.
// By default, any weight value of 50 or higher, will create a connection just for this channel alone.
func newIrcChannel(name string, weight int) *ircChannel {
	if weight > ConnectionCapacity {
		weight = ConnectionCapacity
	}
	return &ircChannel{
		name:   strings.ToLower(name),
		weight: weight,
	}
}
