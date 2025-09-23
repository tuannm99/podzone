package auth

import (
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/controller/grpchandler"
	"github.com/tuannm99/podzone/internal/auth/domain"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/model"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/repository"
	pbAuth "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/pkg/pdlogv2"
)

var Module = fx.Options(
	fx.Provide(
		config.NewAuthConfig,

		fx.Annotate(repository.NewGoogleOauthImpl, fx.As(new(outputport.GoogleOauthExternal))),
		fx.Annotate(repository.NewOauthStateRepositoryImpl, fx.As(new(outputport.OauthStateRepository))),
		fx.Annotate(repository.NewUserRepositoryImpl, fx.As(new(outputport.UserRepository))),

		fx.Annotate(domain.NewTokenUsecase, fx.As(new(inputport.TokenUsecase))),
		fx.Annotate(domain.NewUserUsecase, fx.As(new(inputport.UserUsecase))),
		fx.Annotate(domain.NewAuthUsecase, fx.As(new(inputport.AuthUsecase))),

		grpchandler.NewAuthServer,
	),
	fx.Invoke(
		RegisterGRPCServer,
		RegisterMigration,
	),
)

func RegisterGRPCServer(server *grpc.Server, authServer *grpchandler.AuthServer, logger pdlogv2.Logger) {
	logger.Info("Registering Auth GRPC handler")
	pbAuth.RegisterAuthServiceServer(server, authServer)
}

type MigrateParams struct {
	fx.In
	Logger pdlogv2.Logger
	DB     *gorm.DB `name:"gorm-auth"`
	V      *viper.Viper
}

func RegisterMigration(p MigrateParams) {
	provider := strings.ToLower(p.V.GetString("postgres.auth.provider"))
	if provider == "mock" {
		p.Logger.Info("Skipping database migration (postgres provider=mock)")
		return
	}

	p.Logger.Info("Migrating database...")
	if err := p.DB.AutoMigrate(&model.User{}); err != nil {
		p.Logger.Error("error migration database", "err", err)
	}
}
