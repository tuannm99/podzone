package outputport

import (
	"github.com/tuannm99/podzone/services/auth/domain/entity"
)

type UserRepository interface {
	GetByID(id string) (*entity.User, error)
	Create(e entity.User) (*entity.User, error)
	Update(e entity.User) error
	UpdateById(id uint, e entity.User) error
	CreateByEmailIfNotExisted(email string) (*entity.User, error)
	GetByUsernameOrEmail(identity string) (*entity.User, error)
}
