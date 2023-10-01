package aggregator

import "github.com/seventv/7tv-bot/internal/aggregator/config"

type Service struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) Init() error {

	return nil
}
