package auth

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	sq "github.com/Masterminds/squirrel"
	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/controller/grpchandler"
	"github.com/tuannm99/podzone/internal/auth/domain"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/repository"
	"github.com/tuannm99/podzone/internal/auth/migrations"
	pbAuth "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/pkg/pdlog"
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
	pbAuth.RegisterAuthServiceServer(server, authServer)
}

type MigrateParams struct {
	fx.In
	Logger pdlog.Logger
	DB     *sqlx.DB `name:"sql-auth"`
	V      *viper.Viper
}

// execDDL runs a list of DDL Sqlizers with context + logging
func execDDL(ctx context.Context, db *sqlx.DB, log pdlog.Logger, steps ...sq.Sqlizer) error {
	for _, s := range steps {
		sqlStr, args, err := s.ToSql()
		if err != nil {
			log.Error("build DDL failed", "err", err)
			return err
		}
		// args will be empty for pure DDL; safe to pass anyway
		if _, err := db.ExecContext(ctx, sqlStr, args...); err != nil {
			log.Error("exec DDL failed", "sql", sqlStr, "err", err)
			return err
		}
	}
	return nil
}

func RegisterMigration(p MigrateParams) {
	p.Logger.Info("Migrating database...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Run steps in order
	err := execDDL(ctx, p.DB, p.Logger, append(migrations.CreateExts, migrations.CreateTableUsers...)...)
	if err != nil {
		p.Logger.Error("Migration failed", "err", err)
		return
	}
	p.Logger.Info("Migration completed")
}
