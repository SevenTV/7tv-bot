package main

import (
	"go.uber.org/zap"

	"github.com/seventv/7tv-bot/internal/aggregator"
	"github.com/seventv/7tv-bot/internal/aggregator/config"
)

func main() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
	cfg := config.New()

	if cfg.LogLevel == "debug" {
		logger, _ = zap.NewDevelopment()
		zap.ReplaceGlobals(logger)
	}
	zap.S().Debug("logging on debug level")

	svc := aggregator.New(cfg)
	err := svc.Init()
	if err != nil {
		panic(err)
	}
	select {
	// TODO: shutdown
	}
}
