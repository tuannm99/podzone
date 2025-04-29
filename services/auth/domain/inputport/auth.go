package inputport

import (
	"context"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
)

type AuthUsecase interface {
	GenerateOAuthURL(ctx context.Context) (string, error)
	HandleOAuthCallback(ctx context.Context, code, state string) (*pb.GoogleCallbackResponse, error)
	VerifyToken(ctx context.Context, token string) (*pb.VerifyTokenResponse, error)
	Logout(ctx context.Context) (*pb.LogoutResponse, error)
}
