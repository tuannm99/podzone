package dto

import "github.com/tuannm99/podzone/internal/auth/domain/entity"

type GoogleCallbackResp struct {
	JwtToken    string       `json:"jwt_token"`
	RedirectUrl string       `json:"redirect_url"`
	UserInfo    UserInfoResp `json:"user_info"`
}

type UserInfoResp struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResp struct {
	JwtToken string      `json:"jwt_token"`
	UserInfo entity.User `json:"user_info"`
}

type RegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterResp struct {
	JwtToken string      `json:"jwt_token"`
	UserInfo entity.User `json:"user_info"`
}
