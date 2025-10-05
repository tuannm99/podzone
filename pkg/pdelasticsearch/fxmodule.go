package pdelasticsearch

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
)

func ModuleFor(name string) fx.Option {
	if name == "" {
		name = "default"
	}

	nameParamTag := `name:"pdelasticsearch-` + name + `"`
	configResultTag := `name:"es-` + name + `-config"`
	clientResultTag := `name:"es-` + name + `"`

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
				NewClientFromConfig,
				fx.ParamTags(configResultTag),
				fx.ResultTags(clientResultTag),
			),
		),
		fx.Invoke(
			fx.Annotate(registerLifecycle, fx.ParamTags(``, clientResultTag, ``, configResultTag)),
		),
	)
}

func registerLifecycle(lc fx.Lifecycle, client *elasticsearch.Client, log pdlog.Logger, cfg *Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
			res, err := client.Ping(client.Ping.WithContext(pingCtx))
			cancel()

			if err != nil {
				log.Error("Elasticsearch ping failed", "error", err, "addresses", cfg.Addresses)
				return err
			}

			if err == nil && !res.IsError() {
				res.Body.Close()
				log.Info("Elasticsearch ping OK", "addresses", cfg.Addresses)
				return nil
			}
			if res != nil {
				res.Body.Close()
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Closing Elasticsearch client", "addresses", cfg.Addresses)
			return nil
		},
	})
}
