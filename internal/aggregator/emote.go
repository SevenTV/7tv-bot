package aggregator

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/seventv/api/data/model"
)

type countedEmote struct {
	count int
	emote model.ActiveEmoteModel
}

func countEmotes(msg *Message) ([]countedEmote, error) {
	// TODO: global emotes & personal emotes
	emotes, err := getEmotesForChannel(msg.Room.ID)
	if err != nil {
		return nil, err
	}

	var result []countedEmote

	// return if we don't have emotes for this channel
	if len(emotes) == 0 {
		return result, nil
	}

	for _, emote := range emotes {
		counted := countedEmote{emote: emote}
		for _, word := range msg.MessageWords {
			if word != emote.Name {
				continue
			}
			counted.count++
		}

		if counted.count == 0 {
			continue
		}

		result = append(result, counted)
	}

	return result, nil
}

// TODO: invalidate cache when event API sends an update
var activeEmotesCache = make(map[string]emoteCache)
var mx = sync.Mutex{}

type emoteCache struct {
	expires time.Time
	emotes  []model.ActiveEmoteModel
}

func getEmotesForChannel(channelID string) ([]model.ActiveEmoteModel, error) {
	mx.Lock()
	defer mx.Unlock()

	cache, ok := activeEmotesCache[channelID]
	if ok {
		if time.Since(cache.expires) <= 0 {
			return cache.emotes, nil
		}
	}
	emotes, err := getEmotesByChannelId(channelID)
	if err != nil {
		if errors.Is(err, ErrEmotesNotEnabled) {
			// set empty slice, so we don't spam the API with requests in the future
			activeEmotesCache[channelID] = emoteCache{
				emotes:  []model.ActiveEmoteModel{},
				expires: time.Now().Add(5 * time.Minute),
			}
			return []model.ActiveEmoteModel{}, nil
		}
		return nil, err
	}

	activeEmotesCache[channelID] = emoteCache{
		emotes:  emotes,
		expires: time.Now().Add(5 * time.Minute),
	}
	return emotes, nil
}

func getEmotesByChannelId(channelID string) ([]model.ActiveEmoteModel, error) {
	req, err := http.NewRequest("GET", "https://7tv.io/v3/users/twitch/"+channelID, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return nil, ErrEmotesNotEnabled
		}
		return nil, ErrUnexpectedStatus
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	userModel := &model.UserConnectionModel{}
	err = json.Unmarshal(body, userModel)
	if err != nil {
		return nil, err
	}

	if userModel.EmoteSet == nil {
		return nil, ErrIncompleteResponse
	}

	return userModel.EmoteSet.Emotes, nil
}
