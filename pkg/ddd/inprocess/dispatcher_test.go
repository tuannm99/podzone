package inprocess

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/ddd"
)

type testEvent struct {
	eventType  string
	occurredAt time.Time
}

func (e testEvent) EventType() string {
	return e.eventType
}

func (e testEvent) OccurredAtTime() time.Time {
	return e.occurredAt
}

func TestEventDispatcherInvokesHandlersSequentially(t *testing.T) {
	t.Parallel()

	handled := make([]string, 0, 2)
	dispatcher := NewEventDispatcher([]ddd.EventHandler{
		ddd.EventHandlerFunc(func(_ context.Context, event ddd.DomainEvent) error {
			handled = append(handled, "first:"+event.EventType())
			return nil
		}),
		ddd.EventHandlerFunc(func(_ context.Context, event ddd.DomainEvent) error {
			handled = append(handled, "second:"+event.EventType())
			return nil
		}),
	})

	err := dispatcher.Dispatch(context.Background(), []ddd.DomainEvent{
		testEvent{eventType: "CustomerOrderReceived", occurredAt: time.Now().UTC()},
	})

	require.NoError(t, err)
	require.Equal(t, []string{"first:CustomerOrderReceived", "second:CustomerOrderReceived"}, handled)
}

func TestNoopEventDispatcherIgnoresEvents(t *testing.T) {
	t.Parallel()

	dispatcher := NewNoopEventDispatcher()

	err := dispatcher.Dispatch(context.Background(), []ddd.DomainEvent{
		testEvent{eventType: "CustomerOrderReceived", occurredAt: time.Now().UTC()},
	})

	require.NoError(t, err)
}
