package infrastructure

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

type Params struct {
	fx.In
	Logger *zap.Logger
	DB     *gorm.DB `name:"gorm-auth"`
}

func NewUserRepository(p Params) *UserRepositoryImpl {
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
	panic("unimplemented")
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
		if err := u.db.Create(&user).Error; err != nil {
			return nil, err
		}
	} else if result.Error != nil {
		return nil, result.Error
	}

	return user.ToEntity(), nil
}
