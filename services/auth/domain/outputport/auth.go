package outputport

import (
	"context"
	"time"

	"golang.org/x/oauth2"
)

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
}

type GoogleOauthExternal interface {
	GetConfig() *oauth2.Config
	ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error)
	FetchUserInfo(accessToken string) (*GoogleUserInfo, error)
}

type OauthStateRepository interface {
	Get(key string) (string, error)
	Set(key string, duration time.Duration) error
	Del(key string) error
}
