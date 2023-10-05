package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/seventv/7tv-bot/internal/database"
	"github.com/seventv/7tv-bot/pkg/bitwise"
	"github.com/seventv/7tv-bot/pkg/types"
	"github.com/seventv/7tv-bot/pkg/util"
)

func writeError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
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

func (s *Server) postChannel(w http.ResponseWriter, r *http.Request) {
	channel := types.Channel{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Bad request")
		return
	}
	err = json.Unmarshal(body, &channel)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Bad request")
		return
	}
	if !util.VerifyChannel(channel) {
		writeError(w, http.StatusBadRequest, "Bad request")
		return
	}

	err = database.InsertChannel(context.TODO(), channel)
	if err != nil {
		if errors.Is(err, database.ErrChannelAlreadyExists) {
			writeError(w, http.StatusAlreadyReported, "Already reported")
			return
		}
		zap.S().Error(err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	writeError(w, http.StatusCreated, "Created")

	// publish insert to NATS
	header := make(nats.Header)
	header.Add("OP", database.Insert)
	s.nc.PublishMsg(&nats.Msg{
		Data:    body,
		Header:  header,
		Subject: s.cfg.Nats.Topic.Api,
	})
}

func (s *Server) deleteChannel(w http.ResponseWriter, r *http.Request) {
	query, ok := r.URL.Query()["id"]
	if !ok {
		writeError(w, http.StatusBadRequest, "Bad request")
		return
	}
	if len(query) == 0 {
		writeError(w, http.StatusBadRequest, "Bad request")
		return
	}
	id, err := strconv.Atoi(query[0])
	if err != nil {
		writeError(w, http.StatusBadRequest, "Bad request")
		return
	}

	// Get channel, to see if it exists & so we can pass the original channel to the IRC reader over NATS
	filter := bson.D{{"user_id", int64(id)}}
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

	err = database.DeleteChannel(context.TODO(), int64(id))
	if err != nil {
		if errors.Is(err, database.ErrChannelNotFound) {
			writeError(w, http.StatusNoContent, "No channel found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.Write([]byte("OK"))

	// publish deleted channel ID to NATS
	header := make(nats.Header)
	header.Add("OP", database.Delete)
	s.nc.PublishMsg(&nats.Msg{
		Data:    data,
		Header:  header,
		Subject: s.cfg.Nats.Topic.Api,
	})
}

func (s *Server) postChannels(w http.ResponseWriter, r *http.Request) {
	channels := []types.Channel{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		zap.S().Errorw("read channel array body", "error", err)
		writeError(w, http.StatusBadRequest, "Bad request")
		return
	}
	err = json.Unmarshal(body, &channels)
	if err != nil {
		zap.S().Errorw("channel array unmarshal", "error", err)
		writeError(w, http.StatusBadRequest, "Bad request")
		return
	}
	for _, channel := range channels {
		if channel.Flags == 0 {
			channel.Flags = bitwise.Set(channel.Flags, bitwise.JOIN_IRC)
		}
		if channel.Weight == 0 {
			channel.Weight = 1
		}
		if channel.Platform == "" {
			channel.Platform = "twitch"
		}
		if !util.VerifyChannel(channel) {
			continue
		}

		err = database.InsertChannel(context.TODO(), channel)
		if err != nil {
			zap.S().Errorw("insert channel to database", "error", err)
			continue
		}

		// publish insert to NATS
		data, err := json.Marshal(channel)
		if err != nil {
			zap.S().Errorw("NATS publish data marshal", "error", err)
			continue
		}
		header := make(nats.Header)
		header.Add("OP", database.Insert)
		s.nc.PublishMsg(&nats.Msg{
			Data:    data,
			Header:  header,
			Subject: s.cfg.Nats.Topic.Api,
		})
	}
	writeError(w, http.StatusCreated, "Created")
}

func notImplemented(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "Not implemented")
}
