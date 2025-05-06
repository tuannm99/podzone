package auth

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	pbAuth "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/services/auth/config"
	"github.com/tuannm99/podzone/services/auth/controller/grpchandler"
	"github.com/tuannm99/podzone/services/auth/domain"
	"github.com/tuannm99/podzone/services/auth/domain/inputport"
	"github.com/tuannm99/podzone/services/auth/domain/outputport"
	"github.com/tuannm99/podzone/services/auth/infrastructure/model"
	"github.com/tuannm99/podzone/services/auth/infrastructure/repository"
)

var Module = fx.Options(
	fx.Provide(
		config.NewAuthConfig,
		fx.Annotate(
			repository.NewGoogleOauthImpl,
			fx.As(new(outputport.GoogleOauthExternal)),
		),
		fx.Annotate(
			repository.NewOauthStateRepositoryImpl,
			fx.As(new(outputport.OauthStateRepository)),
		),
		fx.Annotate(
			repository.NewUserRepositoryImpl,
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
	fx.Invoke(
		RegisterGRPCServer,
		RegisterMigration,
	),
)

func RegisterGRPCServer(server *grpc.Server, authServer *grpchandler.AuthServer, logger *zap.Logger) {
	logger.Info("Registering Auth GRPC handler")
	pbAuth.RegisterAuthServiceServer(server, authServer)
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
