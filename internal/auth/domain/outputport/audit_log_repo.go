package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log entity.AuditLog) error
	ListByActor(ctx context.Context, actorUserID uint, limit int) ([]entity.AuditLog, error)
}
