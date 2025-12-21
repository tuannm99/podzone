package grpchandler

import (
	"context"

	"github.com/tuannm99/podzone/internal/auth/domain/dto"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/pkg/toolkit"

	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

type AuthServer struct {
	pbauthv1.UnimplementedAuthServiceServer
	usecase inputport.AuthUsecase
}

func NewAuthServer(usecase inputport.AuthUsecase) *AuthServer {
	return &AuthServer{
		usecase: usecase,
	}
}

func (s *AuthServer) GoogleLogin(
	ctx context.Context,
	req *pbauthv1.GoogleLoginRequest,
) (*pbauthv1.GoogleLoginResponse, error) {
	authURL, err := s.usecase.GenerateOAuthURL(ctx)
	if err != nil {
		return nil, err
	}
	return &pbauthv1.GoogleLoginResponse{
		RedirectUrl: authURL,
	}, nil
}

func (s *AuthServer) GoogleCallback(
	ctx context.Context,
	req *pbauthv1.GoogleCallbackRequest,
) (*pbauthv1.GoogleCallbackResponse, error) {
	callbackResp, err := s.usecase.HandleOAuthCallback(ctx, req.Code, req.State)
	if err != nil {
		return nil, err
	}
	resp, err := toolkit.MapStruct[dto.GoogleCallbackResp, pbauthv1.GoogleCallbackResponse](*callbackResp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pbauthv1.LogoutRequest) (*pbauthv1.LogoutResponse, error) {
	redirectUrl, _ := s.usecase.Logout(ctx)
	return &pbauthv1.LogoutResponse{
		Success:     true,
		RedirectUrl: redirectUrl,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pbauthv1.LoginRequest) (*pbauthv1.LoginResponse, error) {
	loginResp, err := s.usecase.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	resp, err := toolkit.MapStruct[dto.LoginResp, pbauthv1.LoginResponse](*loginResp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *AuthServer) Register(ctx context.Context, req *pbauthv1.RegisterRequest) (*pbauthv1.RegisterResponse, error) {
	registerDto, err := toolkit.MapStruct[*pbauthv1.RegisterRequest, dto.RegisterReq](req)
	if err != nil {
		return nil, err
	}
	registerResp, err := s.usecase.Register(ctx, *registerDto)
	if err != nil {
		return nil, err
	}
	resp, err := toolkit.MapStruct[dto.RegisterResp, pbauthv1.RegisterResponse](*registerResp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
