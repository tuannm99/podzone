package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

type UserRepository interface {
	GetByID(id string) (*entity.User, error)
	Create(e entity.User) (*entity.User, error)
	Update(e entity.User) error
	UpdateById(id uint, e entity.User) error
	CreateByEmailIfNotExisted(email string) (*entity.User, error)
	GetByUsernameOrEmail(identity string) (*entity.User, error)
	List(ctx context.Context, query collection.Query) (collection.Page[entity.User], error)
}
