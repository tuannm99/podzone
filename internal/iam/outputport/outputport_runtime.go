package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/iam/entity"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log entity.AuditLog) error
}

type UserDirectory interface {
	GetByIdentity(ctx context.Context, identity string) (*entity.User, error)
	EnsureByEmail(ctx context.Context, email string) (*entity.User, bool, error)
	GetByID(ctx context.Context, userID uint) (*entity.User, error)
}
