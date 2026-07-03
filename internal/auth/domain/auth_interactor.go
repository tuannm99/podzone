package domain

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"maps"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
)

var _ inputport.AuthUsecase = (*authInteractorImpl)(nil)

func NewAuthUsecase(
	userUC inputport.UserUsecase,
	tokenUC inputport.TokenUsecase,
	oauthExternal outputport.GoogleOauthExternal,
	oauthStateRepotory outputport.OauthStateRepository,
	userRepository outputport.UserRepository,
	sessionRepository outputport.SessionRepository,
	refreshTokenRepository outputport.RefreshTokenRepository,
	tenantAccessChecker outputport.TenantAccessChecker,
	roleAssumer outputport.RoleAssumer,
	accountBootstrapper outputport.AccountBootstrapper,
	cfg config.AuthConfig,
) *authInteractorImpl {
	return &authInteractorImpl{
		jwtSecret:            cfg.JWTSecret,
		jwtKey:               cfg.JWTKey,
		appRedirectURL:       cfg.AppRedirectURL,
		userUC:               userUC,
		tokenUC:              tokenUC,
		oauthExternal:        oauthExternal,
		oauthStateRepository: oauthStateRepotory,
		userRepository:       userRepository,
		sessionRepository:    sessionRepository,
		refreshTokenRepo:     refreshTokenRepository,
		tenantAccessChecker:  tenantAccessChecker,
		roleAssumer:          roleAssumer,
		accountBootstrapper:  accountBootstrapper,
	}
}

type authInteractorImpl struct {
	jwtSecret      string
	jwtKey         string
	appRedirectURL string

	userUC  inputport.UserUsecase
	tokenUC inputport.TokenUsecase

	oauthExternal        outputport.GoogleOauthExternal
	oauthStateRepository outputport.OauthStateRepository
	userRepository       outputport.UserRepository
	sessionRepository    outputport.SessionRepository
	refreshTokenRepo     outputport.RefreshTokenRepository
	tenantAccessChecker  outputport.TenantAccessChecker
	roleAssumer          outputport.RoleAssumer
	accountBootstrapper  outputport.AccountBootstrapper
}

func (u *authInteractorImpl) newSessionAuthResult(
	ctx context.Context,
	user *entity.User,
	tenantID string,
) (*inputport.AuthResult, error) {
	now := time.Now().UTC()
	session := entity.Session{
		ID:             uuid.NewString(),
		UserID:         user.Id,
		ActiveTenantID: tenantID,
		SessionPolicy:  nil,
		Status:         entity.SessionStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
		ExpiresAt:      now.Add(30 * 24 * time.Hour),
	}
	if err := u.sessionRepository.Create(ctx, session); err != nil {
		return nil, err
	}

	refreshToken, refreshEntity, err := u.newRefreshToken(session.ID, now)
	if err != nil {
		return nil, err
	}
	if err := u.refreshTokenRepo.Create(ctx, refreshEntity); err != nil {
		return nil, err
	}

	accessToken, err := u.issueSessionAccessToken(*user, session)
	if err != nil {
		return nil, err
	}
	return &inputport.AuthResult{
		JwtToken:     accessToken,
		RefreshToken: refreshToken,
		UserInfo:     *user,
	}, nil
}

func (u *authInteractorImpl) loadOwnedActiveSession(
	ctx context.Context,
	userID uint,
	accessToken string,
) (*entity.Session, *entity.User, time.Time, error) {
	user, err := u.userRepository.GetByID(fmt.Sprintf("%d", userID))
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	session, err := u.sessionFromAccessToken(accessToken)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	if session.UserID != userID {
		return nil, nil, time.Time{}, entity.ErrSessionRevoked
	}
	now := time.Now().UTC()
	if session.Status != entity.SessionStatusActive || session.RevokedAt != nil || now.After(session.ExpiresAt) {
		return nil, nil, time.Time{}, entity.ErrSessionRevoked
	}
	return session, user, now, nil
}

func (u *authInteractorImpl) issueSessionAccessToken(user entity.User, session entity.Session) (string, error) {
	if session.AssumedRoleID == 0 && session.AssumedRoleName == "" {
		if len(session.SessionPolicy) == 0 {
			return u.tokenUC.CreateJwtTokenForSession(user, session.ActiveTenantID, session.ID)
		}
		return u.tokenUC.CreateJwtTokenForScopedSession(user, session.ActiveTenantID, session.ID, session.SessionPolicy)
	}
	return u.tokenUC.CreateJwtTokenForSessionState(user, session)
}

func (u *authInteractorImpl) newRefreshToken(sessionID string, now time.Time) (string, entity.RefreshToken, error) {
	raw, err := randomToken(48)
	if err != nil {
		return "", entity.RefreshToken{}, err
	}
	refresh := entity.RefreshToken{
		ID:        uuid.NewString(),
		SessionID: sessionID,
		TokenHash: entity.HashToken(raw),
		ExpiresAt: now.Add(30 * 24 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	return raw, refresh, nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func cloneStringMap(items map[string]string) map[string]string {
	if len(items) == 0 {
		return nil
	}
	out := make(map[string]string, len(items))
	maps.Copy(out, items)
	return out
}

func (u *authInteractorImpl) sessionFromAccessToken(raw string) (*entity.Session, error) {
	if raw == "" {
		return nil, entity.ErrSessionNotFound
	}
	claims := &entity.JWTClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(tok *jwt.Token) (any, error) {
		if tok.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(u.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, entity.ErrSessionNotFound
	}
	if u.jwtKey != "" && claims.Key != u.jwtKey {
		return nil, entity.ErrSessionNotFound
	}
	if claims.SessionID == "" {
		return nil, entity.ErrSessionNotFound
	}
	session, err := u.sessionRepository.GetByID(context.Background(), claims.SessionID)
	if err != nil {
		return nil, err
	}
	return session, nil
}
