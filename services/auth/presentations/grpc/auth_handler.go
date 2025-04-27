package grpc

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/services/auth/usecases"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	logger      *zap.Logger
	userUsecase *usecases.UserUcImpl
}

type AuthServerParams struct {
	fx.In

	Logger      *zap.Logger
	UserUsecase *usecases.UserUcImpl
}

func NewAuthServer(p AuthServerParams) *AuthServer {
	return &AuthServer{
		logger:      p.Logger,
		userUsecase: p.UserUsecase,
	}
}

func (s *AuthServer) GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.GoogleLoginResponse, error) {
	url, err := s.userUsecase.GenerateRedirectUrl(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GoogleLoginResponse{
		RedirectUrl: url,
	}, nil
}
