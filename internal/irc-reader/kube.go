package irc_reader

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	ErrNoOAuthToken = errors.New("no OAuth token found in kubernetes secret data")
	ErrNoSecret     = errors.New("no secret found")
)

func (c *Controller) kubeInit() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	c.kube, err = kubernetes.NewForConfig(config)
	return err
}

func (c *Controller) watchKube(ctx context.Context, cb func() error) error {
	watcher, err := c.kube.CoreV1().Secrets(c.cfg.Kube.Namespace).Watch(
		ctx,
		metav1.SingleObject(metav1.ObjectMeta{
			Name: c.cfg.Kube.Oauthsecret,
		}),
	)
	if err != nil {
		return err
	}
	go func() {
		for range watcher.ResultChan() {
			err = cb()
			if err != nil {
				zap.S().Infow("failed to update OAuth token from kubernetes secret", err)
			}
		}
	}()
	return nil
}

func (c *Controller) updateOauthFromKubeSecret() error {
	oauth, err := c.getOauthFromKubeSecret(context.Background())
	if err != nil {
		return err
	}
	c.twitch.UpdateOauth(oauth)
	zap.S().Info("updated OAuth token from kubernetes secret")
	return nil
}

func (c *Controller) getOauthFromKubeSecret(ctx context.Context) (string, error) {
	secret, err := c.kube.CoreV1().Secrets(c.cfg.Kube.Namespace).Get(
		ctx,
		c.cfg.Kube.Oauthsecret,
		metav1.GetOptions{},
	)
	if err != nil {
		return "", err
	}
	if secret == nil {
		return "", ErrNoSecret
	}
	data, ok := secret.Data["access-token"]
	if !ok || len(data) == 0 {
		return "", ErrNoOAuthToken
	}
	return fmt.Sprintf("oauth:%v", string(data)), nil
}
