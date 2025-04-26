package externals

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/auth/usecase"
	"github.com/tuannm99/podzone/services/auth/web/dtos"
)

var _ usecase.GoogleOauthExternal = (*GoogleOauthExternal)(nil)

type GoogleOauthExternal struct {
	logger *zap.Logger
}

func NewGoogleOauthExternal(logger *zap.Logger) *GoogleOauthExternal {
	return &GoogleOauthExternal{logger: logger}
}

// FetchUserInfo implements usecase.GoogleOauthExternal.
func (g *GoogleOauthExternal) FetchUserInfo(accessToken string) (*usecase.GoogleUserInfo, error) {
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

	return toolkit.MapStruct[dtos.GoogleUserInfoResp, usecase.GoogleUserInfo](ggUserInfoResp)
}
