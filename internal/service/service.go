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
	nc, err := nats.Connect(c.cfg.Nats.URL)
	if err != nil {
		return err
	}
	// make sure all messages are actually written to NATS on shutdown
	defer nc.Flush()
	c.queue = nc

	c.twitch = manager.New(c.cfg.Twitch.User, c.cfg.Twitch.Oauth)
	c.twitch.OnMessage(c.onMessage)

	// feed back twitch channels that got disconnected to the IRC
	go c.handleOrphanedChannels()

	err = c.twitch.Init()
	if err != nil {
		return err
	}

	// TODO: mongo & redis

	return nil
}
