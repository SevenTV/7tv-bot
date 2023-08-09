package config

import (
	"time"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

type Config struct {
	LogLevel string

	RateLimit struct {
		Limit int
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
		Collection       string
	}
	Nats struct {
		URL   string
		Topic struct {
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
	cfg := &Config{}
	config.AddDriver(yaml.Driver)

	err := config.LoadFiles("config.yaml")
	if err != nil {
		panic(err)
	}

	err = config.Decode(cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
