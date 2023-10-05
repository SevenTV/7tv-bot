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
	Name      string                     `bson:"name"`
	EmoteID   primitive.ObjectID         `bson:"emote_id"`
	Flags     model.ActiveEmoteFlagModel `bson:"flags"`
	State     []model.EmoteVersionState  `bson:"state,omitempty"`
	URL       string                     `bson:"url"`
	CreatedAt time.Time                  `bson:"created_at"`
	UpdatedAt time.Time                  `bson:"updated_at"`
	Count     int                        `bson:"count"`
}
