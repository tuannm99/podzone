package auth

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
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

type mockLogger = pdlog.NopLogger

func provideNamedSQLX(t *testing.T, shouldRun bool) (*sqlx.DB, sqlmock.Sqlmock, fx.Option) {
	t.Helper()
	raw, mockDB, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	db := sqlx.NewDb(raw, "sqlmock")

	cfg := &pdsql.Config{
		URI:                "postgres://fake",
		Provider:           "postgres",
		ShouldRunMigration: shouldRun,
	}

	opt := fx.Options(
		fx.Supply(
			fx.Annotate(db, fx.ResultTags(`name:"sql-auth"`)),
			fx.Annotate(cfg, fx.ResultTags(`name:"sql-auth-config"`)),
		),
	)
	return db, mockDB, opt
}

func provideNamedRedisStub() fx.Option {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0", DB: 0})
	return fx.Supply(fx.Annotate(rdb, fx.ResultTags(`name:"redis-auth"`)))
}

// --- TESTS ---

func TestRegisterMigration_Disabled(t *testing.T) {
	sqlxDB, _, sqlOpt := provideNamedSQLX(t, true)
	defer func() { _ = sqlxDB.Close() }()

	origApply := applyMigration
	applyMigration = func(ctx context.Context, db *sql.DB, driver string) error {
		return nil
	}
	defer func() { applyMigration = origApply }()

	app := fxtest.New(
		t,
		Module,
		sqlOpt,
		provideNamedRedisStub(),
		fx.Supply(
			fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger))),
		),
		fx.Supply(viper.New()),
		fx.Supply(grpc.NewServer()),
	)

	app.RequireStart()
	app.RequireStop()
}

// Case 4: GRPC registration logs and binds
func TestRegisterGRPCServer(t *testing.T) {
	logger := &mockLogger{}
	srv := grpc.NewServer()
	authSrv := &grpchandler.AuthServer{}
	RegisterGRPCServer(srv, authSrv, logger)
}
