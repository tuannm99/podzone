package auth

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redismock/v9"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"google.golang.org/grpc"

	"github.com/tuannm99/podzone/internal/auth/controller/grpchandler"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdsql"
)

func provideNamedSQLX(t *testing.T, shouldRunMigration bool) (*sqlx.DB, sqlmock.Sqlmock, fx.Option) {
	t.Helper()
	raw, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	db := sqlx.NewDb(raw, "sqlmock")

	cfg := &pdsql.Config{
		URI:                "postgres://fake",
		Provider:           "postgres",
		ShouldRunMigration: shouldRunMigration,
	}

	opt := fx.Options(
		fx.Supply(
			fx.Annotate(db, fx.ResultTags(`name:"sql-auth"`)),
			fx.Annotate(cfg, fx.ResultTags(`name:"sql-auth-config"`)),
		),
	)
	return db, mockDB, opt
}

func provideNamedRedisStub(t *testing.T) (*redis.Client, redismock.ClientMock, fx.Option) {
	t.Helper()
	rdb, mock := redismock.NewClientMock()
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

	return rdb, mock, opt
}

// --- TESTS ---

func TestRegisterMigration_Disabled(t *testing.T) {
	sqlxDB, _, sqlOpt := provideNamedSQLX(t, true)
	rdb, _, rdbOpt := provideNamedRedisStub(t)
	defer func() {
		_ = sqlxDB.Close()
		_ = rdb.Close()
	}()

	origApply := applyMigration
	applyMigration = func(ctx context.Context, db *sql.DB, driver string) error {
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
		fx.Supply(viper.New()),
		fx.Supply(grpc.NewServer()),
	)

	app.RequireStart()
	app.RequireStop()
}

func TestRegisterGRPCServer(t *testing.T) {
	srv := grpc.NewServer()
	authSrv := &grpchandler.AuthServer{}
	RegisterGRPCServer(srv, authSrv, pdlog.NopLogger{})
}
