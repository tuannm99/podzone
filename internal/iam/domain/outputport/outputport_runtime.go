package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log entity.AuditLog) error
}

type UserDirectory interface {
	GetByIdentity(ctx context.Context, identity string) (*entity.User, error)
	EnsureByEmail(ctx context.Context, email string) (*entity.User, bool, error)
	GetByID(ctx context.Context, userID uint) (*entity.User, error)
	List(ctx context.Context, query collection.Query) (collection.Page[entity.User], error)
}
