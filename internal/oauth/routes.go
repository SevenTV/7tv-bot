package oauth

import (
	"net/http"

	"github.com/seventv/7tv-bot/pkg/router"
)

func (s *Service) routes() []router.Route {
	return []router.Route{
		{
			Pattern:     "/",
			Method:      http.MethodGet,
			Handler:     s.index,
			Description: "index",
		},
		{
			Pattern:     "/",
			Method:      http.MethodPost,
			Handler:     s.index,
			Description: "index",
		},
	}
}
