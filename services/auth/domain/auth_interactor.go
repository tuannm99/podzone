package domain

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/auth/config"
	"github.com/tuannm99/podzone/services/auth/domain/dto"
	"github.com/tuannm99/podzone/services/auth/domain/entity"
	"github.com/tuannm99/podzone/services/auth/domain/inputport"
	"github.com/tuannm99/podzone/services/auth/domain/outputport"
)

var _ inputport.AuthUsecase = (*authInteractorImpl)(nil)

type authInteractorImpl struct {
	jwtSecret      []byte
	appRedirectURL string

	userUC inputport.UserUsecase
	// tokenUC inputport.TokenUsecase

	oauthExternal        outputport.GoogleOauthExternal
	oauthStateRepository outputport.OauthStateRepository
}

func NewAuthUsecase(
	userUC inputport.UserUsecase,
	// tokenUC inputport.TokenUsecase,
	oauthExternal outputport.GoogleOauthExternal,
	oauthStateRepotory outputport.OauthStateRepository,
	cfg config.AuthConfig,
) *authInteractorImpl {
	return &authInteractorImpl{
		jwtSecret:      cfg.JWTSecret,
		appRedirectURL: cfg.AppRedirectURL,
		userUC:         userUC,
		// tokenUC:              tokenUC,
		oauthExternal:        oauthExternal,
		oauthStateRepository: oauthStateRepotory,
	}
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
) (*dto.GoogleCallbackResp, error) {
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

	userEntityMapped := *toolkit.MapStruct[outputport.GoogleUserInfo, entity.User](*userInfo)
	if _, err := u.userUC.CreateNewAfterAuthCallback(userEntityMapped); err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	jwtToken, err := u.createJWT(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT: %w", err)
	}

	redirectURL := fmt.Sprintf("%s?token=%s", u.appRedirectURL, jwtToken)

	userInfoResp := toolkit.MapStruct[outputport.GoogleUserInfo, dto.UserInfoResp](*userInfo)

	return &dto.GoogleCallbackResp{
		JwtToken:    jwtToken,
		RedirectUrl: redirectURL,
		UserInfo:    *userInfoResp,
	}, nil
}

func (u *authInteractorImpl) Logout(ctx context.Context) (string, error) {
	return "/", nil
}

func (u *authInteractorImpl) createJWT(userInfo *outputport.GoogleUserInfo) (string, error) {
	claims := entity.JWTClaims{
		Email: userInfo.Email,
		Name:  userInfo.Name,
		Sub:   userInfo.Sub,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(u.jwtSecret)
}
