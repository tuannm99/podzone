package pdredis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

func ModuleFor(name string) fx.Option {
	if name == "" {
		name = "default"
	}
	nameTag := `name:"` + name + `"`
	resultTag := `name:"redis-` + name + `"`

	return fx.Options(
		fx.Supply(fx.Annotate(name, fx.ResultTags(nameTag))),
		fx.Provide(
			fx.Annotate(GetConfigFromViper, fx.ParamTags(nameTag)),     // -> *pdredis.Config
			fx.Annotate(NewClientFromConfig, fx.ResultTags(resultTag)), // -> *redis.Client[name="redis-<name>"]
		),
		fx.Invoke(
			fx.Annotate(registerLifecycle, fx.ParamTags("", resultTag, "", "")), // lc, client(tagged), log, cfg
		),
	)
}

func registerLifecycle(lc fx.Lifecycle, client *redis.Client, log pdlog.Logger, cfg *Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Pinging Redis...", "uri", cfg.URI)
			return client.Ping(ctx).Err()
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing Redis client")
			return client.Close()
		},
	})
}
