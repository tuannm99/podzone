package fulfillment

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUpdateShipmentMarksInTransitAndEmitsEvent(t *testing.T) {
	t.Parallel()

	order, err := RehydrateFulfillmentOrder(FulfillmentOrderSnapshot{
		OrderID: "ord-1",
		Partner: "Fulfill Fast",
		Status:  StatusAwaitingLabel,
	})
	require.NoError(t, err)

	now := time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC)
	systemChange, noteChange, shouldMarkOrderShipped, err := order.UpdateShipment(ShipmentUpdate{
		Status:         StatusInTransit,
		Carrier:        "DHL",
		TrackingNumber: "TRACK-1",
		Notes:          "Handed off to carrier",
	}, now)

	require.NoError(t, err)
	require.True(t, shouldMarkOrderShipped)
	require.Contains(t, systemChange.Message, "in transit via DHL")
	require.NotNil(t, noteChange)

	snapshot := order.Snapshot()
	require.Equal(t, StatusInTransit, snapshot.Status)
	require.NotNil(t, snapshot.ShippedAt)
	require.Nil(t, snapshot.DeliveredAt)

	events := order.PullEvents()
	require.Len(t, events, 1)
	require.Equal(t, "ShipmentInTransit", events[0].EventType())
	require.Empty(t, order.PullEvents())
}

func TestUpdateShipmentRequiresPartnerBeforeShipping(t *testing.T) {
	t.Parallel()

	order, err := RehydrateFulfillmentOrder(FulfillmentOrderSnapshot{
		OrderID: "ord-1",
		Status:  StatusAwaitingLabel,
	})
	require.NoError(t, err)

	_, _, _, err = order.UpdateShipment(ShipmentUpdate{
		Status:  StatusInTransit,
		Carrier: "DHL",
	}, time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC))

	require.Error(t, err)
	require.Contains(t, err.Error(), "partner")
}

func TestUpdateShipmentRequiresCarrierForTransit(t *testing.T) {
	t.Parallel()

	order, err := RehydrateFulfillmentOrder(FulfillmentOrderSnapshot{
		OrderID: "ord-1",
		Partner: "Fulfill Fast",
		Status:  StatusAwaitingLabel,
	})
	require.NoError(t, err)

	_, _, _, err = order.UpdateShipment(ShipmentUpdate{
		Status: StatusInTransit,
	}, time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC))

	require.Error(t, err)
	require.Contains(t, err.Error(), "carrier")
}

func TestRehydrateFulfillmentOrderDoesNotEmitEvents(t *testing.T) {
	t.Parallel()

	order, err := RehydrateFulfillmentOrder(FulfillmentOrderSnapshot{
		OrderID: "ord-1",
		Partner: "Fulfill Fast",
		Status:  StatusAwaitingLabel,
	})

	require.NoError(t, err)
	require.Empty(t, order.PullEvents())
}
