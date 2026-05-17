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
	return t.CreateJwtTokenForSession(user, "", "")
}

func (t *tokenUCImpl) CreateJwtTokenForTenant(user entity.User, activeTenantID string) (string, error) {
	return t.CreateJwtTokenForSession(user, activeTenantID, "")
}

func (t *tokenUCImpl) CreateJwtTokenForSession(user entity.User, activeTenantID, sessionID string) (string, error) {
	return t.CreateJwtTokenForScopedSession(user, activeTenantID, sessionID, nil)
}

func (t *tokenUCImpl) CreateJwtTokenForScopedSession(
	user entity.User,
	activeTenantID, sessionID string,
	sessionPolicy []entity.SessionPolicyStatement,
) (string, error) {
	return t.CreateJwtTokenForSessionState(user, entity.Session{
		ID:             sessionID,
		ActiveTenantID: activeTenantID,
		SessionPolicy:  sessionPolicy,
	})
}

func (t *tokenUCImpl) CreateJwtTokenForSessionState(
	user entity.User,
	session entity.Session,
) (string, error) {
	claims := entity.JWTClaims{
		UserID:              user.Id,
		Email:               user.Email,
		Username:            user.Username,
		ActiveTenantID:      session.ActiveTenantID,
		SessionID:           session.ID,
		SessionPolicy:       session.SessionPolicy,
		AssumedRoleID:       session.AssumedRoleID,
		AssumedRoleScope:    session.AssumedRoleScope,
		AssumedRoleName:     session.AssumedRoleName,
		AssumedRoleTenantID: session.AssumedRoleTenantID,
		Key:                 t.cfg.JWTKey,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(t.cfg.JWTSecret)
	return token.SignedString(secret)
}
