package usage_api

import (
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/seventv/7tv-bot/pkg/pubsub"
	"github.com/seventv/7tv-bot/pkg/types"
)

// Stream is a stream of emotes, read from NATS, and published to the websocket in batches
type Stream struct {
	mx     sync.Mutex
	emotes []types.EmoteCount
}

func newStream() *Stream {
	return &Stream{
		emotes: []types.EmoteCount{},
	}
}

// Init starts the worker that publishes emotes to the websocket in batches every 500ms
func (s *Stream) Init() {
	for range time.Tick(500 * time.Millisecond) {
		s.mx.Lock()
		emotes := s.emotes
		s.emotes = []types.EmoteCount{}
		s.mx.Unlock()

		data, err := json.Marshal(emotes)
		if err != nil {
			zap.S().Errorw("failed to marshal emotes", "error", err)
			continue
		}
		pubsub.Publish(data)
	}
}

// AddEmote adds an emote to the current batch of emotes
func (s *Stream) AddEmote(addedEmote types.EmoteCount) {
	s.mx.Lock()
	defer s.mx.Unlock()
	for i, emote := range s.emotes {
		if emote.EmoteID == addedEmote.EmoteID {
			s.emotes[i].Count += addedEmote.Count
			return
		}
	}

	s.emotes = append(s.emotes, addedEmote)
}
