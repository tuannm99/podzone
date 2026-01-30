package auth

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"
	"google.golang.org/grpc"

	"github.com/tuannm99/podzone/internal/auth/controller/grpchandler"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdsql"
	"github.com/tuannm99/podzone/pkg/testkit"
)

func provideNamedSQLX(t *testing.T, shouldRunMigration bool) (*sqlx.DB, fx.Option) {
	t.Helper()
	db := testkit.PostgresDB(t)

	cfg := &pdsql.Config{
		URI:                testkit.PostgresDSN(t),
		Provider:           "postgres",
		ShouldRunMigration: shouldRunMigration,
	}

	opt := fx.Options(
		fx.Supply(
			fx.Annotate(db, fx.ResultTags(`name:"sql-auth"`)),
			fx.Annotate(cfg, fx.ResultTags(`name:"sql-auth-config"`)),
		),
	)
	return db, opt
}

func provideNamedRedisStub(t *testing.T) (*redis.Client, fx.Option) {
	t.Helper()
	rdb := testkit.RedisClient(t)
	opt := fx.Options(
		fx.Supply(
			fx.Annotate(rdb, fx.ResultTags(`name:"redis-auth"`)),
		),
		fx.Provide(
			fx.Annotate(
				func(c *redis.Client) redis.Cmdable { return c },
				fx.ParamTags(`name:"redis-auth"`),
				fx.ResultTags(`name:"redis-auth"`),
			),
		),
	)

	return rdb, opt
}

// --- TESTS ---

func TestRegisterMigration_Disabled(t *testing.T) {
	sqlxDB, sqlOpt := provideNamedSQLX(t, false)
	rdb, rdbOpt := provideNamedRedisStub(t)
	defer func() {
		_ = sqlxDB.Close()
		_ = rdb.Close()
	}()

	origApply := applyMigration
	applied := false
	applyMigration = func(ctx context.Context, db *sql.DB, driver string) error {
		applied = true
		return nil
	}
	defer func() { applyMigration = origApply }()

	app := fxtest.New(
		t,
		Module,
		sqlOpt,
		rdbOpt,
		fx.Supply(
			fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger))),
		),
		fx.Supply(koanf.New(".")),
		fx.Supply(grpc.NewServer()),
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
	)

	app.RequireStart()
	app.RequireStop()
	require.False(t, applied, "migration should be disabled")
}

func TestRegisterGRPCServer(t *testing.T) {
	srv := grpc.NewServer()
	authSrv := &grpchandler.AuthServer{}
	RegisterGRPCServer(srv, authSrv, pdlog.NopLogger{})
}
