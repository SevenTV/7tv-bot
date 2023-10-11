package usage_api

import (
	"net/http"

	"github.com/seventv/7tv-bot/pkg/router"
)

func (s *Server) routes() []router.Route {
	return []router.Route{
		{
			Pattern:     "/",
			Method:      http.MethodGet,
			Handler:     s.index,
			Description: "index",
		},
		{
			Pattern:     "/emotes/top",
			Method:      http.MethodGet,
			Handler:     s.getTopEmotes,
			Description: "Get top emotes, optional limit (default 20) and page (default 1) url query parameters.",
		},
		{
			Pattern: "/ws",
			Method:  http.MethodGet,
			Handler: s.ws,
			Description: "Websocket endpoint for receiving global real time emote activity. " +
				"Emotes are sent as an array of JSON objects. " +
				"PINGs are sent every 30 seconds and must be responded to with a PONG.",
		},
	}
}
