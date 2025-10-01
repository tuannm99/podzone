package grpchandler

import (
	"context"

	"github.com/tuannm99/podzone/internal/auth/domain/dto"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	pbAuth "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/pkg/toolkit"
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
	callbackResp, err := s.usecase.HandleOAuthCallback(ctx, req.Code, req.State)
	if err != nil {
		return nil, err
	}
	resp, err := toolkit.MapStruct[dto.GoogleCallbackResp, pbAuth.GoogleCallbackResponse](*callbackResp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pbAuth.LogoutRequest) (*pbAuth.LogoutResponse, error) {
	redirectUrl, _ := s.usecase.Logout(ctx)
	return &pbAuth.LogoutResponse{
		Success:     true,
		RedirectUrl: redirectUrl,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pbAuth.LoginRequest) (*pbAuth.LoginResponse, error) {
	loginResp, err := s.usecase.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	resp, err := toolkit.MapStruct[dto.LoginResp, pbAuth.LoginResponse](*loginResp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AuthServer) Register(ctx context.Context, req *pbAuth.RegisterRequest) (*pbAuth.RegisterResponse, error) {
	registerDto, err := toolkit.MapStruct[*pbAuth.RegisterRequest, dto.RegisterReq](req)
	if err != nil {
		return nil, err
	}
	registerResp, err := s.usecase.Register(ctx, *registerDto)
	if err != nil {
		return nil, err
	}
	resp, err := toolkit.MapStruct[dto.RegisterResp, pbAuth.RegisterResponse](*registerResp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
