package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/tuannm99/podzone/services/auth/domain/outputport"
)

var _ outputport.OauthStateRepository = (*OauthStateRepositoryImpl)(nil)

type OauthStateRepoParams struct {
	fx.In
	RedisClient *redis.Client `name:"redis-auth"`
	Logger      *zap.Logger
}

func NewOauthStateRepositoryImpl(p OauthStateRepoParams) *OauthStateRepositoryImpl {
	return &OauthStateRepositoryImpl{
		redisClient: p.RedisClient,
		logger:      p.Logger,
	}
}

type OauthStateRepositoryImpl struct {
	redisClient *redis.Client
	logger      *zap.Logger
}

func (o *OauthStateRepositoryImpl) Del(key string) error {
	_, err := o.redisClient.Del(context.Background(), key).Result()
	if err != nil {
		o.logger.Info("del state error", zap.Error(err))
	}
	return nil
}

func (o *OauthStateRepositoryImpl) Get(key string) (string, error) {
	data, err := o.redisClient.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", errors.New("invalid state: not found")
	} else if err != nil {
		return "", fmt.Errorf("failed to get state: %w", err)
	}
	o.logger.Debug("get state data success", zap.String("data", data))
	return data, nil
}

func (o *OauthStateRepositoryImpl) Set(key string, duration time.Duration) error {
	if err := o.redisClient.Set(context.Background(), key, time.Now().String(), 10*time.Minute).Err(); err != nil {
		return fmt.Errorf("error storing state: %w", err)
	}
	return nil
}
