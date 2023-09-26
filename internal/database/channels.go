package database

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

// GetChannel finds a single channel based on the BSON filter provided
func GetChannel(ctx context.Context, filter bson.D) (*types.Channel, error) {
	channel := &types.Channel{}
	err := collection.FindOne(ctx, filter).Decode(channel)
	return channel, err
}

// InsertChannel inserts a channel, returns error if it already exists
func InsertChannel(ctx context.Context, channel types.Channel) error {
	err := collection.FindOne(ctx, bson.D{{"user_id", channel.ID}}).Err()
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return ErrChannelAlreadyExists
	}

	if channel.CreatedAt.IsZero() {
		channel.CreatedAt = time.Now()
	}
	channel.UpdatedAt = time.Now()

	_, err = collection.InsertOne(ctx, channel)
	return err
}

// UpsertChannel inserts a channel, or updates it if it already exists
func UpsertChannel(ctx context.Context, channel types.Channel) error {
	if channel.CreatedAt.IsZero() {
		channel.CreatedAt = time.Now()
	}
	channel.UpdatedAt = time.Now()
	filter := bson.D{{"user_id", channel.ID}}
	update := bson.D{{"$set", channel}}
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// DeleteChannel deletes a channel by ID, returns an error if the channel ID was not found
func DeleteChannel(ctx context.Context, id int64) error {
	res, err := collection.DeleteOne(ctx, bson.D{{"user_id", id}})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrChannelNotFound
	}
	return nil
}
