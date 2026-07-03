package outputport

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

type ConnectionStore interface {
	EnsureIndexes(ctx context.Context) error
	Upsert(ctx context.Context, info entity.ConnectionInfo) error
	SoftDelete(ctx context.Context, tenantID string, infraType entity.InfraType, name string) error
	Get(ctx context.Context, tenantID string, infraType entity.InfraType, name string) (*entity.ConnectionInfo, error)
	ListConnections(
		ctx context.Context,
		tenantID string,
		includeDeleted bool,
		query collection.Query,
	) (collection.Page[entity.ConnectionInfo], error)
	ListEvents(
		ctx context.Context,
		tenantID string,
		query collection.Query,
	) (collection.Page[entity.ConnectionEvent], error)
	AppendEvent(ctx context.Context, ev entity.ConnectionEvent) error
	EnqueueOutbox(ctx context.Context, msg entity.OutboxMessage) error
	FindDueOutbox(ctx context.Context, limit int) ([]entity.OutboxMessage, error)
	MarkOutboxDone(ctx context.Context, eventID string) error
	MarkOutboxFailed(ctx context.Context, eventID string, nextRetry time.Time) error
}
