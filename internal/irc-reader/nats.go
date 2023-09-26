package irc_reader

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/seventv/7tv-bot/internal/database"
	"github.com/seventv/7tv-bot/pkg/bitwise"
	"github.com/seventv/7tv-bot/pkg/types"
)

// Ensure we have a JetStream set up for deduplication on the producer.
// Keep in mind, each message published to this stream needs to have the Nats-Msg-Id header set for deduplication.
func (c *Controller) ensureStream(js nats.JetStreamContext) error {
	cfg := &nats.StreamConfig{
		Name:      c.cfg.Nats.Stream,
		Subjects:  []string{fmt.Sprintf("%v.>", c.cfg.Nats.Topic.Raw)},
		Retention: nats.InterestPolicy,
		Discard:   nats.DiscardNew,
		// TODO: 0 seconds sets this to default value (2 min), find optimal value for our case
		Duplicates: 0 * time.Second,
	}

	_, err := js.StreamInfo(cfg.Name)
	if errors.Is(err, nats.ErrStreamNotFound) {
		_, err = js.AddStream(cfg)
		return err
		//created = true
	}
	_, err = js.UpdateStream(cfg)

	return err
}

// Watch NATS subject for database changes sent from the API,
// we don't use jetstream here because we don't need this buffered
func (c *Controller) watchChanges(nc *nats.Conn) {
	nc.Subscribe(c.cfg.Nats.Topic.Api, func(msg *nats.Msg) {
		channel := types.Channel{}
		err := json.Unmarshal(msg.Data, &channel)
		if err != nil {
			zap.S().Error("failed to unmarshal insert", err)
			return
		}

		op := msg.Header.Get("OP")

		switch op {
		case database.Insert:
			// if the channel is not flagged to join, we skip past it
			if !bitwise.Has(channel.Flags, bitwise.JOIN_IRC) {
				return
			}
			println("joining: " + channel.Username)
			c.joinChannel(channel)
		case database.Update:
			// TODO: implement
			return
		case database.Delete:
			println("parting: " + channel.Username)
			c.twitch.Part(channel.Username)
		default:
			zap.S().Error("unknown database operation", op)
		}
	})
}
