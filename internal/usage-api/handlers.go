package usage_api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/seventv/7tv-bot/pkg/database/emotes"
	"github.com/seventv/7tv-bot/pkg/pubsub"
	"github.com/seventv/7tv-bot/pkg/util"
)

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

func (s *Server) getTopEmotes(w http.ResponseWriter, r *http.Request) {
	limit := chi.URLParam(r, "limit")
	if limit == "" {
		limit = "20"
	}
	page := chi.URLParam(r, "page")
	if page == "" {
		page = "1"
	}
	lim, err := strconv.Atoi(limit)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("limit must be an integer"))
		return
	}
	pg, err := strconv.Atoi(page)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("page must be an integer"))
		return
	}
	emotes, err := emotes.GetTopEmotes(context.Background(), int64(lim), int64(pg))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	data, err := json.Marshal(emotes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(data)
}

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	zap.S().Debugf("websocket connection from %s", r.RemoteAddr)
	// upgrade connection to websocket, check for errors and close if any
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.S().Error("error upgrading connection to websocket ", zap.Error(err))
		return
	}
	defer conn.Close()

	closer := util.Closer{}
	closer.Reset()
	// create pubsub connection to write messages from NATS to websocket
	pubsubconn := pubsub.NewConnection(func(data []byte) {
		if err = conn.WriteMessage(websocket.TextMessage, data); err != nil {
			zap.S().Error("error writing message to websocket, closing connection ", zap.Error(err))
			closer.Close()
		}
	})
	defer pubsubconn.Close()

	closed := false

	lastPong := time.Now()
	// listen for pong messages
	conn.SetPongHandler(func(string) error {
		zap.S().Debug("received pong")
		lastPong = time.Now()
		return nil
	})
	conn.SetPingHandler(func(string) error {
		err = conn.WriteMessage(websocket.PongMessage, nil)
		if err != nil {
			zap.S().Error("error writing pong to websocket ", zap.Error(err))
		}
		return err
	})
	conn.SetCloseHandler(func(code int, text string) error {
		zap.S().Info("websocket closed with code ", code, " and text ", text)
		closed = true
		return nil
	})

	go func() {
		for {
			// read messages from websocket to enable handlers to work
			_, _, err := conn.ReadMessage()
			if err != nil {
				zap.S().Error("error reading message from websocket ", zap.Error(err))
				closed = true
				return
			}
		}
	}()

	// ping every 30 seconds, close connection if we don't receive a pong in over 2 minutes
	for {
		if closed {
			return
		}
		if time.Since(lastPong) > 2*time.Minute {
			// send close message to websocket
			err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "connection closed due to inactivity"))
			zap.S().Error("websocket connection timed out")
			return
		}
		if err = conn.WriteMessage(websocket.PingMessage, []byte("PING")); err != nil {
			zap.S().Error("error pinging websocket ", zap.Error(err))
			return
		}
		select {
		case <-closer.C:
			return
		case <-time.After(30 * time.Second):
		}
	}
}
