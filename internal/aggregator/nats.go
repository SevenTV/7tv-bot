package aggregator

import (
	"context"
	"errors"
	"time"

	"github.com/nats-io/nats.go"
)

func (s *Service) subscribe(ctx context.Context, cb func(msg *nats.Msg) error) error {
	js, err := s.newJetStream()
	if err != nil {
		return err
	}
	sub, err := js.QueueSubscribeSync(s.cfg.Nats.Topic.Raw, s.cfg.Nats.Consumer)
	if err != nil {
		panic(err)
	}

	for {
		msg, err := sub.NextMsgWithContext(ctx)
		if err != nil {
			panic(err)
		}
		_, err = msg.Metadata()
		if err != nil {
			panic(err)
		}

		if err := msg.InProgress(); err != nil {
			_ = msg.Nak()
			continue
		}

		err = cb(msg)
		if err != nil {
			// If we cannot process the message, send Nak, so another consumer can try again.
			// we don't need to explicitly do this, but it does speed things up
			_ = msg.Nak()
			continue
		}
		_ = msg.Ack()
	}
	return nil
}

func (s *Service) newJetStream() (nats.JetStreamContext, error) {
	nc, err := nats.Connect(s.cfg.Nats.URL)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}
	err = s.ensureStream(js)
	if err != nil {
		return nil, err
	}
	err = s.ensureConsumer(js)
	return js, err
}

func (s *Service) ensureConsumer(js nats.JetStreamContext) error {
	consumercfg := &nats.ConsumerConfig{
		Name:           s.cfg.Nats.Consumer,
		Durable:        s.cfg.Nats.Consumer,
		DeliverGroup:   s.cfg.Nats.Consumer,
		MaxDeliver:     3,
		AckWait:        1 * time.Minute,
		AckPolicy:      nats.AckExplicitPolicy,
		DeliverPolicy:  nats.DeliverAllPolicy,
		FilterSubject:  s.cfg.Nats.Topic.Raw,
		DeliverSubject: nats.NewInbox(),
	}
	_, err := js.ConsumerInfo(s.cfg.Nats.Stream, consumercfg.Name)
	if errors.Is(err, nats.ErrConsumerNotFound) {
		_, err = js.AddConsumer(s.cfg.Nats.Stream, consumercfg)
		return err
	}
	_, err = js.UpdateConsumer(s.cfg.Nats.Stream, consumercfg)
	return err
}

// Ensure we have a JetStream set up for deduplication on the producer.
// Keep in mind, each message published to this stream needs to have the Nats-Msg-Id header set for deduplication.
func (s *Service) ensureStream(js nats.JetStreamContext) error {
	cfg := &nats.StreamConfig{
		Name:      s.cfg.Nats.Stream,
		Subjects:  []string{s.cfg.Nats.Topic.Raw},
		Retention: nats.InterestPolicy,
		Discard:   nats.DiscardNew,
		// TODO: 0 seconds sets this to default value (2 min), find optimal value for our case
		Duplicates: 0 * time.Second,
	}

	_, err := js.StreamInfo(cfg.Name)
	if errors.Is(err, nats.ErrStreamNotFound) {
		_, err = js.AddStream(cfg)
		return err
	}
	_, err = js.UpdateStream(cfg)

	return err
}
