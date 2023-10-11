package main

import (
	"go.uber.org/zap"

	usage_api "github.com/seventv/7tv-bot/internal/usage-api"
	"github.com/seventv/7tv-bot/internal/usage-api/config"
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

	svc := usage_api.New(cfg)
	svc.Init()
	select {}
}
