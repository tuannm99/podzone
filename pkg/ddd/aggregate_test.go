package ddd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAggregateBaseRequiresID(t *testing.T) {
	t.Parallel()

	_, err := NewAggregateBase("", 0)

	require.Error(t, err)
}

func TestAggregateBaseTracksVersionAndEvents(t *testing.T) {
	t.Parallel()

	id, err := ParseID("ord_1")
	require.NoError(t, err)
	aggregate, err := NewAggregateBase(id, 7)
	require.NoError(t, err)

	aggregate.RecordEvent(testDomainEvent{
		eventType:  "CustomerOrderReceived",
		occurredAt: time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC),
	})

	require.Equal(t, id, aggregate.AggregateID())
	require.Equal(t, Version(7), aggregate.AggregateVersion())
	events := aggregate.PullEvents()
	require.Len(t, events, 1)
	require.Equal(t, "CustomerOrderReceived", events[0].EventType())
	require.Empty(t, aggregate.PullEvents())

	aggregate.SetAggregateVersion(8)
	require.Equal(t, Version(8), aggregate.AggregateVersion())
}

func TestAggregateBaseRecordAndApply(t *testing.T) {
	t.Parallel()

	id, err := ParseID("ord_1")
	require.NoError(t, err)
	aggregate, err := NewAggregateBase(id, 0)
	require.NoError(t, err)
	applier := &recordingEventApplier{}
	event := testDomainEvent{
		eventType:  "CustomerOrderCancelled",
		occurredAt: time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC),
	}

	err = aggregate.RecordAndApply(applier, event)

	require.NoError(t, err)
	require.Equal(t, []string{"CustomerOrderCancelled"}, applier.applied)
	require.Len(t, aggregate.PullEvents(), 1)
}

func TestAggregateBaseRecordAndApplyRequiresApplier(t *testing.T) {
	t.Parallel()

	id, err := ParseID("ord_1")
	require.NoError(t, err)
	aggregate, err := NewAggregateBase(id, 0)
	require.NoError(t, err)

	err = aggregate.RecordAndApply(nil, testDomainEvent{
		eventType:  "CustomerOrderCancelled",
		occurredAt: time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC),
	})

	require.ErrorIs(t, err, ErrMissingEventApplier)
	require.Empty(t, aggregate.PullEvents())
}

type recordingEventApplier struct {
	applied []string
}

func (a *recordingEventApplier) Apply(event DomainEvent) error {
	a.applied = append(a.applied, event.EventType())
	return nil
}
