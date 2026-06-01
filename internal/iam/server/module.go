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
	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdsql"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var Module = fx.Options(
	iam.Module,
	SharedModule,
	fx.Provide(
		iamgrpchandler.NewIAMCommandServer,
		iamgrpchandler.NewIAMQueryServer,
		iamgrpchandler.NewIAMServer,
	),
	fx.Invoke(
		RegisterGRPCServer,
		RegisterMigration,
	),
)

var CommandModule = fx.Options(
	iam.CommandModule,
	SharedModule,
	fx.Provide(iamgrpchandler.NewIAMCommandServer),
	fx.Invoke(
		RegisterCommandGRPCServer,
		RegisterMigration,
	),
)

var QueryModule = fx.Options(
	iam.QueryModule,
	SharedModule,
	fx.Provide(iamgrpchandler.NewIAMQueryServer),
	fx.Invoke(
		RegisterQueryGRPCServer,
		RegisterMigration,
	),
)

var LegacyModule = fx.Options(
	iam.Module,
	SharedModule,
	fx.Provide(
		iamgrpchandler.NewIAMCommandServer,
		iamgrpchandler.NewIAMQueryServer,
		iamgrpchandler.NewIAMServer,
	),
	fx.Invoke(
		RegisterLegacyGRPCServer,
		RegisterMigration,
	),
)

var SharedModule = fx.Provide(
	iamconfig.NewServerConfig,
	fx.Annotate(iamrepo.NewAuditLogRepository, fx.As(new(outputport.AuditLogRepository))),
	fx.Annotate(authclient.NewUserDirectory, fx.As(new(outputport.UserDirectory))),
)

// Module currently exposes the legacy IAM service plus the CQRS command/query
// services from one runtime. The registration functions below are intentionally
// split so a future cmd/iam-command or cmd/iam-query binary can reuse the same
// wiring and register only its side without changing the handler contracts.
func RegisterGRPCServer(
	server *grpc.Server,
	iamServer *iamgrpchandler.IAMServer,
	commandServer *iamgrpchandler.IAMCommandServer,
	queryServer *iamgrpchandler.IAMQueryServer,
	logger pdlog.Logger,
) {
	logger.Info("Registering IAM GRPC handlers")
	RegisterLegacyGRPCServer(server, iamServer)
	RegisterCommandGRPCServer(server, commandServer)
	RegisterQueryGRPCServer(server, queryServer)
}

func RegisterLegacyGRPCServer(server *grpc.Server, iamServer *iamgrpchandler.IAMServer) {
	pbiamv1.RegisterIAMServiceServer(server, iamServer)
}

func RegisterCommandGRPCServer(server *grpc.Server, commandServer *iamgrpchandler.IAMCommandServer) {
	pbiamv1.RegisterIAMCommandServiceServer(server, commandServer)
}

func RegisterQueryGRPCServer(server *grpc.Server, queryServer *iamgrpchandler.IAMQueryServer) {
	pbiamv1.RegisterIAMQueryServiceServer(server, queryServer)
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
