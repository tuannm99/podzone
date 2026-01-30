package pdsql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/testkit"
)

func TestSQLLifecycle_StartStop_WithRealDB(t *testing.T) {
	t.Parallel()

	dbName := fmt.Sprintf("podzone_fx_%d", time.Now().UnixNano())
	uri := testkit.PostgresDSNWithDB(t, dbName)

	k := koanf.New(".")
	k.Set("sql.default.uri", uri)
	k.Set("sql.default.provider", "postgres")
	k.Set("sql.default.should_run_migration", false)

	app := fxtest.New(
		t,
		ModuleFor(""),
		fx.Supply(k),
		fx.Supply(fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger)))),
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, app.Start(ctx))
	require.NoError(t, app.Stop(ctx))
}

func TestModuleFor_DefaultName_Wiring(t *testing.T) {
	t.Parallel()

	uri := testkit.PostgresDSNWithDB(t, fmt.Sprintf("podzone_fxw_%d", time.Now().UnixNano()))

	k := koanf.New(".")
	k.Set("sql.default.uri", uri)
	k.Set("sql.default.provider", "postgres")
	k.Set("sql.default.should_run_migration", false)

	app := fx.New(
		ModuleFor(""),
		fx.Supply(k),
		fx.Supply(fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger)))),
	)

	require.NoError(t, app.Err())
}
