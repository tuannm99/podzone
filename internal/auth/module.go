package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/controller/grpchandler"
	"github.com/tuannm99/podzone/internal/auth/domain"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/repository"
	"github.com/tuannm99/podzone/internal/auth/migrations"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdsql"
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

func RegisterGRPCServer(server *grpc.Server, authServer *grpchandler.AuthServer, logger pdlog.Logger) {
	logger.Info("Registering Auth GRPC handler")
	pbauthv1.RegisterAuthServiceServer(server, authServer)
}

type MigrateParams struct {
	fx.In
	Logger       pdlog.Logger
	V            *koanf.Koanf
	AuthDBConfig *pdsql.Config `name:"sql-auth-config"`
	AuthDB       *sqlx.DB      `name:"sql-auth"`
}

var applyMigration = func(ctx context.Context, db *sql.DB, dialect string) error {
	return migrations.Apply(ctx, db, dialect)
}

func RegisterMigration(p MigrateParams) {
	if !p.AuthDBConfig.ShouldRunMigration {
		p.Logger.Info("Disabled migration ...")
		return
	}

	p.Logger.Info("Migrating database with goose...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := applyMigration(ctx, p.AuthDB.DB, "postgres"); err != nil {
		p.Logger.Error("Migration failed", "err", err)
		return
	}
	p.Logger.Info("Migration completed")
}
