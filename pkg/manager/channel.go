package manager

import (
	"strings"
)

type IRCChannel struct {
	Name string
	// weight can be increased for busy channels, so they'll count towards more capacity being taken from the connection
	// this is important to keep track of for when we PART this channel
	Weight   int
	isJoined bool

	// connectionKey contains the key of the assigned connection
	connectionKey uint
}

// NewIrcChannel returns a new IRCChannel with given name & weight.
// Weight determines how much capacity the channel takes on the connection.
// By default, any weight value of 50 or higher, will create a connection just for this channel alone.
func NewIrcChannel(name string, weight int) *IRCChannel {
	if weight > ConnectionCapacity {
		weight = ConnectionCapacity
	}
	if weight <= 0 {
		weight = 1
	}
	return &IRCChannel{
		Name:   strings.ToLower(name),
		Weight: weight,
	}
}
