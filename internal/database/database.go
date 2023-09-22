package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	db         *mongo.Database
	collection *mongo.Collection
)

func Connect(uri, database string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	opt := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		return err
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return err
	}
	db = client.Database(database)
	return nil
}

func Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return db.Client().Ping(ctx, readpref.Primary())
}

func EnsureCollection(coll string) {
	ctx := context.Background()
	db.CreateCollection(ctx, coll)

	collection = db.Collection(coll)
	indexes := []mongo.IndexModel{
		// TODO: find optimal indexes
		{Keys: bson.D{{"user_id", -1}}},
		{Keys: bson.D{{"flags", -1}}},
		{Keys: bson.D{{"weight", -1}}},
	}

	collection.Indexes().CreateMany(ctx, indexes)
}

// TODO: watch for changes on the collection, so we can join/leave channels when they get added or removed
