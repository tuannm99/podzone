package inputport

type TokenUsecase interface {
	CreateJwtToken() (string, error)
}
