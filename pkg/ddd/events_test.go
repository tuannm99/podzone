package ddd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testDomainEvent struct {
	eventType  string
	occurredAt time.Time
}

func (e testDomainEvent) EventType() string {
	return e.eventType
}

func (e testDomainEvent) OccurredAtTime() time.Time {
	return e.occurredAt
}

func TestEventRecorderIgnoresNilEvents(t *testing.T) {
	t.Parallel()

	var recorder EventRecorder

	recorder.Record(nil)

	require.Empty(t, recorder.Pull())
}

func TestEventRecorderPeekDoesNotDrain(t *testing.T) {
	t.Parallel()

	var recorder EventRecorder
	recorder.Record(testDomainEvent{eventType: "StoreCreated", occurredAt: time.Now().UTC()})

	require.Len(t, recorder.Peek(), 1)
	require.Len(t, recorder.Pull(), 1)
	require.Empty(t, recorder.Pull())
}
