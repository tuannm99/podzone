package pdsql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

type mockSQLDB struct {
	pingErr  error
	closeErr error
	pingN    int
	closeN   int
}

func (m *mockSQLDB) PingContext(ctx context.Context) error {
	m.pingN++
	return m.pingErr
}

func (m *mockSQLDB) Close() error {
	m.closeN++
	return m.closeErr
}

func TestSQLLifecycle_Table(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		pingErr      error
		closeErr     error
		wantStartErr bool
		wantStopErr  bool
	}{
		{name: "start_ok_stop_ok"},
		{name: "start_fail_ping", pingErr: errors.New("ping failed"), wantStartErr: true},
		{name: "stop_fail_close", closeErr: errors.New("close failed"), wantStopErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := &mockSQLDB{pingErr: tc.pingErr, closeErr: tc.closeErr}
			cfg := &Config{URI: "postgres://unit-test"}

			app := fxtest.New(
				t,
				fx.Supply(
					fx.Annotate(m, fx.As(new(SQLDB))),
					cfg,
					fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger))),
				),
				fx.Invoke(registerLifecycle),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			startErr := app.Start(ctx)
			if tc.wantStartErr {
				require.Error(t, startErr)
				require.GreaterOrEqual(t, m.pingN, 1)
				_ = app.Stop(context.Background())
				return
			}
			require.NoError(t, startErr)
			require.Equal(t, 1, m.pingN)

			stopErr := app.Stop(ctx)
			if tc.wantStopErr {
				require.Error(t, stopErr)
			} else {
				require.NoError(t, stopErr)
			}
			require.Equal(t, 1, m.closeN)
		})
	}
}

func TestModuleFor_DefaultName_Wiring(t *testing.T) {
	t.Parallel()

	k := koanf.New(".")
	k.Set("sql.default.uri", "postgres://u:p@127.0.0.1:5432/testdb?sslmode=disable")
	k.Set("sql.default.provider", "postgres")
	k.Set("sql.default.should_run_migration", false)

	rawDB, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = rawDB.Close() })

	dbx := sqlx.NewDb(rawDB, "sqlmock")

	app := fx.New(
		ModuleFor(""), // => default

		fx.Supply(k),
		fx.Supply(fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger)))),

		// Keep these replaces: they make wiring deterministic and avoid real DB ops.
		fx.Replace(
			fx.Annotate(&Config{
				URI:                "postgres://u:p@127.0.0.1:5432/testdb?sslmode=disable",
				Provider:           "postgres",
				ShouldRunMigration: false,
			}, fx.ResultTags(`name:"sql-default-config"`)),
		),
		fx.Replace(
			fx.Annotate(dbx, fx.ResultTags(`name:"sql-default"`)),
		),
	)

	require.NoError(t, app.Err())
}
