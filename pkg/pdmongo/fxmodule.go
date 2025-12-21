package pdmongo

import (
	"context"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
)

type MongoClient interface {
	Ping(ctx context.Context) error
	Disconnect(ctx context.Context) error
}

type mongoClientAdapter struct {
	c *mongo.Client
}

func (a mongoClientAdapter) Ping(ctx context.Context) error {
	return a.c.Ping(ctx, nil)
}

func (a mongoClientAdapter) Disconnect(ctx context.Context) error {
	return a.c.Disconnect(ctx)
}

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
			fx.Annotate(
				func(c *mongo.Client) MongoClient { return mongoClientAdapter{c: c} },
				fx.ParamTags(clientResultTag),
				fx.ResultTags(clientResultTag),
			),
		),
		fx.Invoke(
			fx.Annotate(registerLifecycle, fx.ParamTags(``, clientResultTag, ``, configResultTag)),
		),
	)
}

func registerLifecycle(lc fx.Lifecycle, client MongoClient, log pdlog.Logger, cfg *Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := client.Ping(ctx)
			if err != nil {
				log.Error("Mongo ping failed", "error", err, "uri", cfg.URI)
				return err
			}
			log.Info("Mongo ping OK", "uri", cfg.URI)
			return nil
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
