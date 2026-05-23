package domain

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

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
