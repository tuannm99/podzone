package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type Client struct {
	*redis.Client
	logger *zap.Logger
}

type Config struct {
	Addr     string
	Password string
	DB       int
	// Additional Redis options can be added here as needed
}

func DefaultConfig() Config {
	return Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
}

func NewClient(config Config, logger *zap.Logger) (*Client, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %v", config.Addr, err)
	}

	logger.Info("Connected to Redis", zap.String("addr", config.Addr))
	return &Client{
		Client: client,
		logger: logger,
	}, nil
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := c.Client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		c.logger.Error("Failed to set value in Redis",
			zap.String("key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Successfully set value in Redis",
		zap.String("key", key),
		zap.Duration("expiration", expiration))
	return nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	value, err := c.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		c.logger.Debug("Key not found in Redis", zap.String("key", key))
		return "", nil
	} else if err != nil {
		c.logger.Error("Failed to get value from Redis",
			zap.String("key", key),
			zap.Error(err))
		return "", err
	}

	c.logger.Debug("Successfully retrieved value from Redis", zap.String("key", key))
	return value, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	err := c.Client.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete key from Redis",
			zap.String("key", key),
			zap.Error(err))
		return err
	}

	c.logger.Debug("Successfully deleted key from Redis", zap.String("key", key))
	return nil
}

func (c *Client) Close() error {
	err := c.Client.Close()
	if err != nil {
		c.logger.Error("Error closing Redis connection", zap.Error(err))
		return err
	}

	c.logger.Info("Redis connection closed successfully")
	return nil
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.Client.Exists(ctx, key).Result()
	if err != nil {
		c.logger.Error("Failed to check key existence in Redis",
			zap.String("key", key),
			zap.Error(err))
		return false, err
	}

	exists := result > 0
	c.logger.Debug("Checked key existence in Redis",
		zap.String("key", key),
		zap.Bool("exists", exists))
	return exists, nil
}
