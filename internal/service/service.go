package service

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"

	"github.com/seventv/twitch-irc-reader/config"
	"github.com/seventv/twitch-irc-reader/internal/database"
	"github.com/seventv/twitch-irc-reader/pkg/manager"
	"github.com/seventv/twitch-irc-reader/pkg/ratelimit"
)

type Controller struct {
	cfg       *config.Config
	jetStream nats.JetStreamContext
	twitch    *manager.IRCManager

	// limit amount of workers for joining channels
	joinSem chan struct{}
}

func New(cfg *config.Config) *Controller {
	return &Controller{
		cfg:     cfg,
		joinSem: make(chan struct{}, 10),
	}
}

func (c *Controller) Init() error {
	nc, err := nats.Connect(c.cfg.Nats.URL)
	if err != nil {
		return err
	}
	// make sure all messages are actually written to NATS on shutdown
	defer nc.Flush()

	js, _ := nc.JetStream()

	err = c.ensureStream(js)
	if err != nil {
		return err
	}

	c.jetStream = js

	redisClient, err := c.initializeRedis()
	if err != nil {
		return err
	}

	// initialize twitch IRC manager with ratelimit
	c.twitch = manager.New(c.cfg.Twitch.User, c.cfg.Twitch.Oauth).
		WithLimit(ratelimit.New(
			redisClient,
			c.cfg.RateLimit.Join,
			c.cfg.RateLimit.Auth,
			c.cfg.RateLimit.Reset))
	c.twitch.OnMessage(c.onMessage)

	// watch for config changes to OAuth
	config.OnChange = func() {
		c.twitch.UpdateOauth(c.cfg.Twitch.Oauth)
	}

	// feed back twitch channels that got disconnected to the IRC
	go c.handleOrphanedChannels()

	err = c.twitch.Init()
	if err != nil {
		return err
	}

	database.GetChannels(
		context.Background(),
		c.joinChannels,
		20,
	)

	return nil
}

func (c *Controller) initializeRedis() (*redis.Client, error) {
	return ratelimit.RedisClient(ratelimit.RedisOptions{
		MasterName: c.cfg.RateLimit.Redis.Master,
		Username:   c.cfg.RateLimit.Redis.Username,
		Password:   c.cfg.RateLimit.Redis.Password,
		Database:   c.cfg.RateLimit.Redis.Database,
		Addresses:  c.cfg.RateLimit.Redis.Addresses,
		Sentinel:   c.cfg.RateLimit.Redis.Sentinel,
		EnableSync: true,
	})
}
