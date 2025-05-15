package grpchandler

import (
	"context"

	pbAuth "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/auth/domain/dto"
	"github.com/tuannm99/podzone/services/auth/domain/inputport"
)

type AuthServer struct {
	pbAuth.UnimplementedAuthServiceServer
	usecase inputport.AuthUsecase
}

func NewAuthServer(usecase inputport.AuthUsecase) *AuthServer {
	return &AuthServer{
		usecase: usecase,
	}
}

func (s *AuthServer) GoogleLogin(
	ctx context.Context,
	req *pbAuth.GoogleLoginRequest,
) (*pbAuth.GoogleLoginResponse, error) {
	authURL, err := s.usecase.GenerateOAuthURL(ctx)
	if err != nil {
		return nil, err
	}

	return &pbAuth.GoogleLoginResponse{
		RedirectUrl: authURL,
	}, nil
}

func (s *AuthServer) GoogleCallback(
	ctx context.Context,
	req *pbAuth.GoogleCallbackRequest,
) (*pbAuth.GoogleCallbackResponse, error) {
	resp, err := s.usecase.HandleOAuthCallback(ctx, req.Code, req.State)
	if err != nil {
		return nil, err
	}
	return toolkit.MapStruct[dto.GoogleCallbackResp, pbAuth.GoogleCallbackResponse](*resp), nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pbAuth.LogoutRequest) (*pbAuth.LogoutResponse, error) {
	redirectUrl, _ := s.usecase.Logout(ctx)

	return &pbAuth.LogoutResponse{
		Success:     true,
		RedirectUrl: redirectUrl,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pbAuth.LoginRequest) (*pbAuth.LoginResponse, error) {
	// s.usecase.Login(ctx)

	return &pbAuth.LoginResponse{
		JwtToken: "",
	}, nil
}

func (s *AuthServer) Register(ctx context.Context, req *pbAuth.RegisterRequest) (*pbAuth.RegisterResponse, error) {
	// s.usecase.Register(ctx)

	return &pbAuth.RegisterResponse{
		JwtToken: "",
	}, nil
}
