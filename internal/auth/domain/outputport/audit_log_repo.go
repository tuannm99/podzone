package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log entity.AuditLog) error
	ListByActor(
		ctx context.Context,
		actorUserID uint,
		query collection.Query,
	) (collection.Page[entity.AuditLog], error)
}
