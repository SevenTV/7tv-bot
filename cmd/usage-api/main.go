package main

import (
	usage_api "github.com/seventv/7tv-bot/internal/usage-api"
	"github.com/seventv/7tv-bot/internal/usage-api/config"
)

func main() {
	cfg := config.New()
	svc := usage_api.New(cfg)
	svc.Init()
	select {}
}
