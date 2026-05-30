package outputport

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/entity"
)

type ConnectionStore interface {
	EnsureIndexes(ctx context.Context) error
	Upsert(info entity.ConnectionInfo) error
	SoftDelete(tenantID string, infraType entity.InfraType, name string) error
	Get(tenantID string, infraType entity.InfraType, name string) (*entity.ConnectionInfo, error)
	ListConnections(
		tenantID string,
		infraType entity.InfraType,
		includeDeleted bool,
		limit, offset int,
	) ([]entity.ConnectionInfo, error)
	ListEvents(
		tenantID string,
		infraType entity.InfraType,
		name string,
		correlationID string,
		limit, offset int,
	) ([]entity.ConnectionEvent, error)
	AppendEvent(ev entity.ConnectionEvent) error
	EnqueueOutbox(msg entity.OutboxMessage) error
	FindDueOutbox(limit int) ([]entity.OutboxMessage, error)
	MarkOutboxDone(eventID string) error
	MarkOutboxFailed(eventID string, nextRetry time.Time) error
}
