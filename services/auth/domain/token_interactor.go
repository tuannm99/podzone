package domain

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/tuannm99/podzone/services/auth/config"
	"github.com/tuannm99/podzone/services/auth/domain/entity"
	"github.com/tuannm99/podzone/services/auth/domain/inputport"
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
		Key:      "jwt-key",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(t.cfg.JWTSecret)
}
