package domain

import (
	"github.com/tuannm99/podzone/services/auth/domain/entity"
	"github.com/tuannm99/podzone/services/auth/domain/inputport"
	"github.com/tuannm99/podzone/services/auth/domain/outputport"
)

var _ inputport.UserUsecase = (*userUCImpl)(nil)

func NewUserUsecase(userRepo outputport.UserRepository) *userUCImpl {
	return &userUCImpl{
		userRepo: userRepo,
	}
}

type userUCImpl struct {
	userRepo outputport.UserRepository
}

func (uc *userUCImpl) CreateNew(user entity.User) (*entity.User, error) {
	return uc.userRepo.Create(user)
}

func (uc *userUCImpl) CreateNewAfterAuthCallback(user entity.User) (*entity.User, error) {
	return uc.userRepo.CreateByEmailIfNotExisted(user.Email)
}

func (uc *userUCImpl) UpdateOne(id string, user entity.User) error {
	return uc.userRepo.UpdateById(id, user)
}
