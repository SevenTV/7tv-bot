package usage_api

import (
	"github.com/nats-io/nats.go"

	"github.com/seventv/7tv-bot/pkg/pubsub"
)

func (s *Server) natsSubscribe() {
	s.nc.Subscribe(s.cfg.Nats.Topic.Emotes, func(m *nats.Msg) {
		pubsub.Publish(m.Data)
	})
}
