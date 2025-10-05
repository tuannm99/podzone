package pdsql

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

func ModuleFor(name string) fx.Option {
	if name == "" {
		name = "default"
	}
	nameParamTag := `name:"pdsql-` + name + `"`
	configResultTag := `name:"sql-` + name + `-config"`
	dbResultTag := `name:"sql-` + name + `"`

	return fx.Options(
		fx.Supply(
			fx.Annotate(name, fx.ResultTags(nameParamTag)),
		),
		fx.Provide(
			fx.Annotate(
				GetConfigFromViper,
				fx.ParamTags(nameParamTag, ``),
				fx.ResultTags(configResultTag),
			),
			fx.Annotate(
				NewDbFromConfig,
				fx.ParamTags(configResultTag),
				fx.ResultTags(dbResultTag),
			),
		),
		fx.Invoke(
			fx.Annotate(
				registerLifecycle,
				fx.ParamTags(``, dbResultTag, ``, configResultTag),
			),
		),
	)
}

func registerLifecycle(lc fx.Lifecycle, db *sqlx.DB, log pdlog.Logger, cfg *Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := db.PingContext(ctx)
			if err != nil {
				log.Error("Database ping failed", "error", err, "uri", cfg.URI)
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing DB connection", "uri", cfg.URI)
			if db == nil {
				return nil
			}
			if err := db.Close(); err != nil {
				log.Error("Close DB failed", "error", err)
				return err
			}
			log.Info("DB connection closed")
			return nil
		},
	})
}
