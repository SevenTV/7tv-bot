package service

import (
	"github.com/nats-io/nats.go"

	"github.com/seventv/twitch-irc-reader/config"
	"github.com/seventv/twitch-irc-reader/pkg/manager"
)

type Controller struct {
	cfg    *config.Config
	queue  *nats.Conn
	twitch *manager.IRCManager
	// TODO: mongo, redis
}

func New(cfg *config.Config) *Controller {
	return &Controller{cfg: cfg}
}

func (c *Controller) Init() error {
	// TODO: implement

	return nil
}
