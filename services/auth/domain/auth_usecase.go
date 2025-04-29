package domain

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	pb "github.com/tuannm99/podzone/pkg/api/proto/auth"
	"github.com/tuannm99/podzone/services/auth/config"
	"github.com/tuannm99/podzone/services/auth/domain/entity"
	"github.com/tuannm99/podzone/services/auth/domain/inputport"
	"github.com/tuannm99/podzone/services/auth/domain/outputport"
)

var _ inputport.AuthUsecase = (*authUC)(nil)

type authUC struct {
	oauthExternal        outputport.GoogleOauthExternal
	oauthStateRepository outputport.OauthStateRepository
	jwtSecret            []byte
	appRedirectURL       string
}

func NewAuthUsecase(
	oauthExternal outputport.GoogleOauthExternal,
	oauthStateRepotory outputport.OauthStateRepository,
	cfg config.AuthConfig,
) *authUC {
	return &authUC{
		oauthExternal:        oauthExternal,
		oauthStateRepository: oauthStateRepotory,
		jwtSecret:            cfg.JWTSecret,
		appRedirectURL:       cfg.AppRedirectURL,
	}
}

func (u *authUC) GenerateOAuthURL(ctx context.Context) (string, error) {
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

func (u *authUC) HandleOAuthCallback(ctx context.Context, code, state string) (*pb.GoogleCallbackResponse, error) {
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

	jwtToken, err := u.createJWT(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT: %w", err)
	}

	redirectURL := fmt.Sprintf("%s?token=%s", u.appRedirectURL, jwtToken)

	return &pb.GoogleCallbackResponse{
		JwtToken:    jwtToken,
		RedirectUrl: redirectURL,
		UserInfo: &pb.UserInfo{
			Id:            userInfo.Sub,
			Email:         userInfo.Email,
			Name:          userInfo.Name,
			GivenName:     userInfo.GivenName,
			FamilyName:    userInfo.FamilyName,
			Picture:       userInfo.Picture,
			EmailVerified: userInfo.EmailVerified,
		},
	}, nil
}

func (u *authUC) VerifyToken(ctx context.Context, tokenString string) (*pb.VerifyTokenResponse, error) {
	if tokenString == "" {
		return &pb.VerifyTokenResponse{
			IsValid: false,
			Error:   "empty token",
		}, nil
	}

	claims := &entity.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return u.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return &pb.VerifyTokenResponse{
			IsValid: false,
			Error:   "invalid token",
		}, nil
	}

	return &pb.VerifyTokenResponse{
		IsValid: true,
		UserInfo: &pb.UserInfo{
			Id:    claims.Sub,
			Email: claims.Email,
			Name:  claims.Name,
		},
	}, nil
}

func (u *authUC) Logout(ctx context.Context) (*pb.LogoutResponse, error) {
	return &pb.LogoutResponse{
		Success:     true,
		RedirectUrl: "/",
	}, nil
}

func (u *authUC) createJWT(userInfo *outputport.GoogleUserInfo) (string, error) {
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
