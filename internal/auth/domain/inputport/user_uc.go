package inputport

import "github.com/tuannm99/podzone/internal/auth/domain/entity"

type UserUsecase interface {
	CreateNewAfterAuthCallback(user entity.User) (*entity.User, error)
	CreateNew(user entity.User) (*entity.User, error)
	UpdateOne(id uint, user entity.User) error
}
