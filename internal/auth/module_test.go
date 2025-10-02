package auth

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"google.golang.org/grpc"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

type (
	nopLogger      = pdlog.NopLogger
	assertiveError string
)

func (e assertiveError) Error() string { return string(e) }

func provideNamedSQLX(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, fx.Option) {
	t.Helper()
	raw, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	db := sqlx.NewDb(raw, "sqlmock")
	opt := fx.Supply(fx.Annotate(db, fx.ResultTags(`name:"sql-auth"`)))
	return db, mock, opt
}

func provideNamedRedisStub() fx.Option {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0", DB: 0})
	return fx.Supply(fx.Annotate(rdb, fx.ResultTags(`name:"redis-auth"`)))
}

func TestModule_StartsAndRunsMigrations_OK(t *testing.T) {
	sqlxDB, mock, sqlOpt := provideNamedSQLX(t)
	defer func() { _ = sqlxDB.Close() }()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS users")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	app := fxtest.New(
		t,
		Module,
		sqlOpt,
		provideNamedRedisStub(),
		// supply interface pdlog.Logger
		fx.Supply(
			fx.Annotate(nopLogger{}, fx.As(new(pdlog.Logger))),
		),
		fx.Supply(viper.New()),
		fx.Supply(grpc.NewServer()),
	)

	app.RequireStart()
	app.RequireStop()

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestModule_Migration_FirstStepFails_AppStillStarts(t *testing.T) {
	sqlxDB, mock, sqlOpt := provideNamedSQLX(t)
	defer func() { _ = sqlxDB.Close() }()

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS users")).
		WillReturnError(assertiveError("boom"))

	app := fxtest.New(
		t,
		Module,
		sqlOpt,
		provideNamedRedisStub(),
		fx.Supply(
			fx.Annotate(nopLogger{}, fx.As(new(pdlog.Logger))),
		),
		fx.Supply(viper.New()),
		fx.Supply(grpc.NewServer()),
	)

	app.RequireStart()
	app.RequireStop()

	require.NoError(t, mock.ExpectationsWereMet())
}
