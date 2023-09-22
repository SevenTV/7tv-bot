package main

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/seventv/twitch-irc-reader/config"
	"github.com/seventv/twitch-irc-reader/internal/database"
	"github.com/seventv/twitch-irc-reader/internal/irc-reader"
)

func main() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)

	cfg := config.New()

	err := database.Connect(cfg.Mongo.ConnectionString, cfg.Mongo.Database)
	if err != nil {
		panic(err)
	}
	database.EnsureCollection(cfg.Mongo.Collection)

	svc := irc_reader.New(cfg)
	err = svc.Init()
	if err != nil {
		panic(err)
	}
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-shutdown:
		zap.L().Info("Shutting down...")
		svc.Shutdown()
	}
}
