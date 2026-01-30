package pdredis

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

func TestModuleFor_DefaultName_Wiring(t *testing.T) {
	t.Parallel()

	k := koanf.New(".")
	k.Set("redis.default.uri", testkit.RedisURI(t, 0))

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

func TestModuleFor_CustomName_Wiring(t *testing.T) {
	t.Parallel()

	k := koanf.New(".")
	k.Set("redis.test.uri", testkit.RedisURI(t, 0))

	app := fxtest.New(
		t,
		ModuleFor("test"),
		fx.Supply(k),
		fx.Supply(fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger)))),
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, app.Start(ctx))
	require.NoError(t, app.Stop(ctx))
}

func TestRedisClientAdapter_PingClose_NoServer(t *testing.T) {
	t.Parallel()

	c := testkit.RedisClient(t)

	a := redisClientAdapter{c: c}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	require.NoError(t, a.Ping(ctx))
	require.NoError(t, a.Close())
}
