package outputport

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/pkg/messaging"
)

type OutboxRepository interface {
	Append(ctx context.Context, tx messaging.Tx, record messaging.OutboxRecord) error
	ListPending(ctx context.Context, limit int) ([]messaging.OutboxRecord, error)
	MarkPublished(ctx context.Context, ids []string, publishedAt time.Time) error
	MarkFailed(ctx context.Context, id string, errText string, nextAttemptAt time.Time) error
}
