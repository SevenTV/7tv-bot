package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type OauthResponse struct {
	AccessToken  string   `json:"access_token"`
	ExpiresIn    int      `json:"expires_in"`
	TokenType    string   `json:"token_type"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
}

func (s *Service) getToken(code string) (*OauthResponse, error) {
	data := url.Values{}
	data.Set("client_id", s.cfg.Twitch.Clientid)
	data.Set("client_secret", s.cfg.Twitch.Clientsecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", s.cfg.Twitch.Redirecturi)

	body, err := postData(data)
	if err != nil {
		return nil, err
	}

	response := &OauthResponse{}
	err = json.Unmarshal(body, response)
	return response, err
}

func (s *Service) refreshToken() (*OauthResponse, error) {
	data := url.Values{}
	data.Set("client_id", s.cfg.Twitch.Clientid)
	data.Set("client_secret", s.cfg.Twitch.Clientsecret)
	data.Set("refresh_token", s.lastOauth.RefreshToken)
	data.Set("grant_type", "refresh_token")

	body, err := postData(data)

	response := &OauthResponse{}
	err = json.Unmarshal(body, response)
	return response, err
}

func postData(data url.Values) ([]byte, error) {
	res, err := http.PostForm("https://id.twitch.tv/oauth2/token", data)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	return body, err
}

func (s *Service) generateUri() string {
	return fmt.Sprintf(
		"https://id.twitch.tv/oauth2/authorize?response_type=code&client_id=%v&redirect_uri=%v&scope=chat%%3Aread%%20chat%%3Aedit&state=%v",
		s.cfg.Twitch.Clientid,
		s.cfg.Twitch.Redirecturi,
		s.cfg.Twitch.State)
}
