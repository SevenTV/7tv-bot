package oauth

import (
	"net/http"

	"go.uber.org/zap"
)

func (s *Service) index(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	code, ok := values["code"]
	if !ok || len(code) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	state, ok := values["state"]
	if !ok || len(code) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if state[0] != s.cfg.Twitch.State {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	auth, err := s.getToken(code[0])
	if err != nil {
		zap.S().Error("failed to get oauth token from twitch", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.setToken(auth)
	if err != nil {
		zap.S().Error("failed to store oauth token in kube secret", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// send signal to the service that we are overriding the current tokens
	s.tokenOverride.Close()

	zap.L().Info("OAuth manually set")

	w.Write([]byte("OK"))
}
