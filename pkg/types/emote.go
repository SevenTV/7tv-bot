package types

import (
	"github.com/seventv/api/data/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CountedEmote struct {
	Count int
	Emote Emote
}

type Emote struct {
	Name    string                     `bson:"name"`
	EmoteID primitive.ObjectID         `bson:"emote_id"`
	Flags   model.ActiveEmoteFlagModel `bson:"flags"`
	State   []model.EmoteVersionState  `bson:"state,omitempty"`
	URL     string                     `bson:"url"`
}
