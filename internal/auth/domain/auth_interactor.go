package domain

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/pkg/toolkit"
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
	tenantAccessChecker TenantAccessChecker,
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
	tenantAccessChecker  TenantAccessChecker
}

func (u *authInteractorImpl) Login(ctx context.Context, username, password string) (*inputport.AuthResult, error) {
	user, err := u.userRepository.GetByUsernameOrEmail(username)
	if err != nil {
		return nil, err
	}

	err = entity.CheckPassword(user.Password, password)
	if err != nil {
		return nil, err
	}

	return u.newSessionAuthResult(ctx, user, "")
}

func (u *authInteractorImpl) Register(ctx context.Context, req inputport.RegisterCmd) (*inputport.AuthResult, error) {
	user, err := u.userRepository.Create(
		entity.User{
			Username: req.Username,
			Password: req.Password,
			Email:    req.Email,
		},
	)
	if err != nil {
		return nil, err
	}

	err = u.userRepository.UpdateById(user.Id, entity.User{InitialFrom: "podzone"})
	if err != nil {
		return nil, err
	}

	return u.newSessionAuthResult(ctx, user, "")
}

func (u *authInteractorImpl) SwitchActiveTenant(
	ctx context.Context,
	userID uint,
	tenantID, accessToken string,
) (*inputport.AuthResult, error) {
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}

	if err := u.tenantAccessChecker.EnsureActiveMembership(ctx, tenantID, userID); err != nil {
		return nil, err
	}

	user, err := u.userRepository.GetByID(fmt.Sprintf("%d", userID))
	if err != nil {
		return nil, err
	}

	session, err := u.sessionFromAccessToken(accessToken)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, entity.ErrSessionRevoked
	}
	now := time.Now().UTC()
	if session.Status != entity.SessionStatusActive || session.RevokedAt != nil || now.After(session.ExpiresAt) {
		return nil, entity.ErrSessionRevoked
	}
	if err := u.sessionRepository.UpdateActiveTenant(ctx, session.ID, tenantID, now); err != nil {
		return nil, err
	}

	token, err := u.tokenUC.CreateJwtTokenForSession(*user, tenantID, session.ID)
	if err != nil {
		return nil, err
	}

	return &inputport.AuthResult{
		JwtToken: token,
		UserInfo: *user,
	}, nil
}

func (u *authInteractorImpl) RefreshAccessToken(
	ctx context.Context,
	refreshToken string,
) (*inputport.AuthResult, error) {
	hashed := entity.HashToken(refreshToken)
	stored, err := u.refreshTokenRepo.GetByTokenHash(ctx, hashed)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if stored.RevokedAt != nil {
		return nil, entity.ErrRefreshTokenInvalid
	}
	if now.After(stored.ExpiresAt) {
		return nil, entity.ErrRefreshTokenExpired
	}

	session, err := u.sessionRepository.GetByID(ctx, stored.SessionID)
	if err != nil {
		return nil, err
	}
	if session.Status != entity.SessionStatusActive || session.RevokedAt != nil || now.After(session.ExpiresAt) {
		return nil, entity.ErrSessionRevoked
	}

	user, err := u.userRepository.GetByID(fmt.Sprintf("%d", session.UserID))
	if err != nil {
		return nil, err
	}

	newRefreshToken, refreshEntity, err := u.newRefreshToken(session.ID, now)
	if err != nil {
		return nil, err
	}
	if err := u.refreshTokenRepo.Revoke(ctx, stored.ID, now, &refreshEntity.ID); err != nil {
		return nil, err
	}
	if err := u.refreshTokenRepo.Create(ctx, refreshEntity); err != nil {
		return nil, err
	}
	accessToken, err := u.tokenUC.CreateJwtTokenForSession(*user, session.ActiveTenantID, session.ID)
	if err != nil {
		return nil, err
	}
	return &inputport.AuthResult{
		JwtToken:     accessToken,
		RefreshToken: newRefreshToken,
		UserInfo:     *user,
	}, nil
}

func (u *authInteractorImpl) GenerateOAuthURL(ctx context.Context) (string, error) {
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", fmt.Errorf("error generating state: %w", err)
	}
	state := base64.StdEncoding.EncodeToString(stateBytes)

	key := "oauth:google:" + state
	if err := u.oauthStateRepository.Set(key, 10*time.Minute); err != nil {
		return "", err
	}

	url := u.oauthExternal.GetConfig().AuthCodeURL(state)
	return url, nil
}

func (u *authInteractorImpl) HandleOAuthCallback(
	ctx context.Context,
	code, state string,
) (*inputport.GoogleCallbackResult, error) {
	key := "oauth:google:" + state
	if _, err := u.oauthStateRepository.Get(key); err != nil {
		return nil, err
	}
	_ = u.oauthStateRepository.Del(key)

	token, err := u.oauthExternal.GetConfig().Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	userInfo, err := u.oauthExternal.FetchUserInfo(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}

	userEntityMapped, err := toolkit.MapStruct[outputport.GoogleUserInfo, entity.User](*userInfo)
	if err != nil {
		return nil, err
	}
	usr, err := u.userUC.CreateNewAfterAuthCallback(*userEntityMapped)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}
	authResult, err := u.newSessionAuthResult(ctx, usr, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create auth session: %w", err)
	}
	exchangeCode, err := randomToken(24)
	if err != nil {
		return nil, fmt.Errorf("failed to create exchange code: %w", err)
	}
	payload, err := json.Marshal(authResult)
	if err != nil {
		return nil, fmt.Errorf("failed to encode auth result: %w", err)
	}
	exchangeKey := "oauth:google:exchange:" + exchangeCode
	if err := u.oauthStateRepository.SetValue(exchangeKey, string(payload), 2*time.Minute); err != nil {
		return nil, fmt.Errorf("failed to persist exchange code: %w", err)
	}
	redirectURL := fmt.Sprintf("%s?exchange_code=%s", u.appRedirectURL, exchangeCode)

	userInfoResp, err := toolkit.MapStruct[outputport.GoogleUserInfo, inputport.GoogleUserInfo](*userInfo)
	if err != nil {
		return nil, err
	}

	return &inputport.GoogleCallbackResult{
		ExchangeCode: exchangeCode,
		RedirectUrl:  redirectURL,
		UserInfo:     *userInfoResp,
	}, nil
}

func (u *authInteractorImpl) ExchangeOAuthLogin(
	ctx context.Context,
	exchangeCode string,
) (*inputport.AuthResult, error) {
	if exchangeCode == "" {
		return nil, entity.ErrRefreshTokenInvalid
	}
	key := "oauth:google:exchange:" + exchangeCode
	raw, err := u.oauthStateRepository.Get(key)
	if err != nil {
		return nil, err
	}
	_ = u.oauthStateRepository.Del(key)
	var result inputport.AuthResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("failed to decode oauth exchange payload: %w", err)
	}
	return &result, nil
}

func (u *authInteractorImpl) Logout(ctx context.Context, accessToken string) (string, error) {
	session, err := u.sessionFromAccessToken(accessToken)
	if err != nil {
		return "/", err
	}
	now := time.Now().UTC()
	if err := u.sessionRepository.Revoke(ctx, session.ID, now); err != nil {
		return "/", err
	}
	if err := u.refreshTokenRepo.RevokeBySession(ctx, session.ID, now); err != nil {
		return "/", err
	}
	return "/", nil
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

	accessToken, err := u.tokenUC.CreateJwtTokenForSession(*user, tenantID, session.ID)
	if err != nil {
		return nil, err
	}
	return &inputport.AuthResult{
		JwtToken:     accessToken,
		RefreshToken: refreshToken,
		UserInfo:     *user,
	}, nil
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

func (u *authInteractorImpl) sessionFromAccessToken(raw string) (*entity.Session, error) {
	if raw == "" {
		return nil, entity.ErrSessionNotFound
	}
	claims := &entity.JWTClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(tok *jwt.Token) (interface{}, error) {
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
