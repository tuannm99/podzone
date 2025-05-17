package domain

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/auth/config"
	"github.com/tuannm99/podzone/services/auth/domain/dto"
	"github.com/tuannm99/podzone/services/auth/domain/entity"
	"github.com/tuannm99/podzone/services/auth/domain/inputport"
	"github.com/tuannm99/podzone/services/auth/domain/outputport"
)

var _ inputport.AuthUsecase = (*authInteractorImpl)(nil)

func NewAuthUsecase(
	userUC inputport.UserUsecase,
	tokenUC inputport.TokenUsecase,
	oauthExternal outputport.GoogleOauthExternal,
	oauthStateRepotory outputport.OauthStateRepository,
	userRepository outputport.UserRepository,
	cfg config.AuthConfig,
) *authInteractorImpl {
	return &authInteractorImpl{
		jwtSecret:            cfg.JWTSecret,
		appRedirectURL:       cfg.AppRedirectURL,
		userUC:               userUC,
		tokenUC:              tokenUC,
		oauthExternal:        oauthExternal,
		oauthStateRepository: oauthStateRepotory,
		userRepository:       userRepository,
	}
}

type authInteractorImpl struct {
	jwtSecret      []byte
	appRedirectURL string

	userUC  inputport.UserUsecase
	tokenUC inputport.TokenUsecase

	oauthExternal        outputport.GoogleOauthExternal
	oauthStateRepository outputport.OauthStateRepository
	userRepository       outputport.UserRepository
}

func (u *authInteractorImpl) Login(ctx context.Context, username, password string) (*dto.LoginResp, error) {
	user, err := u.userRepository.GetByUsernameOrEmail(username)
	if err != nil {
		return nil, err
	}

	if user.Password != password {
		return nil, entity.ErrWrongPassword
	}

	token, err := u.tokenUC.CreateJwtToken(*user)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResp{
		JwtToken: token,
		UserInfo: *user,
	}, nil
}

func (u *authInteractorImpl) Register(ctx context.Context, req dto.RegisterReq) (*dto.RegisterResp, error) {
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

	err = u.userRepository.UpdateById(user.ID, entity.User{InitialFrom: "podzone"})
	if err != nil {
		return nil, err
	}

	token, err := u.tokenUC.CreateJwtToken(*user)
	if err != nil {
		return nil, err
	}

	return &dto.RegisterResp{
		JwtToken: token,
		UserInfo: *user,
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
	usr, err := u.userUC.CreateNewAfterAuthCallback(userEntityMapped)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	jwtToken, err := u.tokenUC.CreateJwtToken(*usr)
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
