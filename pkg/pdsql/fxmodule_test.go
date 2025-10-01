package pdsql

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx/fxtest"
)

var mockLogger = &pdlog.NopLogger{}

func newSQLXMock(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	t.Helper()

	raw, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err, "sqlmock.NewWithOptions")

	sqlxDB := sqlx.NewDb(raw, "sqlmock")
	cleanup := func() { _ = sqlxDB.Close() }
	return sqlxDB, mock, cleanup
}

func TestRegisterLifecycle_StartStop_OK(t *testing.T) {
	db, mock, cleanup := newSQLXMock(t)
	defer cleanup()

	mock.ExpectPing()
	mock.ExpectClose()

	lc := fxtest.NewLifecycle(t)
	cfg := &Config{URI: "sqlmock://ok"}

	registerLifecycle(lc, db, mockLogger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, lc.Start(ctx), "lc.Start should succeed")
	require.NoError(t, lc.Stop(ctx), "lc.Stop should succeed")

	require.NoError(t, mock.ExpectationsWereMet(), "sqlmock expectations")
}

func TestRegisterLifecycle_Start_ContextCanceled(t *testing.T) {
	db, mock, cleanup := newSQLXMock(t)
	defer cleanup()

	mock.ExpectPing().WillReturnError(errors.New("transient ping error"))

	lc := fxtest.NewLifecycle(t)
	cfg := &Config{URI: "sqlmock://cancel"}
	registerLifecycle(lc, db, mockLogger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel()

	err := lc.Start(ctx)
	require.Error(t, err, "lc.Start should return error when context cancels")

	require.NoError(t, mock.ExpectationsWereMet(), "sqlmock expectations")
}

func TestRegisterLifecycle_Stop_DBIsNil_NoOp(t *testing.T) {
	lc := fxtest.NewLifecycle(t)
	cfg := &Config{URI: "sqlmock://nil"}

	registerLifecycle(lc, nil, mockLogger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	assert.NoError(t, lc.Stop(ctx), "Stop with nil DB should be no-op")
}

func TestRegisterLifecycle_Start_RetryThenOK(t *testing.T) {
	db, mock, cleanup := newSQLXMock(t)
	defer cleanup()

	mock.ExpectPing().WillReturnError(errors.New("ping1"))
	mock.ExpectPing().WillReturnError(errors.New("ping2"))
	mock.ExpectPing() // success
	mock.ExpectClose()

	lc := fxtest.NewLifecycle(t)
	cfg := &Config{URI: "sqlmock://retry-ok"}
	registerLifecycle(lc, db, mockLogger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, lc.Start(ctx), "lc.Start should succeed after retries")
	require.NoError(t, lc.Stop(ctx), "lc.Stop should succeed")

	require.NoError(t, mock.ExpectationsWereMet(), "sqlmock expectations")
}

// Guard compile: confirm methods exist (PingContext/Close)
func Test_sqlxDB_MethodsCompile(t *testing.T) {
	var _ interface {
		PingContext(context.Context) error
		Close() error
	} = (*sqlx.DB)(nil)

	var _ interface {
		PingContext(context.Context) error
		Close() error
	} = (*sql.DB)(nil)
}

func TestRegisterLifecycle_Stop_CloseError_Propagates(t *testing.T) {
	db, mock, cleanup := newSQLXMock(t)
	defer cleanup()

	mock.ExpectPing()
	mock.ExpectClose().WillReturnError(errors.New("close-failed"))

	lc := fxtest.NewLifecycle(t)
	cfg := &Config{URI: "sqlmock://close-err"}

	registerLifecycle(lc, db, mockLogger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	require.NoError(t, lc.Start(ctx), "start should succeed (ping OK)")

	err := lc.Stop(ctx)
	require.Error(t, err, "stop should return error when close fails")
	assert.Contains(t, err.Error(), "close-failed")

	require.NoError(t, mock.ExpectationsWereMet())
}
