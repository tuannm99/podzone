package mongofx

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func ModuleFor(name string, uri string) fx.Option {
	uriName := fmt.Sprintf("%s-mongo-uri", name)
	resultName := fmt.Sprintf("mongo-%s", name)

	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func() string { return uri },
				fx.ResultTags(fmt.Sprintf(`name:"%s"`, uriName)),
			),
			fx.Annotate(
				NewMongoClient,
				fx.ParamTags(``, ``, fmt.Sprintf(`name:"%s"`, uriName)),
				fx.ResultTags(fmt.Sprintf(`name:"%s"`, resultName)),
			),
		),
	)
}

func NewMongoClient(lc fx.Lifecycle, logger *zap.Logger, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect failed: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err := client.Ping(pingCtx, nil); err != nil {
				return fmt.Errorf("mongo ping failed: %w", err)
			}

			logger.Info("MongoDB is reachable", zap.String("uri", uri))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing MongoDB connection", zap.String("uri", uri))
			return client.Disconnect(ctx)
		},
	})

	return client, nil
}
