package auth

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/tuannm99/podzone/pkg/persistents/redis"
)

type RedisStateStore struct {
	client     *redis.Client
	logger     *zap.Logger
	keyPrefix  string
	expiration time.Duration
}

func NewRedisStateStore(redisClient *redis.Client, logger *zap.Logger) *RedisStateStore {
	return &RedisStateStore{
		client:     redisClient,
		logger:     logger,
		keyPrefix:  "oauth_state:",
		expiration: 10 * time.Minute,
	}
}

func (s *RedisStateStore) Add(state string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := s.keyPrefix + state
	if err := s.client.Set(ctx, key, time.Now().String(), s.expiration); err != nil {
		s.logger.Error("Failed to store state in Redis", zap.Error(err))
		return err
	}

	s.logger.Debug("State added to Redis", zap.String("state", state))
	return nil
}

func (s *RedisStateStore) Verify(state string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := s.keyPrefix + state
	result, err := s.client.Get(ctx, key)

	if err != nil || result == "" {
		s.logger.Warn("State not found in Redis", zap.String("state", state))
		return false
	}

	s.client.Delete(ctx, key)

	s.logger.Debug("State verified and removed from Redis",
		zap.String("state", state),
		zap.String("created_at", result))

	return true
}
