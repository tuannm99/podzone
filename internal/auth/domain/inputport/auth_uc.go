package inputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/auth/domain/dto"
)

type AuthUsecase interface {
	GenerateOAuthURL(ctx context.Context) (string, error)
	HandleOAuthCallback(ctx context.Context, code, state string) (*dto.GoogleCallbackResp, error)
	Login(ctx context.Context, username, password string) (*dto.LoginResp, error)
	Register(ctx context.Context, req dto.RegisterReq) (*dto.RegisterResp, error)
	Logout(ctx context.Context) (string, error)
}
