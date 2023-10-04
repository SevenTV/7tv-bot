package main

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/seventv/7tv-bot/internal/api"
	"github.com/seventv/7tv-bot/internal/api/config"
	"github.com/seventv/7tv-bot/internal/database"
)

func main() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)

	cfg := config.New()
	err := database.Connect(
		cfg.Mongo.ConnectionString,
		cfg.Mongo.Username,
		cfg.Mongo.Password,
		cfg.Mongo.Database,
	)
	if err != nil {
		panic(err)
	}
	database.EnsureCollection(cfg.Mongo.Collection)

	server := api.New(cfg)
	err = server.Init()
	if err != nil {
		panic(err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-shutdown:
		zap.L().Info("Shutting down...")
		server.Shutdown()
	}
}
