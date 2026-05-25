package messaging

import (
	"context"
	"errors"
	"strings"
	"time"
)

type InboxDecision string

const (
	InboxDecisionAcquired   InboxDecision = "acquired"
	InboxDecisionDuplicate  InboxDecision = "duplicate"
	InboxDecisionInProgress InboxDecision = "in_progress"
)

var ErrInboxConsumerNameRequired = errors.New("messaging: inbox consumer name is required")

func IdempotentConsumer(store InboxStore, consumerName string, now func() time.Time) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg Envelope) error {
			if store == nil || strings.TrimSpace(msg.ID) == "" {
				return next.Handle(ctx, msg)
			}

			name := strings.TrimSpace(consumerName)
			if name == "" {
				return ErrInboxConsumerNameRequired
			}

			clock := time.Now().UTC
			if now != nil {
				clock = now
			}

			decision, err := store.Begin(ctx, name, msg.ID, clock())
			if err != nil {
				return err
			}
			if decision == InboxDecisionDuplicate || decision == InboxDecisionInProgress {
				return nil
			}

			if err := next.Handle(ctx, msg); err != nil {
				_ = store.Fail(ctx, name, msg.ID, err.Error(), clock())
				return err
			}
			return store.Complete(ctx, name, msg.ID, clock())
		})
	}
}
