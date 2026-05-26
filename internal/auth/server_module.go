package auth

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/tuannm99/podzone/internal/auth/controller/grpchandler"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

var ServerModule = fx.Options(
	Module,
	fx.Provide(
		grpchandler.NewAuthServer,
	),
	fx.Invoke(
		RegisterGRPCServer,
		RegisterMigration,
	),
)

func RegisterGRPCServer(server *grpc.Server, authServer *grpchandler.AuthServer, logger pdlog.Logger) {
	logger.Info("Registering Auth GRPC handler")
	pbauthv1.RegisterAuthServiceServer(server, authServer)
}
