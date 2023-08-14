package config

import (
	"time"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

type Config struct {
	LogLevel string

	RateLimit struct {
		Join  int
		Auth  int
		Reset time.Duration
		Redis struct {
			// TODO: implement redis
		}
	}
	Twitch struct {
		User  string
		Oauth string
	}
	Mongo struct {
		ConnectionString string
		Database         string
		Collection       string
	}
	Nats struct {
		URL    string
		Stream string
		Topic  struct {
			Raw string
			// provisions support for optional JSON or other parsed format output in the future
		}
	}
	Health struct {
		Enabled bool
		Port    string
	}
	Prometheus struct {
		Enabled bool
		Port    string
	}
}

func New() *Config {
	// TODO: watch for config changes during runtime. For example twitch oauth.
	cfg := &Config{}
	c := config.NewWithOptions("loader", config.ParseTime)
	c.AddDriver(yaml.Driver)

	err := c.LoadFiles("config.yaml")
	if err != nil {
		panic(err)
	}

	err = c.Decode(cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
