package elasticsearchfx

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func ModuleFor(name string, url string) fx.Option {
	urlName := fmt.Sprintf("%s-es-url", name)
	clientName := fmt.Sprintf("es-%s", name)

	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func() string { return url },
				fx.ResultTags(fmt.Sprintf(`name:"%s"`, urlName)),
			),
			fx.Annotate(
				NewElasticsearchClient,
				fx.ParamTags(``, ``, fmt.Sprintf(`name:"%s"`, urlName)),
				fx.ResultTags(fmt.Sprintf(`name:"%s"`, clientName)),
			),
		),
	)
}

func NewElasticsearchClient(logger *zap.Logger, lc fx.Lifecycle, url string) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("elasticsearch client error: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping elastic client", zap.String("url", url))
			return nil
		},
	})

	logger.Info("Connected to Elasticsearch", zap.String("url", url))
	return client, nil
}
