package emotes

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/seventv/7tv-bot/pkg/types"
)

func IncrementEmote(ctx context.Context, emote types.EmoteCount) error {
	now := time.Now().UTC()
	emote.CreatedAt = &now
	emote.UpdatedAt = &now
	res, err := collections.GlobalStats.UpdateOne(
		ctx,
		bson.D{{"emote_id", emote.EmoteID}},
		bson.M{"$setOnInsert": emote},
		options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	if res.UpsertedCount > 0 {
		return nil
	}

	_, err = collections.GlobalStats.UpdateOne(
		ctx,
		bson.D{{"emote_id", emote.EmoteID}},
		bson.M{
			"$inc": bson.M{
				"count": emote.Count,
			},
			"$set": bson.M{
				"updated_at": now,
			},
		})

	return err
}

func GetTopEmotes(ctx context.Context, limit int64, page int64) ([]types.EmoteCount, error) {
	skip := limit * (page - 1)
	cursor, err := collections.GlobalStats.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{"count", -1}}).SetLimit(limit).SetSkip(skip))
	if err != nil {
		return nil, err
	}

	var emotes []types.EmoteCount
	if err := cursor.All(ctx, &emotes); err != nil {
		return nil, err
	}

	return emotes, nil
}
