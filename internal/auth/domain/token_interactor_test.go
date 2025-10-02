package domain

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
)

func TestTokenUsecase_CreateJwtToken_Success(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret: "secret",
		JWTKey:    "app-key",
	}
	uc := NewTokenUsecase(cfg)

	user := entity.User{
		Id:       42,
		Email:    "neo@mx.io",
		Username: "neo",
	}

	tokenStr, err := uc.CreateJwtToken(user)
	require.NoError(t, err)
	require.NotEmpty(t, tokenStr)

	var claims entity.JWTClaims
	parsed, err := jwt.ParseWithClaims(tokenStr, &claims, func(tok *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid, "token should be valid")

	assert.Equal(t, uint(42), claims.UserID)
	assert.Equal(t, "neo@mx.io", claims.Email)
	assert.Equal(t, "neo", claims.Username)
	assert.Equal(t, "app-key", claims.Key)

	now := time.Now().Unix()
	assert.Greater(t, claims.ExpiresAt, now)
	assert.LessOrEqual(t, claims.ExpiresAt, now+int64(25*time.Hour.Seconds()))
	assert.LessOrEqual(t, now, claims.IssuedAt)
}

func TestTokenUsecase_CreateJwtToken_DifferentUsers_ProduceDifferentTokens(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret: "secret",
		JWTKey:    "my-key",
	}
	uc := NewTokenUsecase(cfg)

	t1, err1 := uc.CreateJwtToken(entity.User{Id: 1, Email: "a@x.io", Username: "a"})
	t2, err2 := uc.CreateJwtToken(entity.User{Id: 2, Email: "b@x.io", Username: "b"})
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotEmpty(t, t1)
	require.NotEmpty(t, t2)
	assert.NotEqual(t, t1, t2, "tokens for different users should differ")
}
