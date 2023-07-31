package irc

// appendChannels is a helper function for the Join & Part methods
func appendChannels(channels ...string) string {
	var result string
	for i, channel := range channels {
		result += "#" + channel

		if i < len(channels)-1 {
			result += ","
		}
	}
	return result
}
