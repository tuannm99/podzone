package redisfx

import (
	"context"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func registerLifecycle(lc fx.Lifecycle, client *redis.Client, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing redis client...")
			return client.Close()
		},
	})
}
