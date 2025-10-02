package repository

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
)

var _ outputport.GoogleOauthExternal = (*googleOauthImpl)(nil)

type googleOauthImpl struct {
	config      *oauth2.Config
	httpClient  *http.Client
	userInfoURL string
}

func NewGoogleOauthImpl() *googleOauthImpl {
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
	return &googleOauthImpl{config: config}
}

func NewGoogleOauthImplWithOptions(cfg *oauth2.Config, hc *http.Client, userInfoURL string) *googleOauthImpl {
	if hc == nil {
		hc = http.DefaultClient
	}
	if userInfoURL == "" {
		userInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
	}
	return &googleOauthImpl{
		config:      cfg,
		httpClient:  hc,
		userInfoURL: userInfoURL,
	}
}

func (g *googleOauthImpl) GetConfig() *oauth2.Config {
	return g.config
}

func (g *googleOauthImpl) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.config.Exchange(ctx, code)
}

func (g *googleOauthImpl) FetchUserInfo(accessToken string) (*outputport.GoogleUserInfo, error) {
	req, err := http.NewRequest("GET", g.userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("access_token", accessToken)
	req.URL.RawQuery = q.Encode()

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var userInfo outputport.GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	return &userInfo, nil
}
