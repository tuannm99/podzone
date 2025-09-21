package pdmongo

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
)

func ModuleFor(name string) fx.Option {
	tag := fmt.Sprintf(`name:"%s"`, "mongo-"+name)

	return fx.Provide(
		fx.Annotate(func(v *viper.Viper, lc fx.Lifecycle, logger pdlog.Logger) (*mongo.Client, error) {
			// mongo
			sub := v.Sub("mongo")
			if sub == nil {
				return nil, fmt.Errorf("missing config block: mongo")
			}
			// mongo.<name>
			sub = sub.Sub(name)
			if sub == nil {
				return nil, fmt.Errorf("missing config block: mongo.%s", name)
			}

			var cfg InstanceConfig
			if err := sub.Unmarshal(&cfg); err != nil {
				return nil, fmt.Errorf("unmarshal mongo.%s failed: %w", name, err)
			}

			factory := Registry.Get()
			client, err := factory(context.Background(), cfg)
			if err != nil {
				return nil, err
			}

			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					if currentFactoryID == "real" {
						logger.Info("Pinging Mongo...").With("dsn", cfg.URI).Send()
						if err := ping(ctx, client); err != nil {
							return fmt.Errorf("mongo ping failed: %w", err)
						}
						logger.Info("Mongo is reachable").With("dsn", cfg.URI).Send()
					} else {
						logger.Info("Using NoopMongoFactory (skip connect/ping)").With("name", name).Send()
					}
					return nil
				},
				OnStop: func(ctx context.Context) error {
					if currentFactoryID == "real" {
						logger.Info("Closing Mongo client").With("name", name).Send()
						return client.Disconnect(ctx)
					}
					return nil
				},
			})

			return client, nil
		}, fx.ResultTags(tag)),
	)
}
