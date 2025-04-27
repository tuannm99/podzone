package usecases

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var _ UserUsecase = (*UserUcImpl)(nil)

type UserUcImpl struct {
	logger              *zap.Logger
	redisClient         *redis.Client
	googleOauthExternal GoogleOauthExternal
}

func NewUserUC(logger *zap.Logger, redis *redis.Client, oauthExternal GoogleOauthExternal) *UserUcImpl {
	return &UserUcImpl{
		logger:              logger,
		redisClient:         redis,
		googleOauthExternal: oauthExternal,
	}
}

// GenerateRedirectUrl implements UserUsecase.
func (uc *UserUcImpl) GenerateRedirectUrl(ctx context.Context) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("error generating state: %w", err)
	}
	state := base64.StdEncoding.EncodeToString(b)

	key := "oauth:google:" + state
	if err := uc.redisClient.Set(ctx, key, time.Now().String(), 10*time.Minute).Err(); err != nil {
		return "", fmt.Errorf("error storing state: %w", err)
	}

	url := uc.googleOauthExternal.GetConfig().AuthCodeURL(state)

	uc.logger.Info("Generated OAuth URL", zap.String("redirect_url", url))
	return url, nil
}
