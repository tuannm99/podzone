package outputport

import (
	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/auth/domain/entity"
)

type UserRepository interface {
	toolkit.GetByID[entity.User]
	toolkit.Create[entity.User]
	toolkit.Update[entity.User]
	toolkit.UpdateById[entity.User]
	CreateByEmailIfNotExisted(email string) (*entity.User, error)
	GetByUsernameOrEmail(identity string) (*entity.User, error)
}
