package repository

import (
	"errors"

	"github.com/tuannm99/podzone/services/auth/domain/entity"
	"github.com/tuannm99/podzone/services/auth/domain/outputport"
	"github.com/tuannm99/podzone/services/auth/infrastructure/model"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var _ outputport.UserRepository = (*UserRepositoryImpl)(nil)

type UserRepoParams struct {
	fx.In
	Logger *zap.Logger
	DB     *gorm.DB `name:"gorm-auth"`
}

func NewUserRepositoryImpl(p UserRepoParams) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		logger: p.Logger,
		db:     p.DB,
	}
}

type UserRepositoryImpl struct {
	logger *zap.Logger
	db     *gorm.DB `name:"gorm-auth"`
}

// Create implements outputport.UserRepository.
func (u *UserRepositoryImpl) Create(entity entity.User) (*entity.User, error) {
	if err := u.db.Create(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// CreateByEmailIfNotExisted implements outputport.UserRepository.
func (u *UserRepositoryImpl) CreateByEmailIfNotExisted(email string) (*entity.User, error) {
	var user model.User

	result := u.db.Where("email = ?", email).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		user = model.User{
			Email:       email,
			InitialFrom: "google",
		}
		createdUser, err := u.Create(*user.ToEntity())
		if err != nil {
			return nil, err
		}
		return createdUser, nil
	} else if result.Error != nil {
		return nil, result.Error
	}

	return user.ToEntity(), nil
}

// Update implements outputport.UserRepository.
func (u *UserRepositoryImpl) Update(entity entity.User) error {
	panic("unimplemented")
}

// GetByID implements outputport.UserRepository.
func (u *UserRepositoryImpl) GetByID(id string) (*entity.User, error) {
	panic("unimplemented")
}

// UpdateById implements outputport.UserRepository.
func (u *UserRepositoryImpl) UpdateById(id string, entity entity.User) error {
	panic("unimplemented")
}
