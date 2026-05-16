package interactor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
)

func TestCreateRoutedOrderSnapshotsCostsAndMargin(t *testing.T) {
	t.Parallel()

	interactor, _, _ := newOrderRoutingTestInteractor(t, map[string]entity.ProductSetupCandidate{
		"cand-1": {
			ID:          "cand-1",
			Title:       "Vintage Tee",
			Partner:     "Print Partner A",
			BaseCost:    "$8.00",
			RetailPrice: "$20.00",
			Status:      entity.ProductSetupCandidateStatusPublishedMock,
		},
	})

	order, err := interactor.CreateRoutedOrder(context.Background(), inputport.CreateRoutedOrderCmd{
		CandidateID:  "cand-1",
		CustomerName: "Alex POD",
		Quantity:     2,
	})
	require.NoError(t, err)
	require.Equal(t, entity.RoutedOrderStatusQueued, order.Status)
	require.Equal(t, entity.RoutedOrderShipmentStatusAwaitingLabel, order.ShipmentStatus)
	require.Equal(t, "unassigned", order.OperatorAssignee)
	require.Equal(t, "$16.00", order.BaseCostSnapshot)
	require.Equal(t, "$16.00", order.FulfillmentCost)
	require.Equal(t, "$40.00", order.Total)
	require.Equal(t, "$24.00", order.RealizedMargin)
	require.Equal(t, entity.RoutedOrderSettlementStatusPending, order.SettlementStatus)
	require.Len(t, order.Timeline, 2)
	require.Len(t, order.ActivityLog, 2)
	require.Equal(t, "system", order.ActivityLog[0].Actor)
	require.NotEmpty(t, order.ActivityLog[0].Details)
}

func TestUpdateOrderSettlementRecalculatesMarginIncludingIssueCost(t *testing.T) {
	t.Parallel()

	interactor, orders, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(entity.RoutedOrder{
		ID:               "ord-1",
		Total:            "$40.00",
		FulfillmentCost:  "$10.00",
		ShippingCost:     "$4.00",
		IssueCost:        "$3.00",
		SettlementStatus: entity.RoutedOrderSettlementStatusPending,
		Timeline:         []string{"created"},
	})

	order, err := interactor.UpdateOrderSettlement(context.Background(), inputport.UpdateOrderSettlementCmd{
		OrderID:          "ord-1",
		FulfillmentCost:  "$12.00",
		ShippingCost:     "$5.50",
		SettlementStatus: entity.RoutedOrderSettlementStatusReconciled,
		Notes:            "Supplier invoice matched",
	})
	require.NoError(t, err)
	require.Equal(t, "$12.00", order.FulfillmentCost)
	require.Equal(t, "$5.50", order.ShippingCost)
	require.Equal(t, "$19.50", order.RealizedMargin)
	require.Equal(t, entity.RoutedOrderSettlementStatusReconciled, order.SettlementStatus)
	require.Equal(t, "Supplier invoice matched", order.SettlementNotes)
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "Settlement")
	got := order.ActivityLog[len(order.ActivityLog)-1]
	require.Equal(t, entity.RoutedOrderActivityTypeSettlementNote, got.Type)
	require.Equal(t, "Supplier invoice matched", got.Message)
	require.Equal(t, "system", got.Actor)
	require.True(t, hasActivityDetail(got.Details, "settlement_status", entity.RoutedOrderSettlementStatusReconciled))
}

func TestUpdateOrderIssueHandlingRequiresActiveIssue(t *testing.T) {
	t.Parallel()

	interactor, orders, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(entity.RoutedOrder{
		ID:               "ord-no-issue",
		Total:            "$40.00",
		FulfillmentCost:  "$10.00",
		ShippingCost:     "$5.00",
		IssueCost:        "$0.00",
		ShipmentStatus:   entity.RoutedOrderShipmentStatusAwaitingLabel,
		SettlementStatus: entity.RoutedOrderSettlementStatusPending,
	})

	_, err := interactor.UpdateOrderIssueHandling(context.Background(), inputport.UpdateOrderIssueHandlingCmd{
		OrderID:         "ord-no-issue",
		IssueCost:       "$6.00",
		IssueResolution: entity.RoutedOrderIssueResolutionReprint,
		Notes:           "Needs reprint",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "active exception or delivery issue")
}

func TestUpdateOrderIssueHandlingRecalculatesMargin(t *testing.T) {
	t.Parallel()

	interactor, orders, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(entity.RoutedOrder{
		ID:               "ord-issue",
		Total:            "$40.00",
		FulfillmentCost:  "$10.00",
		ShippingCost:     "$5.00",
		IssueCost:        "$0.00",
		ExceptionType:    "reprint_request",
		ExceptionStatus:  entity.RoutedOrderExceptionStatusOpen,
		ShipmentStatus:   entity.RoutedOrderShipmentStatusAwaitingLabel,
		SettlementStatus: entity.RoutedOrderSettlementStatusPending,
		Timeline:         []string{"created"},
	})

	order, err := interactor.UpdateOrderIssueHandling(context.Background(), inputport.UpdateOrderIssueHandlingCmd{
		OrderID:         "ord-issue",
		IssueCost:       "$6.00",
		IssueResolution: entity.RoutedOrderIssueResolutionReprint,
		Notes:           "Reprint approved",
	})
	require.NoError(t, err)
	require.Equal(t, "$6.00", order.IssueCost)
	require.Equal(t, entity.RoutedOrderIssueResolutionReprint, order.IssueResolution)
	require.Equal(t, "$19.00", order.RealizedMargin)
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "Issue handling")
	got := order.ActivityLog[len(order.ActivityLog)-1]
	require.Equal(t, entity.RoutedOrderActivityTypeIssueNote, got.Type)
	require.Equal(t, "Reprint approved", got.Message)
	require.Equal(t, "system", got.Actor)
	require.True(t, hasActivityDetail(got.Details, "issue_resolution", entity.RoutedOrderIssueResolutionReprint))
}

func TestUpdateOrderQueueControlNormalizesAssigneeAndPersistsSLA(t *testing.T) {
	t.Parallel()

	shipmentSLA := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	issueSLA := shipmentSLA.Add(4 * time.Hour)
	interactor, orders, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(entity.RoutedOrder{
		ID:       "ord-queue",
		Timeline: []string{"created"},
	})

	order, err := interactor.UpdateOrderQueueControl(context.Background(), inputport.UpdateOrderQueueControlCmd{
		OrderID:          "ord-queue",
		OperatorAssignee: "  ",
		ShipmentSlaDueAt: &shipmentSLA,
		IssueSlaDueAt:    &issueSLA,
	})
	require.NoError(t, err)
	require.Equal(t, "unassigned", order.OperatorAssignee)
	require.NotNil(t, order.ShipmentSlaDueAt)
	require.True(t, order.ShipmentSlaDueAt.Equal(shipmentSLA))
	require.NotNil(t, order.IssueSlaDueAt)
	require.True(t, order.IssueSlaDueAt.Equal(issueSLA))
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "Queue ownership updated")
}

func TestBulkUpdateRoutedOrdersUpdatesSelectedOrders(t *testing.T) {
	t.Parallel()

	shipmentSLA := time.Date(2026, 5, 15, 18, 30, 0, 0, time.UTC)
	assignee := "ops.lead"
	settlementStatus := entity.RoutedOrderSettlementStatusPaid
	interactor, orders, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(entity.RoutedOrder{
		ID:               "ord-1",
		OperatorAssignee: "unassigned",
		SettlementStatus: entity.RoutedOrderSettlementStatusPending,
		Timeline:         []string{"created"},
	})
	orders.mustSeed(entity.RoutedOrder{
		ID:               "ord-2",
		OperatorAssignee: "unassigned",
		SettlementStatus: entity.RoutedOrderSettlementStatusPending,
		Timeline:         []string{"created"},
	})

	updated, err := interactor.BulkUpdateRoutedOrders(context.Background(), inputport.BulkUpdateRoutedOrdersCmd{
		OrderIDs:         []string{"ord-1", " ord-2 "},
		OperatorAssignee: &assignee,
		ShipmentSlaDueAt: &shipmentSLA,
		SettlementStatus: &settlementStatus,
	})
	require.NoError(t, err)
	require.Len(t, updated, 2)
	for _, order := range updated {
		require.Equal(t, "ops.lead", order.OperatorAssignee)
		require.NotNil(t, order.ShipmentSlaDueAt)
		require.True(t, order.ShipmentSlaDueAt.Equal(shipmentSLA))
		require.Equal(t, entity.RoutedOrderSettlementStatusPaid, order.SettlementStatus)
		require.Contains(t, order.Timeline[len(order.Timeline)-1], "Bulk queue update applied")
	}
}

func TestListRoutedOrderActivitiesFiltersAndSorts(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	interactor, orders, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(entity.RoutedOrder{
		ID:               "ord-1",
		ProductTitle:     "Vintage Tee",
		OperatorAssignee: "ops.lead",
		ActivityLog: []entity.RoutedOrderActivity{
			{Type: entity.RoutedOrderActivityTypeSystem, Actor: "system", Message: "Queued", CreatedAt: now.Add(-2 * time.Hour)},
			{Type: entity.RoutedOrderActivityTypeShipmentNote, Actor: "user:12", Message: "Carrier assigned", CreatedAt: now.Add(-1 * time.Hour)},
		},
	})
	orders.mustSeed(entity.RoutedOrder{
		ID:               "ord-2",
		ProductTitle:     "Poster",
		OperatorAssignee: "ops.a",
		ActivityLog: []entity.RoutedOrderActivity{
			{Type: entity.RoutedOrderActivityTypeShipmentNote, Actor: "user:15", Message: "Packed", CreatedAt: now.Add(-30 * time.Minute)},
		},
	})

	firstPage, err := interactor.ListRoutedOrderActivities(context.Background(), inputport.ListRoutedOrderActivitiesQuery{
		ActivityType:  entity.RoutedOrderActivityTypeShipmentNote,
		ActorContains: "user:",
		Since:         ptrTime(now.Add(-90 * time.Minute)),
		Limit:         1,
		IncludeSystem: false,
	})
	require.NoError(t, err)
	require.Equal(t, 2, firstPage.Total)
	require.Len(t, firstPage.Entries, 1)
	require.Equal(t, "ord-2", firstPage.Entries[0].OrderID)
	require.NotNil(t, firstPage.NextCursor)
	require.NotEmpty(t, *firstPage.NextCursor)

	secondPage, err := interactor.ListRoutedOrderActivities(context.Background(), inputport.ListRoutedOrderActivitiesQuery{
		ActivityType:  entity.RoutedOrderActivityTypeShipmentNote,
		ActorContains: "user:",
		Since:         ptrTime(now.Add(-90 * time.Minute)),
		Limit:         1,
		After:         *firstPage.NextCursor,
		IncludeSystem: false,
	})
	require.NoError(t, err)
	require.Len(t, secondPage.Entries, 1)
	require.Equal(t, "ord-1", secondPage.Entries[0].OrderID)
	require.Nil(t, secondPage.NextCursor)
}

func TestAdvanceRoutedOrderBlocksWhenExceptionActive(t *testing.T) {
	t.Parallel()

	interactor, orders, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(entity.RoutedOrder{
		ID:              "ord-blocked",
		Status:          entity.RoutedOrderStatusQueued,
		ExceptionType:   "partner_delay",
		ExceptionStatus: entity.RoutedOrderExceptionStatusOpen,
	})

	_, err := interactor.AdvanceRoutedOrder(context.Background(), "ord-blocked")
	require.Error(t, err)
	require.Contains(t, err.Error(), "resolve the active exception")
}

func TestAdvanceRoutedOrderTransitionsQueuedToProduction(t *testing.T) {
	t.Parallel()

	interactor, orders, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(entity.RoutedOrder{
		ID:       "ord-advance",
		Status:   entity.RoutedOrderStatusQueued,
		Partner:  "Print Partner A",
		Timeline: []string{"created"},
	})

	order, err := interactor.AdvanceRoutedOrder(context.Background(), "ord-advance")
	require.NoError(t, err)
	require.Equal(t, entity.RoutedOrderStatusInProduction, order.Status)
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "POD production")
}

func TestOpenAndResolveOrderException(t *testing.T) {
	t.Parallel()

	interactor, orders, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(entity.RoutedOrder{
		ID:       "ord-exception",
		Timeline: []string{"created"},
	})

	opened, err := interactor.OpenOrderException(context.Background(), inputport.OpenOrderExceptionCmd{
		OrderID:       "ord-exception",
		ExceptionType: "reprint_request",
	})
	require.NoError(t, err)
	require.Equal(t, "reprint_request", opened.ExceptionType)
	require.Equal(t, entity.RoutedOrderExceptionStatusOpen, opened.ExceptionStatus)
	require.Contains(t, opened.Timeline[len(opened.Timeline)-1], "Exception opened")

	resolved, err := interactor.UpdateOrderExceptionStatus(context.Background(), inputport.UpdateOrderExceptionStatusCmd{
		OrderID: "ord-exception",
		Status:  entity.RoutedOrderExceptionStatusResolved,
	})
	require.NoError(t, err)
	require.Equal(t, entity.RoutedOrderExceptionStatusResolved, resolved.ExceptionStatus)
	require.Contains(t, resolved.Timeline[len(resolved.Timeline)-1], "Exception resolved")
}

func TestUpdateOrderShipmentMarksInTransitAndAppendsTracking(t *testing.T) {
	t.Parallel()

	interactor, orders, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(entity.RoutedOrder{
		ID:             "ord-ship",
		Status:         entity.RoutedOrderStatusQueued,
		ShipmentStatus: entity.RoutedOrderShipmentStatusAwaitingLabel,
		Partner:        "Print Partner A",
		Timeline:       []string{"created"},
	})

	order, err := interactor.UpdateOrderShipment(context.Background(), inputport.UpdateOrderShipmentCmd{
		OrderID:        "ord-ship",
		ShipmentStatus: entity.RoutedOrderShipmentStatusInTransit,
		Carrier:        "DHL",
		TrackingNumber: "TRACK-123",
		TrackingURL:    "https://tracking.example/TRACK-123",
		Notes:          "Handed off to carrier",
	})
	require.NoError(t, err)
	require.Equal(t, entity.RoutedOrderStatusShipped, order.Status)
	require.NotNil(t, order.ShippedAt)
	require.Nil(t, order.DeliveredAt)
	require.Equal(t, "DHL", order.ShipmentCarrier)
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "in transit via DHL")
	got := order.ActivityLog[len(order.ActivityLog)-1]
	require.Equal(t, entity.RoutedOrderActivityTypeShipmentNote, got.Type)
	require.Equal(t, "Handed off to carrier", got.Message)
	require.Equal(t, "system", got.Actor)
	require.True(t, hasActivityDetail(got.Details, "shipment_status", entity.RoutedOrderShipmentStatusInTransit))
}

func TestUpdateOrderShipmentMarksDelivered(t *testing.T) {
	t.Parallel()

	shippedAt := time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC)
	interactor, orders, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(entity.RoutedOrder{
		ID:             "ord-delivered",
		Status:         entity.RoutedOrderStatusShipped,
		ShipmentStatus: entity.RoutedOrderShipmentStatusInTransit,
		ShippedAt:      &shippedAt,
		Timeline:       []string{"created"},
	})

	order, err := interactor.UpdateOrderShipment(context.Background(), inputport.UpdateOrderShipmentCmd{
		OrderID:        "ord-delivered",
		ShipmentStatus: entity.RoutedOrderShipmentStatusDelivered,
		Notes:          "Delivered to customer",
	})
	require.NoError(t, err)
	require.NotNil(t, order.DeliveredAt)
	require.NotNil(t, order.ShippedAt)
	require.True(t, order.ShippedAt.Equal(shippedAt))
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "marked delivered")
}
