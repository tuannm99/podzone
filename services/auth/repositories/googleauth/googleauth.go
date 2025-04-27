package googleauth

import (
	"encoding/json"
	"net/http"
	"os"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/auth/presentations/grpcgateway/dtos"
	"github.com/tuannm99/podzone/services/auth/usecases"
)

var _ usecases.GoogleOauthExternal = (*GoogleOauthExternalImpl)(nil)

type GoogleOauthExternalImpl struct {
	logger      *zap.Logger
	oauthConfig *oauth2.Config
}

func NewGoogleOauthExternal(logger *zap.Logger) *GoogleOauthExternalImpl {
	return &GoogleOauthExternalImpl{
		logger: logger,
		oauthConfig: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("OAUTH_REDIRECT_URL"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
	}
}

// GetConfig implements usecases.GoogleOauthExternal.
func (g *GoogleOauthExternalImpl) GetConfig() *oauth2.Config {
	return g.oauthConfig
}

// FetchUserInfo implements usecase.GoogleOauthExternal.
func (g *GoogleOauthExternalImpl) FetchUserInfo(accessToken string) (*usecases.GoogleUserInfo, error) {
	g.logger.Debug("Fetching user info from google")
	resp, err := http.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + accessToken)
	if err != nil {
		g.logger.Error("Failed to get user info from Google", zap.Error(err))
		return nil, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			g.logger.Warn("Error closing response body", zap.Error(err))
		}
	}()

	var ggUserInfoResp dtos.GoogleUserInfoResp
	err = json.NewDecoder(resp.Body).Decode(&ggUserInfoResp)
	if err != nil {
		g.logger.Error("Failed to decode user info", zap.Error(err))
		return nil, err
	}

	g.logger.Debug("Successfully retrieved user info",
		zap.String("email", ggUserInfoResp.Email),
		zap.String("id", ggUserInfoResp.Sub))

	return toolkit.MapStruct[dtos.GoogleUserInfoResp, usecases.GoogleUserInfo](ggUserInfoResp)
}
