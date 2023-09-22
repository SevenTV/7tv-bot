package irc_reader

import (
	"strings"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/seventv/twitch-irc-reader/pkg/bitwise"
	"github.com/seventv/twitch-irc-reader/pkg/irc"
	"github.com/seventv/twitch-irc-reader/pkg/types"
)

func (c *Controller) onMessage(msg *irc.Message, err error) {
	// skip anything that's not a channel message
	if msg.GetType() != irc.PrivMessage {
		return
	}

	// publish message to nats topic
	subject := c.cfg.Nats.Topic.Raw
	subject += ".privmsg." + parseChannel(msg.String())

	// set message ID as header, so we can filter out duplicate messages with JetStream
	header := nats.Header{}
	header.Add("Nats-Msg-Id", parseMessageId(msg.String()))

	_, err = c.jetStream.PublishMsg(&nats.Msg{
		Subject: subject,
		Header:  header,
		Data:    []byte(msg.String()),
	})

	if err != nil {
		zap.L().Error(
			"failed publish to NATS",
			zap.String("error", err.Error()),
		)
	}
}

func (c *Controller) handleOrphanedChannels() {
	for channel := range c.twitch.OrphanedChannels {
		c.joinSem <- struct{}{}
		ch := channel
		go func() {
			err := c.twitch.Join(ch.Name, ch.Weight)
			if err != nil {
				zap.L().Error(
					"failed to rejoin orphaned channel",
					zap.String("error", err.Error()),
					zap.String("channel", ch.Name),
				)
			}
			<-c.joinSem
		}()
	}
}

func (c *Controller) joinChannels(channels []types.Channel) {
	for _, channel := range channels {
		c.joinSem <- struct{}{}
		ch := channel
		go func() {
			// TODO: filter out channels based on user ID & shard ID, so we can spread the load across kubernetes statefulset

			// make sure the channel is flagged to be joined
			if !bitwise.Has(ch.Flags, bitwise.JOIN_IRC) {
				return
			}

			err := c.twitch.Join(ch.Username, ch.Weight)
			if err != nil {
				zap.L().Error(
					"failed to join channel",
					zap.String("error", err.Error()),
					zap.String("channel", ch.Username),
				)
			}
			<-c.joinSem
		}()
	}
}

// parses out the channel name from a PRIVMSG,
// don't use on any other type of message seen as though there's no slice length checks
func parseChannel(in string) string {
	return strings.TrimPrefix(strings.Split(strings.Split(in, "PRIVMSG")[1], " ")[1], "#")
}

// parses out the message ID from a PRIVMSG with optional tags
// don't use on any other type of message seen as though there's no slice length checks
func parseMessageId(in string) string {
	return strings.Split(strings.Split(in, ";id=")[1], ";")[0]
}
