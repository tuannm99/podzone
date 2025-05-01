package auth

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/services/auth/config"
	"github.com/tuannm99/podzone/services/auth/controller/grpchandler"
	"github.com/tuannm99/podzone/services/auth/controller/middleware"
	"github.com/tuannm99/podzone/services/auth/domain"
	"github.com/tuannm99/podzone/services/auth/domain/inputport"
	"github.com/tuannm99/podzone/services/auth/domain/outputport"
	"github.com/tuannm99/podzone/services/auth/infrastructure"
	"github.com/tuannm99/podzone/services/auth/infrastructure/model"
)

var Module = fx.Options(
	fx.Provide(
		config.NewAuthConfig,
		fx.Annotate(
			infrastructure.NewGoogleOauth,
			fx.As(new(outputport.GoogleOauthExternal)),
		),
		fx.Annotate(
			infrastructure.NewOauthStateRepository,
			fx.As(new(outputport.OauthStateRepository)),
		),
		fx.Annotate(
			infrastructure.NewUserRepository,
			fx.As(new(outputport.UserRepository)),
		),
		fx.Annotate(
			domain.NewUserUsecase,
			fx.As(new(inputport.UserUsecase)),
		),
		fx.Annotate(
			domain.NewAuthUsecase,
			fx.As(new(inputport.AuthUsecase)),
		),
		grpchandler.NewAuthServer,
	),
	fx.Provide(
		fx.Annotate(
			middleware.NewRedirectResponseModifier,
			fx.ResultTags(`group:"gateway-options"`),
		),
		// fx.Annotate(
		// 	middleware.AuthMiddleware,
		// 	fx.ResultTags(`group:"http-middleware"`),
		// ),
	),
	fx.Invoke(
		// register grpc auth handler for grpcserver, grpcgateway
		RegisterGRPCServer,
		RegisterGatewayHandler,
		RegisterMigration,
	),
)

func RegisterGRPCServer(server *grpc.Server, authServer *grpchandler.AuthServer, logger *zap.Logger) {
	logger.Info("Registering Auth GRPC handler")
	pb.RegisterAuthServiceServer(server, authServer)
}

func RegisterGatewayHandler(mux *runtime.ServeMux, conn *grpc.ClientConn, logger *zap.Logger) error {
	logger.Info("Registering Auth HTTP handler (gRPC-Gateway)")
	return pb.RegisterAuthServiceHandler(context.Background(), mux, conn)
}

type MigrateParams struct {
	fx.In
	Logger *zap.Logger
	DB     *gorm.DB `name:"gorm-auth"`
}

func RegisterMigration(p MigrateParams) {
	p.Logger.Info("Mirating database...")
	err := p.DB.AutoMigrate(&model.User{})
	if err != nil {
		p.Logger.Error("error migration database", zap.Error(err))
	}
}
