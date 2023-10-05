package oauth

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func (s *Service) kubeInit() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	s.kube, err = kubernetes.NewForConfig(config)
	return err
}

func (s *Service) setKubeSecret(ctx context.Context, auth *OauthResponse) error {
	secret := v1.Secret(s.cfg.Kube.Oauthsecret, s.cfg.Kube.Namespace).
		WithType("Opaque")
	secret.StringData = make(map[string]string)
	secret.StringData["access-token"] = fmt.Sprintf("oauth:%v", auth.AccessToken)
	secret.StringData["refresh-token"] = auth.RefreshToken

	opts := metav1.ApplyOptions{}
	opts.FieldManager = "7tv-auth"

	_, err := s.kube.CoreV1().Secrets(s.cfg.Kube.Namespace).Apply(ctx, secret, opts)
	return err
}
