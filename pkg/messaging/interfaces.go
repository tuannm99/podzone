package messaging

import (
	"context"
	"time"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, key string, msg Envelope) error
	PublishBatch(ctx context.Context, topic string, msgs []PublishRequest) error
}

type Consumer interface {
	Run(ctx context.Context) error
}

type Handler interface {
	Handle(ctx context.Context, msg Envelope) error
}

type Observer interface {
	// Implementations can add logs, distributed tracing, metrics, or error tracking.
	Observe(ctx context.Context, event DeliveryEvent)
}

type ErrorClassifier interface {
	Classify(ctx context.Context, msg Envelope, err error) FailureClassification
}

type Tx interface{}

type OutboxStore interface {
	Append(ctx context.Context, tx Tx, record OutboxRecord) error
	ListPending(ctx context.Context, limit int) ([]OutboxRecord, error)
	MarkPublished(ctx context.Context, ids []string, publishedAt time.Time) error
	MarkFailed(ctx context.Context, id string, errText string, nextAttemptAt time.Time) error
}

type InboxStore interface {
	Begin(ctx context.Context, consumerName string, messageID string, now time.Time) (InboxDecision, error)
	Complete(ctx context.Context, consumerName string, messageID string, processedAt time.Time) error
	Fail(ctx context.Context, consumerName string, messageID string, errText string, failedAt time.Time) error
}

type SagaStep interface {
	Name() string
	Execute(ctx context.Context, msg Envelope) error
	Compensate(ctx context.Context, msg Envelope) error
}
