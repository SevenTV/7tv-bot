package types

import "github.com/seventv/api/data/model"

type CountedEmote struct {
	Count int
	Emote model.ActiveEmoteModel
}
