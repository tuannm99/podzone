package pdkafka

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// Fx wiring ModuleFor
func ModuleFor(name string) fx.Option {
	return fx.Module(
		"pdkafka-"+name,
		fx.Provide(
			fx.Annotate(
				func(v *viper.Viper) (*Config, error) { return NewConfigFromViper(v, name) },
				fx.ResultTags(fmt.Sprintf(`name:"kafka-%s-config"`, name)),
			),
		),
		fx.Provide(
			fx.Annotate(func(lc fx.Lifecycle, cfg *Config) (*Client, error) {
				cli, err := CreateClientFromConfig(cfg)
				if err != nil {
					return nil, err
				}
				lc.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error { return cli.Ping(ctx) },
						OnStop:  func(ctx context.Context) error { return cli.Close() },
					},
				)
				return cli, nil
			}, fx.ParamTags(
				fmt.Sprintf(`name:"kafka-%s-config"`, name)),
				fx.ResultTags(fmt.Sprintf(`name:"kafka-%s"`, name)),
			),
		),
	)
}
