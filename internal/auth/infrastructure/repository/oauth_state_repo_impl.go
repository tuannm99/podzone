package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

var _ outputport.OauthStateRepository = (*OauthStateRepositoryImpl)(nil)

type OauthStateRepoParams struct {
	fx.In
	RedisClient redis.Cmdable `name:"redis-auth"`
	Logger      pdlog.Logger
}

func NewOauthStateRepositoryImpl(p OauthStateRepoParams) *OauthStateRepositoryImpl {
	return &OauthStateRepositoryImpl{
		redisClient: p.RedisClient,
		logger:      p.Logger,
	}
}

type OauthStateRepositoryImpl struct {
	redisClient redis.Cmdable
	logger      pdlog.Logger
}

func (o *OauthStateRepositoryImpl) Del(key string) error {
	_, err := o.redisClient.Del(context.Background(), key).Result()
	if err != nil {
		o.logger.Info("del state error", "err", err)
		return err
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
	o.logger.Debug("get state data success", "err", err)
	return data, nil
}

func (o *OauthStateRepositoryImpl) Set(key string, duration time.Duration) error {
	if err := o.redisClient.Set(context.Background(), key, time.Now().String(), duration).Err(); err != nil {
		return fmt.Errorf("error storing state: %w", err)
	}
	return nil
}
