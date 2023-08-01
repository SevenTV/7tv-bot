package manager

type channel struct {
	name string
	// weight can be increased for busy channels, so they'll count towards more capacity being taken from the connection
	// this is important to keep track of for when we PART this channel
	weight   int
	isJoined bool
}

// newChannel returns a new channel with given name & weight.
// Weight determines how much capacity the channel takes on the connection.
// By default, any weight value of 50 or higher, will create a connection just for this channel alone.
func newChannel(name string, weight int) *channel {
	return &channel{
		name:   name,
		weight: weight,
	}
}
