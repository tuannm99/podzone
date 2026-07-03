package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/iamclient"
	"github.com/tuannm99/podzone/internal/auth/infrastructure/repository"
	"github.com/tuannm99/podzone/internal/auth/migrations"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdsql"
)

var Module = fx.Options(
	fx.Provide(
		config.NewAuthConfig,

		fx.Annotate(repository.NewGoogleOauthImpl, fx.As(new(outputport.GoogleOauthExternal))),
		fx.Annotate(repository.NewOauthStateRepositoryImpl, fx.As(new(outputport.OauthStateRepository))),
		fx.Annotate(repository.NewUserRepositoryImpl, fx.As(new(outputport.UserRepository))),
		fx.Annotate(repository.NewSessionRepositoryImpl, fx.As(new(outputport.SessionRepository))),
		fx.Annotate(repository.NewRefreshTokenRepositoryImpl, fx.As(new(outputport.RefreshTokenRepository))),
		fx.Annotate(repository.NewAuditLogRepositoryImpl, fx.As(new(outputport.AuditLogRepository))),
		fx.Annotate(iamclient.NewTenantAccessChecker, fx.As(new(outputport.TenantAccessChecker))),
		fx.Annotate(iamclient.NewRoleAssumer, fx.As(new(outputport.RoleAssumer))),
		fx.Annotate(iamclient.NewAccountBootstrapper, fx.As(new(outputport.AccountBootstrapper))),

		fx.Annotate(domain.NewTokenUsecase, fx.As(new(inputport.TokenUsecase))),
		fx.Annotate(domain.NewUserUsecase, fx.As(new(inputport.UserUsecase))),
		fx.Annotate(domain.NewAuthUsecase, fx.As(new(inputport.AuthUsecase))),
	),
)

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
