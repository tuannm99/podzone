package domain

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
)

var _ inputport.TokenUsecase = (*tokenUCImpl)(nil)

func NewTokenUsecase(cfg config.AuthConfig) *tokenUCImpl {
	return &tokenUCImpl{
		cfg: cfg,
	}
}

type tokenUCImpl struct {
	cfg config.AuthConfig
}

// CreateJwtToken implements inputport.TokenUsecase.
func (t *tokenUCImpl) CreateJwtToken(user entity.User) (string, error) {
	claims := entity.JWTClaims{
		UserID:   user.Id,
		Email:    user.Email,
		Username: user.Username,
		Key:      t.cfg.JWTKey,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(t.cfg.JWTSecret)
	return token.SignedString(secret)
}
