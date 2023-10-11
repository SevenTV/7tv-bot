package irc_reader

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"

	"github.com/seventv/7tv-bot/internal/database"
	"github.com/seventv/7tv-bot/internal/irc-reader/config"
	"github.com/seventv/7tv-bot/pkg/manager"
	"github.com/seventv/7tv-bot/pkg/ratelimit"
)

type Controller struct {
	cfg       *config.Config
	jetStream nats.JetStreamContext
	kube      *kubernetes.Clientset
	twitch    *manager.IRCManager

	shardID int

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
	if c.cfg.Replicas > 1 {
		c.shardID = getShardID()
	}
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

	if c.cfg.Kube.Oauthsecret != "" {
		err = c.kubeInit()
		if err != nil {
			return err
		}
	}

	oauth := c.cfg.Twitch.Oauth
	if oauth == "" {
		oauth, err = c.getOauthFromKubeSecret(context.Background())
		if err != nil {
			return err
		}
	}

	// initialize twitch IRC manager with ratelimit
	c.twitch = manager.New(c.cfg.Twitch.User, oauth).
		WithLimit(ratelimit.New(
			redisClient,
			c.cfg.RateLimit.Join,
			c.cfg.RateLimit.Auth,
			c.cfg.RateLimit.Reset))
	c.twitch.OnMessage(c.onMessage)

	// watch for config changes to OAuth
	config.OnChange = func() {
		if c.cfg.Twitch.Oauth == "" {
			return
		}
		c.twitch.UpdateOauth(c.cfg.Twitch.Oauth)
	}

	// watch for changes to OAuth in kubernetes secret
	if c.cfg.Kube.Oauthsecret != "" {
		err = c.watchKube(context.Background(), c.updateOauthFromKubeSecret)
		if err != nil {
			return err
		}
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

	// get changes to database over NATS
	c.watchChanges(nc)

	return nil
}

func getShardID() int {
	env := os.Getenv("HOSTNAME")
	zap.S().Infof("HOSTNAME: %v", env)
	split := strings.Split(env, "-")
	if len(split) == 0 {
		return 0
	}
	id := split[len(split)-1]
	result, _ := strconv.Atoi(id)
	return result
}

func (c *Controller) Shutdown() {
	wg := c.twitch.Shutdown()
	wg.Wait()
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
