package order

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestReceiveCustomerOrderQueuesWhenPartnerSelected(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC)
	order, changes, err := ReceiveCustomerOrder(ReceiveCustomerOrderInput{
		ID:           "ord-1",
		StoreID:      "store-1",
		CandidateID:  "cand-1",
		ProductTitle: "Vintage Tee",
		Quantity:     2,
		Total:        "$40.00",
		CustomerName: "Alex POD",
		Partner:      "Fulfill Fast",
		Now:          now,
	})

	require.NoError(t, err)
	require.Len(t, changes, 2)
	require.Equal(t, "Order created for Vintage Tee", changes[0].Message)
	require.Equal(t, "Queued for Fulfill Fast", changes[1].Message)

	snapshot := order.Snapshot()
	require.Equal(t, StatusQueued, snapshot.Status)
	require.Equal(t, "Fulfill Fast", snapshot.Partner)
	require.Equal(t, "unassigned", snapshot.OperatorAssignee)
	require.Equal(t, SettlementStatusPending, snapshot.SettlementStatus)
	require.Equal(t, 2, snapshot.Quantity)
	require.Equal(t, "$40.00", snapshot.Total)

	events := order.PullEvents()
	require.Len(t, events, 2)
	require.Equal(t, "CustomerOrderReceived", events[0].EventType())
	require.Equal(t, "CustomerOrderQueued", events[1].EventType())
	require.Empty(t, order.PullEvents())
}

func TestReceiveCustomerOrderBlocksWhenPartnerMissing(t *testing.T) {
	t.Parallel()

	order, changes, err := ReceiveCustomerOrder(ReceiveCustomerOrderInput{
		ID:                 "ord-1",
		StoreID:            "store-1",
		CandidateID:        "cand-1",
		ProductTitle:       "Poster",
		Quantity:           0,
		Total:              "$8.00",
		RoutingBlockCode:   "negative_margin",
		RoutingBlockReason: "all eligible partners have negative expected margin",
		Now:                time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	})

	require.NoError(t, err)
	require.Len(t, changes, 2)
	require.Contains(t, changes[1].Message, "Routing blocked")

	snapshot := order.Snapshot()
	require.Equal(t, StatusRoutingBlocked, snapshot.Status)
	require.Empty(t, snapshot.Partner)
	require.Equal(t, 1, snapshot.Quantity)
	require.Equal(t, "Sample customer", snapshot.CustomerName)
	require.Equal(t, "negative_margin", snapshot.RoutingBlockCode)

	events := order.PullEvents()
	require.Len(t, events, 2)
	require.Equal(t, "CustomerOrderReceived", events[0].EventType())
	require.Equal(t, "CustomerOrderRoutingBlocked", events[1].EventType())
}

func TestRehydrateCustomerOrderDoesNotEmitEvents(t *testing.T) {
	t.Parallel()

	order, err := RehydrateCustomerOrder(CustomerOrderSnapshot{
		ID:       "ord-1",
		StoreID:  "store-1",
		Status:   StatusQueued,
		Quantity: 1,
	})

	require.NoError(t, err)
	require.Empty(t, order.PullEvents())
}

func TestRouteManuallyClearsRoutingBlock(t *testing.T) {
	t.Parallel()

	order, _, err := ReceiveCustomerOrder(ReceiveCustomerOrderInput{
		ID:                 "ord-1",
		StoreID:            "store-1",
		CandidateID:        "cand-1",
		ProductTitle:       "Poster",
		Quantity:           1,
		RoutingBlockCode:   "negative_margin",
		RoutingBlockReason: "all eligible partners have negative expected margin",
		Now:                time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	order.PullEvents()

	change, err := order.RouteManually("Fulfill Fast", time.Date(2026, 6, 4, 11, 0, 0, 0, time.UTC))

	require.NoError(t, err)
	require.Equal(t, "Routing unblocked: manually rerouted to Fulfill Fast", change.Message)
	snapshot := order.Snapshot()
	require.Equal(t, StatusQueued, snapshot.Status)
	require.Equal(t, "Fulfill Fast", snapshot.Partner)
	require.Empty(t, snapshot.RoutingBlockCode)
	require.Empty(t, snapshot.RoutingBlockReason)

	events := order.PullEvents()
	require.Len(t, events, 1)
	require.Equal(t, "CustomerOrderRoutingResolved", events[0].EventType())
}

func TestRouteManuallyRequiresBlockedOrder(t *testing.T) {
	t.Parallel()

	order, _, err := ReceiveCustomerOrder(ReceiveCustomerOrderInput{
		ID:           "ord-1",
		StoreID:      "store-1",
		CandidateID:  "cand-1",
		ProductTitle: "Vintage Tee",
		Quantity:     1,
		Partner:      "Fulfill Fast",
		Now:          time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC),
	})
	require.NoError(t, err)

	_, err = order.RouteManually("Print Partner A", time.Date(2026, 6, 4, 11, 0, 0, 0, time.UTC))

	require.Error(t, err)
}
