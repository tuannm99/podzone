package usecase

type (
	// External entity
	GoogleUserInfo struct {
		Email string
		Name  string
		Sub   string
	}

	// Infrastructure Interface
	UserRepository interface {
		SaveUser() error
	}
	GoogleOauthExternal interface {
		FetchUserInfo(accessToken string) (*GoogleUserInfo, error)
	}
)

type UserAction struct{}

func NewUserUC() *UserAction {
	return &UserAction{}
}
