package grpchandler

import (
	"context"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/services/auth/domain/inputport"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	usecase inputport.AuthUsecase
}

func NewAuthServer(usecase inputport.AuthUsecase) *AuthServer {
	return &AuthServer{
		usecase: usecase,
	}
}

func (s *AuthServer) GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.GoogleLoginResponse, error) {
	authURL, err := s.usecase.GenerateOAuthURL(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GoogleLoginResponse{
		RedirectUrl: authURL,
	}, nil
}

func (s *AuthServer) GoogleCallback(
	ctx context.Context,
	req *pb.GoogleCallbackRequest,
) (*pb.GoogleCallbackResponse, error) {
	resp, err := s.usecase.HandleOAuthCallback(ctx, req.Code, req.State)
	if err != nil {
		return nil, err
	}

	return &pb.GoogleCallbackResponse{
		JwtToken:    resp.JwtToken,
		RedirectUrl: resp.RedirectUrl,
		UserInfo: &pb.UserInfo{
			Id:            resp.UserInfoResp.Id,
			Email:         resp.UserInfoResp.Email,
			Name:          resp.UserInfoResp.Name,
			GivenName:     resp.UserInfoResp.GivenName,
			FamilyName:    resp.UserInfoResp.FamilyName,
			Picture:       resp.UserInfoResp.Picture,
			EmailVerified: resp.UserInfoResp.EmailVerified,
		},
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	redirectUrl, _ := s.usecase.Logout(ctx)

	return &pb.LogoutResponse{
		Success:     true,
		RedirectUrl: redirectUrl,
	}, nil
}
