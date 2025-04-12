package myredis

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/tuannm99/podzone/pkg/logging"
	"go.uber.org/zap"
)

var (
	client       *redis.Client
	once         sync.Once
	ErrInitRedis error
)

type Config struct {
	Addr          string
	Password      string
	DB            int
	Timeout       time.Duration
	RetryAttempts int
}

func defaultConfig() Config {
	return Config{
		Addr:          "localhost:6379",
		Password:      "",
		DB:            0,
		Timeout:       10 * time.Second,
		RetryAttempts: 3,
	}
}

func GetClient() (*redis.Client, error) {
	once.Do(func() {
		logger := logging.GetLogger()

		redisAddr := os.Getenv("REDIS_ADDR")
		redisPassword := os.Getenv("REDIS_PASSWORD")

		config := defaultConfig()
		if redisAddr != "" {
			config.Addr = redisAddr
		}
		if redisPassword != "" {
			config.Password = redisPassword
		}

		logger.Info("Initializing Redis connection", zap.String("addr", config.Addr))

		client = redis.NewClient(&redis.Options{
			Addr:     config.Addr,
			Password: config.Password,
			DB:       config.DB,
		})

		var err error
		for attempt := 1; attempt <= config.RetryAttempts; attempt++ {
			ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
			err = client.Ping(ctx).Err()
			cancel()

			if err == nil {
				logger.Info("Successfully connected to Redis", zap.String("addr", config.Addr))
				break
			}

			if attempt < config.RetryAttempts {
				retryDelay := time.Duration(attempt) * time.Second
				logger.Warn("Redis connection failed, retrying",
					zap.String("addr", config.Addr),
					zap.Int("attempt", attempt),
					zap.Duration("retry_delay", retryDelay),
					zap.Error(err))
				time.Sleep(retryDelay)
			}
		}

		if err != nil {
			logger.Error("Failed to connect to Redis after multiple attempts",
				zap.String("addr", config.Addr),
				zap.Int("attempts", config.RetryAttempts),
				zap.Error(err))
			ErrInitRedis = fmt.Errorf("failed to connect to Redis at %s: %v", config.Addr, err)
		}
	})

	return client, ErrInitRedis
}

func Close() error {
	if client != nil {
		logger := logging.GetLogger()
		logger.Info("Closing Redis connection")
		return client.Close()
	}
	return nil
}
