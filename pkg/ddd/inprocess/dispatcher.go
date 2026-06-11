package inprocess

import (
	"context"
	"fmt"

	"github.com/tuannm99/podzone/pkg/ddd"
)

type EventDispatcher struct {
	handlers []ddd.EventHandler
}

var _ ddd.EventDispatcher = (*EventDispatcher)(nil)

func NewEventDispatcher(handlers []ddd.EventHandler) *EventDispatcher {
	return &EventDispatcher{handlers: append([]ddd.EventHandler(nil), handlers...)}
}

func NewNoopEventDispatcher() *EventDispatcher {
	return NewEventDispatcher(nil)
}

// Dispatch invokes handlers sequentially and fails fast. Use an outbox or
// message broker backed dispatcher for integration events that need durable
// delivery across process crashes.
func (d *EventDispatcher) Dispatch(ctx context.Context, domainEvents []ddd.DomainEvent) error {
	for _, event := range domainEvents {
		if event == nil {
			continue
		}
		for _, handler := range d.handlers {
			if err := handler.Handle(ctx, event); err != nil {
				return fmt.Errorf("handle domain event %s: %w", event.EventType(), err)
			}
		}
	}
	return nil
}
