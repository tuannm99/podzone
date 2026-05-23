package messaging

import (
	"context"
	"time"
)

type Tx interface{}

type OutboxRecord struct {
	ID            string
	Topic         string
	MessageKey    string
	Envelope      Envelope
	Status        string
	Attempts      int
	NextAttemptAt time.Time
	PublishedAt   *time.Time
	ErrorText     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type OutboxStore interface {
	Append(ctx context.Context, tx Tx, record OutboxRecord) error
	ListPending(ctx context.Context, limit int) ([]OutboxRecord, error)
	MarkPublished(ctx context.Context, ids []string, publishedAt time.Time) error
	MarkFailed(ctx context.Context, id string, errText string, nextAttemptAt time.Time) error
}
