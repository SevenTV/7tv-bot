package aggregator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

func (s *Service) subscribe(ctx context.Context, cb func(msg *nats.Msg) error) error {
	js, err := s.newJetStream()
	if err != nil {
		return err
	}

	sub, err := js.PullSubscribe(s.cfg.Nats.Topic.Raw+".>", s.cfg.Nats.Consumer)
	if err != nil {
		panic(err)
	}

	// limit the amount of workers for the callback
	sem := make(chan struct{}, s.cfg.Maxworkers)

	for {
		msg, err := s.fetchOne(ctx, sub)
		if err != nil {
			if err.Error() == "fetch: context deadline exceeded" {
				continue
			}
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
		sem <- struct{}{}
		go func() {
			err := cb(msg)
			if err != nil {
				zap.S().Error("couldn't process message from NATS", err)
				// If we cannot process the message, send Nak, so another consumer can try again.
				// we don't need to explicitly do this, but it does speed things up
				_ = msg.Nak()
				<-sem
				return
			}
			_ = msg.Ack()
			<-sem
		}()
	}
	return nil
}

func (s *Service) fetchOne(ctx context.Context, sub *nats.Subscription) (*nats.Msg, error) {
	msgs, err := sub.Fetch(1, nats.Context(ctx))
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	if len(msgs) == 0 {
		return nil, errors.New("no messages")
	}

	return msgs[0], nil
}

func (s *Service) newJetStream() (nats.JetStreamContext, error) {
	nc, err := nats.Connect(s.cfg.Nats.URL)
	if err != nil {
		return nil, err
	}
	s.nc = nc
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
		Durable:       s.cfg.Nats.Consumer,
		DeliverGroup:  s.cfg.Nats.Consumer,
		MaxDeliver:    3,
		AckWait:       1 * time.Minute,
		AckPolicy:     nats.AckExplicitPolicy,
		DeliverPolicy: nats.DeliverAllPolicy,
		FilterSubject: s.cfg.Nats.Topic.Raw + ".>",
	}
	_, err := js.ConsumerInfo(s.cfg.Nats.Stream, consumercfg.Durable)
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
		Subjects:  []string{s.cfg.Nats.Topic.Raw + ".>"},
		MaxAge:    1 * time.Hour,
		Retention: nats.InterestPolicy,
		Discard:   nats.DiscardNew,
		// TODO: 0 seconds sets this to default value (2 min), find optimal value for our case
		Duplicates: 1 * time.Minute,
	}

	_, err := js.StreamInfo(cfg.Name)
	if errors.Is(err, nats.ErrStreamNotFound) {
		_, err = js.AddStream(cfg)
		return err
	}
	_, err = js.UpdateStream(cfg)

	return err
}
