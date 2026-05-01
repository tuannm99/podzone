package iam

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	authpkg "github.com/tuannm99/podzone/internal/auth"
	authconfig "github.com/tuannm99/podzone/internal/auth/config"
	authgrpchandler "github.com/tuannm99/podzone/internal/auth/controller/grpchandler"
	authoutputport "github.com/tuannm99/podzone/internal/auth/domain/outputport"
	authrepo "github.com/tuannm99/podzone/internal/auth/infrastructure/repository"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdsql"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var ServerModule = fx.Options(
	Module,
	fx.Provide(
		authconfig.NewAuthConfig,
		fx.Annotate(authrepo.NewAuditLogRepositoryImpl, fx.As(new(authoutputport.AuditLogRepository))),
		authgrpchandler.NewIAMServer,
	),
	fx.Invoke(
		RegisterGRPCServer,
		RegisterMigration,
	),
)

func RegisterGRPCServer(server *grpc.Server, iamServer *authgrpchandler.AuthServer, logger pdlog.Logger) {
	logger.Info("Registering IAM GRPC handler")
	pbauthv1.RegisterIAMServiceServer(server, iamServer)
}

type MigrateParams struct {
	fx.In
	Logger       pdlog.Logger
	AuthDBConfig *pdsql.Config `name:"sql-auth-config"`
	AuthDB       *sqlx.DB      `name:"sql-auth"`
}

func RegisterMigration(p MigrateParams) {
	if !p.AuthDBConfig.ShouldRunMigration {
		p.Logger.Info("Disabled migration ...")
		return
	}

	p.Logger.Info("Migrating database with goose...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := authpkg.ApplyMigrationForIAM(ctx, p.AuthDB.DB, "postgres"); err != nil {
		p.Logger.Error("Migration failed", "err", err)
		return
	}
	p.Logger.Info("Migration completed")
}
