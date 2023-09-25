package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/seventv/7tv-bot/pkg/types"
)

// GetChannels gets all channels, and runs the given callback in batches,
// where channels with the highest weight get the highest priority
func GetChannels(ctx context.Context, cb func([]types.Channel), batchSize int) error {
	// TODO: test if this query & indexes are optimal
	opts := options.Find().SetSort(bson.D{{"weight", -1}})
	cursor, err := collection.Find(ctx, bson.M{"flags": bson.M{"$gt": 0}}, opts)
	if err != nil {
		return err
	}

	var channels []types.Channel
	for cursor.Next(ctx) {
		channel := types.Channel{}
		err = cursor.Decode(&channel)
		if err != nil {
			continue
		}
		channels = append(channels, channel)

		if len(channels) >= batchSize {
			cb(channels)
			channels = []types.Channel{}
		}
	}
	if len(channels) > 0 {
		cb(channels)
	}
	return nil
}

func GetChannel(ctx context.Context, filter bson.D) (*types.Channel, error) {
	channel := &types.Channel{}
	err := collection.FindOne(ctx, filter).Decode(channel)
	if err != nil {
		return nil, err
	}
	return channel, nil
}

// UpsertChannel inserts a channel, or updates it if it already exists
func UpsertChannel(ctx context.Context, channel types.Channel) error {
	filter := bson.D{{"user_id", channel.ID}}
	update := bson.D{{"$set", channel}}
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}
