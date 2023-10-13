package usage_api

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"

	"github.com/seventv/7tv-bot/pkg/types"
)

func (s *Server) natsSubscribe() {
	s.nc.Subscribe(s.cfg.Nats.Topic.Emotes, func(m *nats.Msg) {
		emote, err := parseEmoteCount(m.Data)
		if err != nil {
			zap.S().Errorw("failed to parse emote from NATS", "error", err)
			return
		}
		s.emoteStream.AddEmote(emote)
	})
}

func parseEmoteCount(data []byte) (types.EmoteCount, error) {
	var emote types.EmoteCount
	err := json.Unmarshal(data, &emote)
	return emote, err
}
