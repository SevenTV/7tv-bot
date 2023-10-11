package usage_api

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/seventv/7tv-bot/internal/usage-api/config"
	"github.com/seventv/7tv-bot/pkg/database"
	"github.com/seventv/7tv-bot/pkg/database/emotes"
	"github.com/seventv/7tv-bot/pkg/pubsub"
	"github.com/seventv/7tv-bot/pkg/router"
)

type Server struct {
	cfg    *config.Config
	router *router.Router
	nc     *nats.Conn

	upgrader websocket.Upgrader
}

func New(cfg *config.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Init() error {
	s.router = router.New().WithRoutes(s.routes())
	server := http.Server{
		Addr:    "0.0.0.0:" + s.cfg.Http.Port,
		Handler: s.router.Router,
	}

	s.upgrader.CheckOrigin = func(r *http.Request) bool {
		// TODO: check origin
		return true
	}

	var err error
	s.nc, err = nats.Connect(s.cfg.Nats.URL)
	if err != nil {
		return err
	}

	err = database.Connect(
		s.cfg.Mongo.ConnectionString,
		s.cfg.Mongo.Username,
		s.cfg.Mongo.Password,
		s.cfg.Mongo.Database)
	if err != nil {
		return err
	}

	emotes.SetCollections(database.EnsureCollection(
		s.cfg.Mongo.Collection,
		[]mongo.IndexModel{
			{Keys: bson.D{{"emote_id", -1}}},
			{Keys: bson.D{{"count", -1}}},
			{Keys: bson.D{{"flags", -1}, {"count", -1}}},
		}))

	pubsub.Init()
	s.natsSubscribe()

	go func() {
		if err := server.ListenAndServe(); err != nil {
			zap.S().Fatal("failed to start server: ", err)
		}
	}()
	return nil
}
