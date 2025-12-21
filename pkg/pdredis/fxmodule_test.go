package pdredis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

func TestModuleFor_DefaultName_Wiring(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set("redis.default.uri", "redis://127.0.0.1:6379/0")

	app := fx.New(
		ModuleFor(""),
		fx.Supply(v),
		fx.Supply(fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger)))),
	)

	require.NoError(t, app.Err())
}

func TestModuleFor_CustomName_Wiring(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set("redis.test.uri", "redis://127.0.0.1:6379/0")

	app := fx.New(
		ModuleFor("test"),
		fx.Supply(v),
		fx.Supply(fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger)))),
	)

	require.NoError(t, app.Err())
}

func TestRedisClientAdapter_PingClose_NoServer(t *testing.T) {
	t.Parallel()

	c := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:1",
		DB:   0,
	})

	a := redisClientAdapter{c: c}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	require.Error(t, a.Ping(ctx))
	require.NoError(t, a.Close())
}
