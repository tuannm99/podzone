package inputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
)

type RegisterCmd struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type AuthResult struct {
	JwtToken     string      `json:"jwt_token"`
	RefreshToken string      `json:"refresh_token"`
	UserInfo     entity.User `json:"user_info"`
}

type GoogleCallbackResult struct {
	ExchangeCode string         `json:"exchange_code"`
	RedirectUrl  string         `json:"redirect_url"`
	UserInfo     GoogleUserInfo `json:"user_info"`
}

type GoogleUserInfo struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}

type AuthUsecase interface {
	GenerateOAuthURL(ctx context.Context) (string, error)
	HandleOAuthCallback(ctx context.Context, code, state string) (*GoogleCallbackResult, error)
	ExchangeOAuthLogin(ctx context.Context, exchangeCode string) (*AuthResult, error)
	Login(ctx context.Context, username, password string) (*AuthResult, error)
	Register(ctx context.Context, req RegisterCmd) (*AuthResult, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	SwitchActiveTenant(ctx context.Context, userID uint, tenantID, accessToken string) (*AuthResult, error)
	Logout(ctx context.Context, accessToken string) (string, error)
}
