package service

import (
	"go.uber.org/zap"

	"github.com/seventv/twitch-irc-reader/pkg/irc"
)

func (c *Controller) onMessage(msg *irc.Message, err error) {
	// publish message to nats topic,
	// TODO: maybe do some filtering so only PRIVMSG goes through? We definitely don't want whispers to get published, cause they'll be replicated for every connection. And/or add different subject suffix for twitch channel?
	err = c.queue.Publish(c.cfg.Nats.Topic.Raw, []byte(msg.String()))
	if err != nil {
		zap.L().Error(
			"failed publish to NATS",
			zap.String("error", err.Error()),
		)
	}
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
