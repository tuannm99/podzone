package kafka

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/pkg/messaging"
)

type Relay struct {
	store     messaging.OutboxStore
	publisher messaging.Publisher
	limit     int
}

func NewRelay(store messaging.OutboxStore, publisher messaging.Publisher, limit int) *Relay {
	if limit <= 0 {
		limit = 100
	}
	return &Relay{
		store:     store,
		publisher: publisher,
		limit:     limit,
	}
}

func (r *Relay) RunOnce(ctx context.Context) error {
	items, err := r.store.ListPending(ctx, r.limit)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return messaging.ErrNoMessages
	}

	now := time.Now().UTC()
	for _, item := range items {
		if err := r.publisher.Publish(ctx, item.Topic, item.MessageKey, item.Envelope); err != nil {
			_ = r.store.MarkFailed(ctx, item.ID, err.Error(), now.Add(time.Minute))
			continue
		}
		if err := r.store.MarkPublished(ctx, []string{item.ID}, now); err != nil {
			return err
		}
	}
	return nil
}
