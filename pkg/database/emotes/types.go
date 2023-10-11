package emotes

import (
	"errors"
	"time"

	"github.com/seventv/api/data/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrMissingData = errors.New("missing data")
)

type EmoteCount struct {
	Name      string                     `bson:"name" json:"name"`
	EmoteID   primitive.ObjectID         `bson:"emote_id" json:"emote_id"`
	Flags     model.ActiveEmoteFlagModel `bson:"flags,omitempty" json:"flags,omitempty"`
	State     []model.EmoteVersionState  `bson:"state,omitempty" json:"state,omitempty"`
	URL       string                     `bson:"url" json:"url"`
	CreatedAt time.Time                  `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time                  `bson:"updated_at" json:"updated_at"`
	Count     int                        `bson:"count" json:"count"`
}
