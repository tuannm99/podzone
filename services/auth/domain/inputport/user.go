package inputport

import "github.com/tuannm99/podzone/services/auth/domain/entity"

type UserUsecase interface {
	CreateNewAfterAuthCallback(user entity.User) (*entity.User, error)
	UpdateOne(id string, user entity.User) error
}
