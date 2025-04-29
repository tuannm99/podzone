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
	return s.usecase.HandleOAuthCallback(ctx, req.Code, req.State)
}

func (s *AuthServer) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	return s.usecase.VerifyToken(ctx, req.Token)
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return s.usecase.Logout(ctx)
}
