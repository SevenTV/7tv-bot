package util

import "github.com/seventv/7tv-bot/pkg/types"

// VerifyChannel returns true if a channel has all required fields
func VerifyChannel(channel types.Channel) bool {
	if channel.ID == 0 {
		return false
	}
	if channel.Username == "" {
		return false
	}
	if channel.Weight == 0 {
		return false
	}
	if channel.Platform == "" {
		return false
	}
	return true
}
