package partner

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	partnerconfig "github.com/tuannm99/podzone/internal/partner/config"
	"github.com/tuannm99/podzone/internal/partner/controller/grpchandler"
	partnerdomain "github.com/tuannm99/podzone/internal/partner/domain"
	"github.com/tuannm99/podzone/internal/partner/infrastructure/repository"
	"github.com/tuannm99/podzone/internal/partner/migrations"
	pbpartnerv1 "github.com/tuannm99/podzone/pkg/api/proto/partner/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdsql"
)

var Module = fx.Options(
	fx.Provide(
		partnerconfig.NewConfigFromKoanf,
		fx.Annotate(repository.NewPartnerRepository, fx.As(new(partnerdomain.PartnerRepository))),
		fx.Annotate(partnerdomain.NewPartnerUsecase, fx.As(new(partnerdomain.PartnerUsecase))),
		fx.Annotate(grpchandler.NewTenantAuthorizer, fx.As(new(grpchandler.TenantAuthorizer))),
		grpchandler.NewPartnerServer,
	),
)

var ServerModule = fx.Options(
	Module,
	fx.Invoke(
		RegisterGRPCServer,
		RegisterMigration,
	),
)

func RegisterGRPCServer(
	server *grpc.Server,
	partnerServer *grpchandler.PartnerServer,
	logger pdlog.Logger,
) {
	logger.Info("Registering Partner GRPC handler")
	pbpartnerv1.RegisterPartnerServiceServer(server, partnerServer)
}

type MigrateParams struct {
	fx.In
	Logger           pdlog.Logger
	PartnerDBConfig *pdsql.Config `name:"sql-partner-config"`
	PartnerDB       *sqlx.DB      `name:"sql-partner"`
}

func RegisterMigration(p MigrateParams) {
	if !p.PartnerDBConfig.ShouldRunMigration {
		p.Logger.Info("Disabled migration ...")
		return
	}

	p.Logger.Info("Migrating partner database with goose...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := migrations.Apply(ctx, p.PartnerDB.DB, "postgres"); err != nil {
		p.Logger.Error("Migration failed", "err", err)
		return
	}
	p.Logger.Info("Migration completed")
}

func ApplyMigration(ctx context.Context, db *sql.DB, dialect string) error {
	return migrations.Apply(ctx, db, dialect)
}
