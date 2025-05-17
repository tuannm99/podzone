package dto

import "github.com/tuannm99/podzone/services/auth/domain/entity"

type GoogleCallbackResp struct {
	JwtToken    string
	RedirectUrl string
	UserInfo    UserInfoResp
}

type UserInfoResp struct {
	Id            string
	Email         string
	Name          string
	GivenName     string
	FamilyName    string
	Picture       string
	EmailVerified bool
}

type LoginReq struct {
	Username string
	Password string
}

type LoginResp struct {
	JwtToken string
	UserInfo entity.User
}

type RegisterReq struct {
	Username string
	Password string
	Email    string
}

type RegisterResp struct {
	JwtToken string
	UserInfo entity.User
}
