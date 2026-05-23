package server

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	authconfig "github.com/tuannm99/podzone/internal/auth/config"
	authoutputport "github.com/tuannm99/podzone/internal/auth/domain/outputport"
	authrepo "github.com/tuannm99/podzone/internal/auth/infrastructure/repository"
	"github.com/tuannm99/podzone/internal/iam"
	iamgrpchandler "github.com/tuannm99/podzone/internal/iam/controller/grpchandler"
	iammigrations "github.com/tuannm99/podzone/internal/iam/migrations"
	iamworker "github.com/tuannm99/podzone/internal/iam/worker"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"github.com/tuannm99/podzone/pkg/messaging"
	messagingkafka "github.com/tuannm99/podzone/pkg/messaging/kafka"
	"github.com/tuannm99/podzone/pkg/pdkafka"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdsql"
	"github.com/tuannm99/podzone/pkg/pdworker"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var Module = fx.Options(
	iam.Module,
	fx.Provide(
		authconfig.NewAuthConfig,
		fx.Annotate(authrepo.NewIAMAuditLogRepositoryImpl, fx.As(new(authoutputport.AuditLogRepository))),
		fx.Annotate(authrepo.NewIAMUserRepositoryImpl, fx.As(new(authoutputport.UserRepository))),
		iamgrpchandler.NewIAMServer,
		fx.Annotate(
			func(producer pdkafka.Producer) messaging.Publisher {
				return messagingkafka.NewPublisher(producer)
			},
			fx.ParamTags(`name:"kafka-iam-producer"`),
		),
		func(store messaging.OutboxStore, publisher messaging.Publisher) *messagingkafka.Relay {
			return messagingkafka.NewRelay(store, publisher, 100)
		},
		iamworker.NewOutboxWorker,
	),
	fx.Invoke(
		RegisterGRPCServer,
		RegisterMigration,
		func(lc fx.Lifecycle, logger pdlog.Logger, w *iamworker.OutboxWorker) {
			pdworker.StartWorker(lc, logger, w)
		},
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
