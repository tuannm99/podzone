package ddd

import (
	"context"
	"time"
)

type DomainEvent interface {
	EventType() string
	OccurredAtTime() time.Time
}

type EventApplier interface {
	Apply(event DomainEvent) error
}

type EventDispatcher interface {
	Dispatch(ctx context.Context, events []DomainEvent) error
}

type EventHandler interface {
	Handle(ctx context.Context, event DomainEvent) error
}

type EventHandlerFunc func(ctx context.Context, event DomainEvent) error

var _ EventHandler = EventHandlerFunc(nil)

func (f EventHandlerFunc) Handle(ctx context.Context, event DomainEvent) error {
	return f(ctx, event)
}

type EventRecorder struct {
	events []DomainEvent
}

func (r *EventRecorder) Record(event DomainEvent) {
	if event == nil {
		return
	}
	r.events = append(r.events, event)
}

func (r *EventRecorder) Pull() []DomainEvent {
	if r == nil || len(r.events) == 0 {
		return nil
	}
	events := r.events
	r.events = nil
	return events
}

func (r *EventRecorder) Peek() []DomainEvent {
	if r == nil || len(r.events) == 0 {
		return nil
	}
	return append([]DomainEvent(nil), r.events...)
}
