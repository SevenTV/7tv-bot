package aggregator

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/seventv/7tv-bot/internal/aggregator/config"
	"github.com/seventv/7tv-bot/pkg/database"
	emotedb "github.com/seventv/7tv-bot/pkg/database/emotes"
)

type Service struct {
	cfg *config.Config
	nc  *nats.Conn
}

func New(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) Init() error {
	err := database.Connect(
		s.cfg.Mongo.ConnectionString,
		s.cfg.Mongo.Username,
		s.cfg.Mongo.Password,
		s.cfg.Mongo.Database,
	)
	if err != nil {
		return err
	}
	coll := database.EnsureCollection(
		s.cfg.Mongo.Collection,
		[]mongo.IndexModel{
			{Keys: bson.D{{"emote_id", -1}}},
			{Keys: bson.D{{"count", -1}}},
			{Keys: bson.D{{"flags", -1}, {"count", -1}}},
		},
	)
	emotedb.SetCollections(coll)

	initCache()
	return s.subscribe(context.TODO(), s.handleMessage)
}
