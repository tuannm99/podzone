package server

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/internal/iam"
	iamconfig "github.com/tuannm99/podzone/internal/iam/config"
	iamgrpchandler "github.com/tuannm99/podzone/internal/iam/controller/grpchandler"
	"github.com/tuannm99/podzone/internal/iam/domain/outputport"
	"github.com/tuannm99/podzone/internal/iam/infrastructure/authclient"
	iamrepo "github.com/tuannm99/podzone/internal/iam/infrastructure/repository"
	iammigrations "github.com/tuannm99/podzone/internal/iam/migrations"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdsql"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var Module = fx.Options(
	iam.Module,
	fx.Provide(
		iamconfig.NewServerConfig,
		fx.Annotate(iamrepo.NewAuditLogRepository, fx.As(new(outputport.AuditLogRepository))),
		fx.Annotate(authclient.NewUserDirectory, fx.As(new(outputport.UserDirectory))),
		iamgrpchandler.NewIAMServer,
	),
	fx.Invoke(
		RegisterGRPCServer,
		RegisterMigration,
	),
)

func RegisterGRPCServer(server *grpc.Server, iamServer *iamgrpchandler.IAMServer, logger pdlog.Logger) {
	logger.Info("Registering IAM GRPC handler")
	pbauthv1.RegisterIAMServiceServer(server, iamServer)
}

type MigrateParams struct {
	fx.In
	Logger      pdlog.Logger
	IAMDBConfig *pdsql.Config `name:"sql-iam-config"`
	IAMDB       *sqlx.DB      `name:"sql-iam"`
}

func RegisterMigration(p MigrateParams) {
	if !p.IAMDBConfig.ShouldRunMigration {
		p.Logger.Info("Disabled migration ...")
		return
	}

	p.Logger.Info("Migrating database with goose...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := iammigrations.Apply(ctx, p.IAMDB.DB, "postgres"); err != nil {
		p.Logger.Error("Migration failed", "err", err)
		return
	}
	p.Logger.Info("Migration completed")
}
