package oauth

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"

	"github.com/seventv/7tv-bot/internal/oauth/config"
	"github.com/seventv/7tv-bot/pkg/router"
	"github.com/seventv/7tv-bot/pkg/util"
)

type Service struct {
	cfg           *config.Config
	router        *router.Router
	lastOauth     *OauthResponse
	tokenOverride util.Closer
	kube          *kubernetes.Clientset
}

func New(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) Init() {
	if s.cfg.Environment == "dev" {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	s.tokenOverride.Reset()
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

	err := s.kubeInit()
	if err != nil {
		zap.S().Fatal("failed to connect to kubernetes API: ", err)
	}
	zap.S().Info("connected to kubernetes API")

	if s.cfg.Twitch.Refreshtoken != "" {
		s.lastOauth = &OauthResponse{RefreshToken: s.cfg.Twitch.Refreshtoken}
	}
	if s.lastOauth == nil {
		zap.S().Warn("OAuth not set up, please follow the Authorization code flow. URI below.")
		println(s.generateUri())

		<-s.tokenOverride.C
		s.tokenOverride.Reset()
	}

	s.refreshLoop()
}

func (s *Service) refreshLoop() {
	for {
		// wait to refresh the token until 70% of the expiry duration passed, or if we get a new token through the http endpoint
		select {
		case <-time.NewTimer(time.Duration(s.lastOauth.ExpiresIn*7/10) * time.Second).C:
		case <-s.tokenOverride.C:
			s.tokenOverride.Reset()
		}

		auth, err := s.refreshToken()
		if err != nil {
			zap.S().Error("failed to get oauth token. If you see this error repeat, consider resetting the OAuth refresh token with the URI below.", err)
			// print the authentication URI in log
			println(s.generateUri())
			// wait a minute then try again
			select {
			case <-time.NewTimer(1 * time.Minute).C:
				// skip the expiry timer on next loop
				s.tokenOverride.Close()
			case <-s.tokenOverride.C:
				s.tokenOverride.Reset()
			}
			continue
		}
		err = s.setToken(auth)
		if err != nil {
			zap.S().Error("failed to store oauth token in kube secret", err)
			continue
		}
	}
}

func (s *Service) setToken(auth *OauthResponse) error {
	s.lastOauth = auth
	return s.setKubeSecret(context.TODO(), auth)
}
