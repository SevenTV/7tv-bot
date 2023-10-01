package aggregator

import (
	"context"

	"github.com/seventv/7tv-bot/internal/aggregator/config"
)

type Service struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) Init() error {
	// TODO: database
	err := s.subscribe(context.TODO(), s.handleMessage)
	if err != nil {
		return err
	}
	return nil
}
