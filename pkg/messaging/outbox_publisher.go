package messaging

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type TransactionalOutboxPublisher struct {
	store OutboxStore
	now   func() time.Time
}

var ErrNilOutboxStore = errors.New("messaging: nil outbox store")

func NewTransactionalOutboxPublisher(store OutboxStore, now func() time.Time) *TransactionalOutboxPublisher {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &TransactionalOutboxPublisher{
		store: store,
		now:   now,
	}
}

func (p *TransactionalOutboxPublisher) Publish(
	ctx context.Context,
	tx Tx,
	topic string,
	key string,
	msg Envelope,
) error {
	if p == nil || p.store == nil {
		return ErrNilOutboxStore
	}
	if err := msg.Validate(); err != nil {
		return err
	}
	now := p.now()
	record := OutboxRecord{
		ID:            msg.ID,
		Topic:         topic,
		MessageKey:    key,
		Envelope:      msg.Clone(),
		Status:        "pending",
		Attempts:      0,
		NextAttemptAt: now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := p.store.Append(ctx, tx, record); err != nil {
		return fmt.Errorf("append outbox record: %w", err)
	}
	return nil
}

func (p *TransactionalOutboxPublisher) PublishBatch(ctx context.Context, tx Tx, items []PublishRequest) error {
	for _, item := range items {
		if err := p.Publish(ctx, tx, item.Topic, item.Key, item.Msg); err != nil {
			return err
		}
	}
	return nil
}
