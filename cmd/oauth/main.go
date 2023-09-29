package main

import (
	"go.uber.org/zap"

	"github.com/seventv/7tv-bot/internal/oauth"
	"github.com/seventv/7tv-bot/internal/oauth/config"
)

func main() {
	logger, _ := zap.NewProduction()
	zap.ReplaceGlobals(logger)
	svc := oauth.New(config.New())
	svc.Init()
	select {}
}
