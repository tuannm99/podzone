package pdpostgres

import (
	"context"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func ModuleFor(name string) fx.Option {
	if name == "" {
		name = "default"
	}
	nameTag := `name:"pdpostgres-` + name + `"`
	resultTag := `name:"gorm-` + name + `"`

	return fx.Options(
		fx.Supply(
			// provide the name string into the container
			fx.Annotate(name, fx.ResultTags(nameTag)),
		),
		fx.Provide(
			fx.Annotate(GetConfigFromViper, fx.ParamTags(nameTag)), // inject the string[name="<name>"]
			fx.Annotate(NewDbFromConfig, fx.ResultTags(resultTag)), // provides *gorm.DB[name="gorm-<name>"]
		),
		fx.Invoke(
			fx.Annotate(
				registerLifecycle,
				fx.ParamTags(``, resultTag, ``, ``), // tag only *gorm.DB
			),
		),
	)
}

func registerLifecycle(lc fx.Lifecycle, db *gorm.DB, log pdlog.Logger, cfg *Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Pinging db...", "uri", cfg.URI)
			sqlDB, _ := db.DB()
			if err := sqlDB.PingContext(ctx); err != nil {
				log.Error("Database ping failed", "error", err)
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing db connection")
			sqlDB, _ := db.DB()
			return sqlDB.Close()
		},
	})
}
