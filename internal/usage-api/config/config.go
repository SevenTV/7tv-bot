package config

import (
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

type Config struct {
	LogLevel string
	Http     struct {
		Port string
	}
	Mongo struct {
		ConnectionString string
		Database         string
		Collection       string
		Username         string
		Password         string
	}
	Nats struct {
		URL   string
		Topic struct {
			Emotes string
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
	loader := config.NewWithOptions("loader", config.ParseTime)
	loader.AddDriver(yaml.Driver)

	err := loader.LoadFiles("config.yaml")
	if err != nil {
		panic(err)
	}

	err = loader.Decode(cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
