package pdmongo

import (
	"context"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
)

func ModuleFor(name string) fx.Option {
	if name == "" {
		name = "default"
	}

	nameParamTag := `name:"pdmongo-` + name + `"`
	configResultTag := `name:"mongo-` + name + `-config"`
	clientResultTag := `name:"mongo-` + name + `"`

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
				NewMongoDbFromConfig,
				fx.ParamTags(configResultTag),
				fx.ResultTags(clientResultTag),
			),
		),
		fx.Invoke(
			fx.Annotate(
				registerLifecycle,
				fx.ParamTags(``, clientResultTag, ``, configResultTag),
			),
		),
	)
}

func registerLifecycle(lc fx.Lifecycle, client *mongo.Client, log pdlog.Logger, cfg *Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := client.Ping(context.Background(), nil)
			log.Error("Mongo ping failed", "error", err, "uri", cfg.URI)
			return err
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing Mongo connection", "uri", cfg.URI)
			if err := client.Disconnect(ctx); err != nil {
				log.Error("Close Mongo failed", "error", err)
				return err
			}
			log.Info("Mongo connection closed")
			return nil
		},
	})
}
