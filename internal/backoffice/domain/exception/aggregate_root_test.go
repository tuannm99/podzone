package exception

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenOrderExceptionEmitsEvent(t *testing.T) {
	t.Parallel()

	exception, err := RehydrateOrderException(OrderExceptionSnapshot{OrderID: "ord-1"})
	require.NoError(t, err)

	change, err := exception.Open("reprint_request", time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC))

	require.NoError(t, err)
	require.Contains(t, change.Message, "Exception opened")
	require.Equal(t, StatusOpen, exception.Snapshot().Status)

	events := exception.PullEvents()
	require.Len(t, events, 1)
	require.Equal(t, "OrderExceptionOpened", events[0].EventType())
	require.Empty(t, exception.PullEvents())
}

func TestUpdateOrderExceptionStatusEmitsEvent(t *testing.T) {
	t.Parallel()

	exception, err := RehydrateOrderException(OrderExceptionSnapshot{
		OrderID: "ord-1",
		Type:    "reprint_request",
		Status:  StatusOpen,
	})
	require.NoError(t, err)

	change, err := exception.UpdateStatus(StatusEscalated, time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC))

	require.NoError(t, err)
	require.Contains(t, change.Message, "Exception escalated")
	require.Equal(t, StatusEscalated, exception.Snapshot().Status)
	require.Len(t, exception.PullEvents(), 1)
}

func TestResolvedExceptionCannotBeReopenedByStatusUpdate(t *testing.T) {
	t.Parallel()

	exception, err := RehydrateOrderException(OrderExceptionSnapshot{
		OrderID: "ord-1",
		Type:    "reprint_request",
		Status:  StatusResolved,
	})
	require.NoError(t, err)

	_, err = exception.UpdateStatus(StatusOpen, time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC))

	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot be reopened")
}

func TestRehydrateOrderExceptionDoesNotEmitEvents(t *testing.T) {
	t.Parallel()

	exception, err := RehydrateOrderException(OrderExceptionSnapshot{
		OrderID: "ord-1",
		Type:    "reprint_request",
		Status:  StatusOpen,
	})

	require.NoError(t, err)
	require.Empty(t, exception.PullEvents())
}
