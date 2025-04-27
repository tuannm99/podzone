package usecases

import (
	"context"

	"golang.org/x/oauth2"
)

type (
	// External entity
	GoogleUserInfo struct {
		Email string
		Name  string
		Sub   string
	}

	// Infrastructure Interface
	UserRepository interface {
		SaveUser() error
	}
	GoogleOauthExternal interface {
		FetchUserInfo(accessToken string) (*GoogleUserInfo, error)
		GetConfig() *oauth2.Config
        // LoginCallback(ctx context.Context)
	}

	UserUsecase interface {
		GenerateRedirectUrl(ctx context.Context) (authURL string, err error)
	}
)
