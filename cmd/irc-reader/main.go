package main

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/seventv/7tv-bot/internal/database"
	"github.com/seventv/7tv-bot/internal/irc-reader"
	"github.com/seventv/7tv-bot/internal/irc-reader/config"
)

func main() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
	cfg := config.New()
	if cfg.LogLevel == "debug" {
		logger, _ = zap.NewDevelopment()
		zap.ReplaceGlobals(logger)
	}
	zap.S().Debug("using debug logger")

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
