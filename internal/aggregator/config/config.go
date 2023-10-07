package config

import (
	"github.com/fsnotify/fsnotify"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"go.uber.org/zap"
)

// OnChange is called when a config change is detected, can be set during runtime
var OnChange func()

type Config struct {
	LogLevel   string
	Maxworkers int
	Mongo      struct {
		ConnectionString string
		Database         string
		Collection       string
		Username         string
		Password         string
	}
	Nats struct {
		URL      string
		Stream   string
		Consumer string
		Topic    struct {
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

	err = watchConfig(cfg, loader)
	if err != nil {
		zap.L().Error("config watcher", zap.String("error", err.Error()))
	}

	return cfg
}

func watchConfig(cfg *Config, loader *config.Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if !event.Has(fsnotify.Write) {
					continue
				}

				zap.L().Info("modified config", zap.String("file", event.Name))

				err := loader.ReloadFiles()
				if err != nil {
					zap.L().Error("reload config", zap.String("error", err.Error()))
				}
				err = loader.Decode(cfg)
				if err != nil {
					panic(err)
				}

				if OnChange == nil {
					continue
				}
				OnChange()
			case err, _ := <-watcher.Errors:
				if err != nil {
					zap.L().Error("watch config", zap.String("error", err.Error()))
				}
			}
		}
	}()

	files := loader.LoadedFiles()
	if len(files) == 0 {
		zap.L().Info("watching 0 files")
		return nil
	}

	for _, file := range files {
		err = watcher.Add(file)
		if err != nil {
			return err
		}
		zap.L().Info("watching config", zap.String("file", file))
	}

	return nil
}
