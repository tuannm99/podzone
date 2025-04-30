package domain

import (
	"github.com/tuannm99/podzone/services/auth/domain/entity"
	"github.com/tuannm99/podzone/services/auth/domain/inputport"
	"github.com/tuannm99/podzone/services/auth/domain/outputport"
)

var _ inputport.UserUsecase = (*userUC)(nil)

func NewUserUsecase(userRepo outputport.UserRepository) *userUC {
	return &userUC{
		userRepo: userRepo,
	}
}

type userUC struct {
	userRepo outputport.UserRepository
}

func (uc *userUC) CreateNewAfterAuthCallback(user entity.User) (*entity.User, error) {
	return uc.userRepo.CreateByEmailIfNotExisted(user.Email)
}

func (u *userUC) UpdateOne(id string, user entity.User) error {
	panic("unimplemented")
}
