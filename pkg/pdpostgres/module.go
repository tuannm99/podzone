package pdpostgres

import (
	"context"
	"fmt"

	"github.com/spf13/viper"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type InstanceConfig struct {
	URI string `mapstructure:"uri"`
}

var Registry = toolkit.NewRegistry[*gorm.DB, InstanceConfig]("real")

func init() {
	Registry.Register("real", PostgresFactory)
	Registry.Register("noop", NoopPostgresFactory)
}

func ModuleFor(name string) fx.Option {
	tag := fmt.Sprintf(`name:"%s"`, "gorm-"+name)

	return fx.Provide(
		fx.Annotate(func(v *viper.Viper, lc fx.Lifecycle, logger pdlog.Logger) (*gorm.DB, error) {
			// expect: postgres.<name>.uri
			sub := v.Sub("postgres")
			if sub == nil {
				return nil, fmt.Errorf("missing config block: postgres")
			}
			sub = sub.Sub(name)
			if sub == nil {
				return nil, fmt.Errorf("missing config block: postgres.%s", name)
			}

			var cfg InstanceConfig
			if err := sub.Unmarshal(&cfg); err != nil {
				return nil, fmt.Errorf("unmarshal postgres.%s failed: %w", name, err)
			}

			factory := Registry.Get()
			db, err := factory(context.Background(), cfg)
			if err != nil {
				return nil, err
			}

			sqlDB, _ := db.DB()
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("Pinging Postgres...").With("dsn", cfg.URI).Send()
					return sqlDB.PingContext(ctx)
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("Closing Postgres connection")
					return sqlDB.Close()
				},
			})

			return db, nil
		}, fx.ResultTags(tag)),
	)
}
