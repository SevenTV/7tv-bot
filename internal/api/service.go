package api

import (
	"net/http"
	"sync"

	"go.uber.org/zap"

	"github.com/seventv/7tv-bot/internal/api/config"
	"github.com/seventv/7tv-bot/pkg/router"
)

type Server struct {
	cfg    *config.Config
	router *router.Router
	wg     sync.WaitGroup
}

func New(cfg *config.Config) *Server {
	return &Server{cfg: cfg}
}

func (s *Server) Init() error {
	s.router = router.New().WithRoutes(s.routes())
	server := http.Server{
		Addr:    "0.0.0.0:" + s.cfg.Http.Port,
		Handler: s.router.Router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			zap.S().Fatal("failed to start server: ", err)
		}
	}()
	return nil
}

func (s *Server) Shutdown() {
	close(s.router.Shutdown)
	s.wg.Wait()
}
