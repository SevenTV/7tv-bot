package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/seventv/7tv-bot/internal/database"
)

func writeError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	// TODO: no body returned for response
	if len(message) != 0 {
		w.Write([]byte(message))
	}
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	for i, route := range s.routes() {
		desc := ""
		if i != 0 {
			desc = "\n\n"
		}
		desc += fmt.Sprintf("path: %s\nmethod: %s\ndescription: %s", route.Pattern, route.Method, route.Description)

		w.Write([]byte(desc))
	}
}

func (s *Server) getChannel(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	value, ok := query["id"]
	id := true
	if !ok {
		id = false
		value, ok = query["username"]
		if !ok {
			writeError(w, http.StatusBadRequest, "Bad request")
			return
		}
	}

	if len(value) == 0 {
		writeError(w, http.StatusBadRequest, "Bad request")
		return
	}

	filter := bson.D{}
	if id {
		userId, err := strconv.Atoi(value[0])
		if err != nil {
			writeError(w, http.StatusBadRequest, "Bad request")
			return
		}
		filter = bson.D{{"user_id", int64(userId)}}
	} else {
		filter = bson.D{{"username", value[0]}}
	}

	channel, err := database.GetChannel(context.TODO(), filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			writeError(w, http.StatusNoContent, "No channel found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	data, err := json.Marshal(channel)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.Write(data)
}
