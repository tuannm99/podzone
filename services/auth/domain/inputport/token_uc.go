package inputport

import "github.com/tuannm99/podzone/services/auth/domain/entity"

type TokenUsecase interface {
	CreateJwtToken(user entity.User) (string, error)
}
