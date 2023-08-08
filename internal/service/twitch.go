package service

import (
	"go.uber.org/zap"

	"github.com/seventv/twitch-irc-reader/pkg/irc"
)

func (c *Controller) onMessage(msg *irc.Message, err error) {
	// TODO: implement
}

func (c *Controller) handleOrphanedChannels() {
	for ch := range c.twitch.OrphanedChannels {
		// TODO: apply rate limit
		err := c.twitch.Join(ch.Name, ch.Weight)
		if err != nil {
			zap.L().Error(
				"failed to rejoin orphaned channel",
				zap.String("error", err.Error()),
				zap.String("channel", ch.Name),
			)
		}
	}
}
