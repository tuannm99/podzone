package auth

import (
	"context"
	// "net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
)

var Module = fx.Options(
	fx.Provide(
		NewAuthServer,
	),
	fx.Provide(
		fx.Annotate(
			NewRedirectResponseModifier,
			fx.ResultTags(`group:"gateway-options"`),
		),
		fx.Annotate(
			AuthMiddleware,
			fx.ResultTags(`group:"http-middleware"`),
		),
	),
	fx.Invoke(
		InitAuthConfig,
		RegisterGRPCServer,
		RegisterGatewayHandler,
	),
)

func RegisterGRPCServer(server *grpc.Server, authServer *AuthServer) {
	pb.RegisterAuthServiceServer(server, authServer)
}

func RegisterGatewayHandler(
	mux *runtime.ServeMux,
	conn *grpc.ClientConn,
	logger *zap.Logger,
) error {
	logger.Info("Registering Auth HTTP handler (gRPC-Gateway)")
	return pb.RegisterAuthServiceHandler(context.Background(), mux, conn)
}
