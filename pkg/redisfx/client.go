package redisfx

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Injected config + logger using fx.In
type Config struct {
	fx.In

	Settings RedisSetting
	Logger   *zap.Logger
}

// Create Redis client and validate connection
func NewClient(conf Config) (*redis.Client, error) {
	conf.Logger.Info("Initializing Redis connection", zap.String("addr", conf.Settings.Addr))

	client := redis.NewClient(&redis.Options{
		Addr:     conf.Settings.Addr,
		Password: conf.Settings.Password,
		DB:       conf.Settings.DB,
	})

	var err error
	for attempt := 1; attempt <= conf.Settings.RetryAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), conf.Settings.Timeout)
		err = client.Ping(ctx).Err()
		cancel()

		if err == nil {
			conf.Logger.Info("Successfully connected to Redis", zap.String("addr", conf.Settings.Addr))
			break
		}

		if attempt < conf.Settings.RetryAttempts {
			retryDelay := time.Duration(attempt) * time.Second
			conf.Logger.Warn("Redis connection failed, retrying...",
				zap.String("addr", conf.Settings.Addr),
				zap.Int("attempt", attempt),
				zap.Duration("retry_delay", retryDelay),
				zap.Error(err))
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		conf.Logger.Error("Failed to connect to Redis after multiple attempts",
			zap.String("addr", conf.Settings.Addr),
			zap.Int("attempts", conf.Settings.RetryAttempts),
			zap.Error(err))
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", conf.Settings.Addr, err)
	}

	return client, nil
}
