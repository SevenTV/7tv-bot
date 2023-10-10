package emotes

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/seventv/7tv-bot/pkg/types"
)

func IncrementEmote(ctx context.Context, emote types.CountedEmote) error {
	res, err := collections.GlobalStats.UpdateOne(
		ctx,
		bson.D{{"emote_id", emote.Emote.EmoteID}},
		bson.M{"$setOnInsert": EmoteCount{
			Name:      emote.Emote.Name,
			EmoteID:   emote.Emote.EmoteID,
			Flags:     emote.Emote.Flags,
			State:     emote.Emote.State,
			URL:       emote.Emote.URL,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Count:     emote.Count,
		}},
		options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	if res.UpsertedCount > 0 {
		return nil
	}

	_, err = collections.GlobalStats.UpdateOne(
		ctx,
		bson.D{{"emote_id", emote.Emote.EmoteID}},
		bson.M{
			"$inc": bson.M{
				"count": emote.Count,
			},
			"$set": bson.M{
				"updated_at": time.Now().UTC(),
			},
		})

	return err
}
