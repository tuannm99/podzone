package dto

type GoogleCallbackResp struct {
	JwtToken     string
	RedirectUrl  string
	UserInfoResp UserInfoResp
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
