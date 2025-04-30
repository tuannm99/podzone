package outputport

import (
	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/auth/domain/entity"
)

type UserRepository interface {
	toolkit.Create[entity.User]
	CreateByEmailIfNotExisted(email string) (*entity.User, error)
}
