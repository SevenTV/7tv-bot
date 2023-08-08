package config

import "time"

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
	// TODO: find good config loader
	return &Config{}
}
