package dto

type GoogleCallbackResp struct {
	JwtToken    string
	RedirectUrl string
	UserInfo    UserInfo
}

type UserInfo struct {
	Id            string
	Email         string
	Name          string
	GivenName     string
	FamilyName    string
	Picture       string
	EmailVerified bool
}
