package api

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
			Pattern:     "/twitch/channel",
			Method:      http.MethodGet,
			Handler:     s.getChannel,
			Description: "Get channel by id (?id=) or username (?username=) url query parameter.",
		},
		{
			Pattern:     "/twitch/channel",
			Method:      http.MethodPost,
			Handler:     s.postChannel,
			Description: "Set channel using JSON body, gives error if channel already exists.",
		},
		// TODO: PUT channel
		{
			Pattern:     "/twitch/channel",
			Method:      http.MethodPut,
			Handler:     notImplemented,
			Description: "Update channel using JSON body, matches with given channelID, gives error if ID does not exist yet.",
		},
		{
			Pattern:     "/twitch/channel",
			Method:      http.MethodDelete,
			Handler:     s.deleteChannel,
			Description: "Delete channel matching the id (?id=) url query parameter.",
		},
	}
}
