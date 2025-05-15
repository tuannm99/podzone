package inputport

import (
	"context"

	"github.com/tuannm99/podzone/services/auth/domain/dto"
)

type AuthUsecase interface {
	GenerateOAuthURL(ctx context.Context) (string, error)
	HandleOAuthCallback(ctx context.Context, code, state string) (*dto.GoogleCallbackResp, error)
    Login(ctx context.Context)
    Register(ctx context.Context)
	Logout(ctx context.Context) (string, error)
}
