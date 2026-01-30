package pdmongo

import (
	"context"
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

func TestMongoLifecycle_StartStop_WithRealDB(t *testing.T) {
	t.Parallel()

	k := koanf.New(".")
	k.Set("mongo.default.uri", testkit.MongoURI(t))
	k.Set("mongo.default.database", "db")
	k.Set("mongo.default.ping_timeout", "2s")
	k.Set("mongo.default.connect_timeout", "2s")

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

	k := koanf.New(".")
	k.Set("mongo.default.uri", testkit.MongoURI(t))
	k.Set("mongo.default.database", "db")
	k.Set("mongo.default.ping_timeout", "2s")
	k.Set("mongo.default.connect_timeout", "2s")

	app := fx.New(
		ModuleFor(""),
		fx.Supply(k),
		fx.Supply(fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger)))),
	)

	require.NoError(t, app.Err())
}
