package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
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
