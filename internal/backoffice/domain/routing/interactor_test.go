package routing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	catalogentity "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/entity"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routinginputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/inputport"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func TestCreateRoutedOrderSnapshotsCostsAndMargin(t *testing.T) {
	t.Parallel()

	interactor, _, _, _ := newOrderRoutingTestInteractor(t, map[string]catalogentity.ProductSetupCandidate{
		"cand-1": {
			ID:          "cand-1",
			Title:       "Vintage Tee",
			Partner:     "Print Partner A",
			BaseCost:    "$8.00",
			RetailPrice: "$20.00",
			Status:      catalogentity.ProductSetupCandidateStatusPublishedMock,
		},
	})

	ctx := toolkit.WithTenantID(context.Background(), "t_demo")
	order, err := interactor.CreateRoutedOrder(ctx, routinginputport.CreateRoutedOrderCmd{
		CandidateID:  "cand-1",
		CustomerName: "Alex POD",
		Quantity:     2,
		ProductType:  "tshirt",
		ShipRegion:   "us",
	})
	require.NoError(t, err)
	require.Equal(t, routingentity.RoutedOrderStatusQueued, order.Status)
	require.Equal(t, routingentity.RoutedOrderShipmentStatusAwaitingLabel, order.ShipmentStatus)
	require.Equal(t, "unassigned", order.OperatorAssignee)
	require.Equal(t, "$16.00", order.BaseCostSnapshot)
	require.Equal(t, "Fulfill Fast", order.Partner)
	require.Equal(t, "$14.00", order.FulfillmentCost)
	require.Equal(t, "$2.00", order.ShippingCost)
	require.Equal(t, "$40.00", order.Total)
	require.Equal(t, "$24.00", order.RealizedMargin)
	require.Equal(t, routingentity.RoutedOrderSettlementStatusPending, order.SettlementStatus)
	require.Len(t, order.Timeline, 2)
	require.Len(t, order.ActivityLog, 2)
	require.Equal(t, "system", order.ActivityLog[0].Actor)
	require.NotEmpty(t, order.ActivityLog[0].Details)
}

func TestRecommendRoutedOrderPartnerPrefersEligibleRequestedPartner(t *testing.T) {
	t.Parallel()

	interactor, _, _, _ := newOrderRoutingTestInteractor(t, map[string]catalogentity.ProductSetupCandidate{
		"cand-1": {
			ID:      "cand-1",
			Title:   "Vintage Tee",
			Partner: "Print Partner A",
			Status:  catalogentity.ProductSetupCandidateStatusPublishedMock,
		},
	})

	ctx := toolkit.WithTenantID(context.Background(), "t_demo")
	recommendation, err := interactor.RecommendRoutedOrderPartner(
		ctx,
		routinginputport.RecommendRoutedOrderPartnerQuery{
			CandidateID:      "cand-1",
			ProductType:      "tshirt",
			ShipRegion:       "us",
			PreferredPartner: "Fulfill Fast",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "cand-1", recommendation.CandidateID)
	require.Equal(t, "Fulfill Fast", recommendation.SelectedPartner)
	require.Equal(t, "Print Partner A", recommendation.CandidatePartner)
	require.NotEmpty(t, recommendation.Options)
	foundPreferred := false
	for _, option := range recommendation.Options {
		if option.Partner.Name == "Fulfill Fast" {
			foundPreferred = true
			require.True(t, option.Eligible)
		}
	}
	require.True(t, foundPreferred)
}

func TestRecommendRoutedOrderPartnerPrefersHigherMarginOverCandidateDefault(t *testing.T) {
	t.Parallel()

	interactor, _, _, _ := newOrderRoutingTestInteractor(t, map[string]catalogentity.ProductSetupCandidate{
		"cand-1": {
			ID:          "cand-1",
			Title:       "Vintage Tee",
			Partner:     "Print Partner A",
			BaseCost:    "$8.00",
			RetailPrice: "$20.00",
			Status:      catalogentity.ProductSetupCandidateStatusPublishedMock,
		},
	})

	ctx := toolkit.WithTenantID(context.Background(), "t_demo")
	recommendation, err := interactor.RecommendRoutedOrderPartner(
		ctx,
		routinginputport.RecommendRoutedOrderPartnerQuery{
			CandidateID: "cand-1",
			ProductType: "tshirt",
			ShipRegion:  "us",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "Fulfill Fast", recommendation.SelectedPartner)
	require.Contains(t, recommendation.Summary, "over candidate default Print Partner A")
	require.Equal(t, "Fulfill Fast", recommendation.Options[0].Partner.Name)
	require.Equal(t, "$11.00", recommendation.Options[0].EstimatedUnitMargin)
}

func TestRecommendRoutedOrderPartnerDoesNotAutoSelectNegativeMarginOption(t *testing.T) {
	t.Parallel()

	interactor, _, _, _ := newOrderRoutingTestInteractor(t, map[string]catalogentity.ProductSetupCandidate{
		"cand-1": {
			ID:          "cand-1",
			Title:       "Poster",
			Partner:     "Print Partner A",
			BaseCost:    "$8.00",
			RetailPrice: "$8.00",
			Status:      catalogentity.ProductSetupCandidateStatusPublishedMock,
		},
	})

	ctx := toolkit.WithTenantID(context.Background(), "t_demo")
	recommendation, err := interactor.RecommendRoutedOrderPartner(
		ctx,
		routinginputport.RecommendRoutedOrderPartnerQuery{
			CandidateID: "cand-1",
			ProductType: "tshirt",
			ShipRegion:  "us",
		},
	)
	require.NoError(t, err)
	require.Empty(t, recommendation.SelectedPartner)
	require.Contains(t, recommendation.Summary, "negative expected margin")
	require.Len(t, recommendation.Options, 2)
	require.Equal(t, "Fulfill Fast", recommendation.Options[0].Partner.Name)
	require.Equal(t, "$-1.00", recommendation.Options[0].EstimatedUnitMargin)
	require.Equal(t, "Print Partner A", recommendation.Options[1].Partner.Name)
	require.Equal(t, "$-5.00", recommendation.Options[1].EstimatedUnitMargin)
	require.Equal(t, "negative_margin", recommendation.BlockedReasonCode)
	require.NotEmpty(t, recommendation.BlockedReason)
}

func TestCreateRoutedOrderCreatesBlockedQueueItemWhenNoViablePartner(t *testing.T) {
	t.Parallel()

	interactor, _, _, _ := newOrderRoutingTestInteractor(t, map[string]catalogentity.ProductSetupCandidate{
		"cand-1": {
			ID:          "cand-1",
			Title:       "Poster",
			Partner:     "Print Partner A",
			BaseCost:    "$8.00",
			RetailPrice: "$8.00",
			Status:      catalogentity.ProductSetupCandidateStatusPublishedMock,
		},
	})

	ctx := toolkit.WithTenantID(context.Background(), "t_demo")
	order, err := interactor.CreateRoutedOrder(ctx, routinginputport.CreateRoutedOrderCmd{
		CandidateID:  "cand-1",
		CustomerName: "Blocked Customer",
		Quantity:     1,
		ProductType:  "tshirt",
		ShipRegion:   "us",
	})
	require.NoError(t, err)
	require.Equal(t, routingentity.RoutedOrderStatusRoutingBlocked, order.Status)
	require.Empty(t, order.Partner)
	require.Equal(t, "negative_margin", order.RoutingBlockCode)
	require.NotEmpty(t, order.RoutingBlockReason)
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "Routing blocked")
	lastActivity := order.ActivityLog[len(order.ActivityLog)-1]
	require.True(t, hasActivityDetail(lastActivity.Details, "routing_block_code", "negative_margin"))
}

func TestForceRerouteBlockedOrderClearsBlockAndQueuesOrder(t *testing.T) {
	t.Parallel()

	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, map[string]catalogentity.ProductSetupCandidate{
		"cand-1": {
			ID:          "cand-1",
			Title:       "Poster",
			Partner:     "Print Partner A",
			BaseCost:    "$8.00",
			RetailPrice: "$8.00",
			Status:      catalogentity.ProductSetupCandidateStatusPublishedMock,
		},
	})
	orders.mustSeed(routingentity.RoutedOrder{
		ID:                 "ord-blocked-reroute",
		CandidateID:        "cand-1",
		ProductTitle:       "Poster",
		Quantity:           1,
		Total:              "$8.00",
		CustomerName:       "Blocked Customer",
		Status:             routingentity.RoutedOrderStatusRoutingBlocked,
		ShipmentStatus:     routingentity.RoutedOrderShipmentStatusAwaitingLabel,
		OperatorAssignee:   "unassigned",
		RoutingBlockCode:   "negative_margin",
		RoutingBlockReason: "all eligible partners have negative expected margin",
		BaseCostSnapshot:   "$8.00",
		FulfillmentCost:    "TBD",
		ShippingCost:       "TBD",
		IssueCost:          "$0.00",
		IssueResolution:    routingentity.RoutedOrderIssueResolutionMonitor,
		RealizedMargin:     "TBD",
		SettlementStatus:   routingentity.RoutedOrderSettlementStatusPending,
		Timeline:           []string{"created", "Routing blocked: all eligible partners have negative expected margin"},
		ActivityLog: []routingentity.RoutedOrderActivity{
			{
				Type:      routingentity.RoutedOrderActivityTypeSystem,
				Actor:     "system",
				Message:   "Order created for Poster",
				CreatedAt: time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC),
				Details: []routingentity.RoutedOrderActivityDetail{
					{Key: "product_type", Value: "poster"},
					{Key: "ship_region", Value: "us"},
				},
			},
		},
	})

	ctx := toolkit.WithTenantID(context.Background(), "t_demo")
	order, err := interactor.ForceRerouteBlockedOrder(ctx, routinginputport.ForceRerouteBlockedOrderCmd{
		OrderID:          "ord-blocked-reroute",
		PreferredPartner: "Fulfill Fast",
	})
	require.NoError(t, err)
	require.Equal(t, routingentity.RoutedOrderStatusQueued, order.Status)
	require.Equal(t, "Fulfill Fast", order.Partner)
	require.Equal(t, "", order.RoutingBlockCode)
	require.Equal(t, "", order.RoutingBlockReason)
	require.Equal(t, "$7.00", order.FulfillmentCost)
	require.Equal(t, "$2.00", order.ShippingCost)
	require.Equal(t, "$-1.00", order.RealizedMargin)
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "Routing unblocked")
	lastActivity := order.ActivityLog[len(order.ActivityLog)-1]
	require.True(t, hasActivityDetail(lastActivity.Details, "manual_reroute", "true"))
}

func TestUpdateOrderSettlementRecalculatesMarginIncludingIssueCost(t *testing.T) {
	t.Parallel()

	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:               "ord-1",
		Total:            "$40.00",
		FulfillmentCost:  "$10.00",
		ShippingCost:     "$4.00",
		IssueCost:        "$3.00",
		SettlementStatus: routingentity.RoutedOrderSettlementStatusPending,
		Timeline:         []string{"created"},
	})

	order, err := interactor.UpdateOrderSettlement(context.Background(), routinginputport.UpdateOrderSettlementCmd{
		OrderID:          "ord-1",
		FulfillmentCost:  "$12.00",
		ShippingCost:     "$5.50",
		SettlementStatus: routingentity.RoutedOrderSettlementStatusReconciled,
		Notes:            "Supplier invoice matched",
	})
	require.NoError(t, err)
	require.Equal(t, "$12.00", order.FulfillmentCost)
	require.Equal(t, "$5.50", order.ShippingCost)
	require.Equal(t, "$19.50", order.RealizedMargin)
	require.Equal(t, routingentity.RoutedOrderSettlementStatusReconciled, order.SettlementStatus)
	require.Equal(t, "Supplier invoice matched", order.SettlementNotes)
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "Settlement")
	got := order.ActivityLog[len(order.ActivityLog)-1]
	require.Equal(t, routingentity.RoutedOrderActivityTypeSettlementNote, got.Type)
	require.Equal(t, "Supplier invoice matched", got.Message)
	require.Equal(t, "system", got.Actor)
	require.True(
		t,
		hasActivityDetail(got.Details, "settlement_status", routingentity.RoutedOrderSettlementStatusReconciled),
	)
}

func TestUpdateOrderIssueHandlingRequiresActiveIssue(t *testing.T) {
	t.Parallel()

	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:               "ord-no-issue",
		Total:            "$40.00",
		FulfillmentCost:  "$10.00",
		ShippingCost:     "$5.00",
		IssueCost:        "$0.00",
		ShipmentStatus:   routingentity.RoutedOrderShipmentStatusAwaitingLabel,
		SettlementStatus: routingentity.RoutedOrderSettlementStatusPending,
	})

	_, err := interactor.UpdateOrderIssueHandling(context.Background(), routinginputport.UpdateOrderIssueHandlingCmd{
		OrderID:         "ord-no-issue",
		IssueCost:       "$6.00",
		IssueResolution: routingentity.RoutedOrderIssueResolutionReprint,
		Notes:           "Needs reprint",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "active exception or delivery issue")
}

func TestUpdateOrderIssueHandlingRecalculatesMargin(t *testing.T) {
	t.Parallel()

	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:               "ord-issue",
		Total:            "$40.00",
		FulfillmentCost:  "$10.00",
		ShippingCost:     "$5.00",
		IssueCost:        "$0.00",
		ExceptionType:    "reprint_request",
		ExceptionStatus:  routingentity.RoutedOrderExceptionStatusOpen,
		ShipmentStatus:   routingentity.RoutedOrderShipmentStatusAwaitingLabel,
		SettlementStatus: routingentity.RoutedOrderSettlementStatusPending,
		Timeline:         []string{"created"},
	})

	order, err := interactor.UpdateOrderIssueHandling(
		context.Background(),
		routinginputport.UpdateOrderIssueHandlingCmd{
			OrderID:         "ord-issue",
			IssueCost:       "$6.00",
			IssueResolution: routingentity.RoutedOrderIssueResolutionReprint,
			Notes:           "Reprint approved",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "$6.00", order.IssueCost)
	require.Equal(t, routingentity.RoutedOrderIssueResolutionReprint, order.IssueResolution)
	require.Equal(t, "$19.00", order.RealizedMargin)
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "Issue handling")
	got := order.ActivityLog[len(order.ActivityLog)-1]
	require.Equal(t, routingentity.RoutedOrderActivityTypeIssueNote, got.Type)
	require.Equal(t, "Reprint approved", got.Message)
	require.Equal(t, "system", got.Actor)
	require.True(t, hasActivityDetail(got.Details, "issue_resolution", routingentity.RoutedOrderIssueResolutionReprint))
}

func TestUpdateOrderQueueControlNormalizesAssigneeAndPersistsSLA(t *testing.T) {
	t.Parallel()

	shipmentSLA := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	issueSLA := shipmentSLA.Add(4 * time.Hour)
	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:       "ord-queue",
		Timeline: []string{"created"},
	})

	order, err := interactor.UpdateOrderQueueControl(context.Background(), routinginputport.UpdateOrderQueueControlCmd{
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
	settlementStatus := routingentity.RoutedOrderSettlementStatusPaid
	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:               "ord-1",
		OperatorAssignee: "unassigned",
		SettlementStatus: routingentity.RoutedOrderSettlementStatusPending,
		Timeline:         []string{"created"},
	})
	orders.mustSeed(routingentity.RoutedOrder{
		ID:               "ord-2",
		OperatorAssignee: "unassigned",
		SettlementStatus: routingentity.RoutedOrderSettlementStatusPending,
		Timeline:         []string{"created"},
	})

	updated, err := interactor.BulkUpdateRoutedOrders(context.Background(), routinginputport.BulkUpdateRoutedOrdersCmd{
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
		require.Equal(t, routingentity.RoutedOrderSettlementStatusPaid, order.SettlementStatus)
		require.Contains(t, order.Timeline[len(order.Timeline)-1], "Bulk queue update applied")
	}
}

func TestListRoutedOrderActivitiesFiltersAndSorts(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:               "ord-1",
		ProductTitle:     "Vintage Tee",
		OperatorAssignee: "ops.lead",
		ActivityLog: []routingentity.RoutedOrderActivity{
			{
				Type:      routingentity.RoutedOrderActivityTypeSystem,
				Actor:     "system",
				Message:   "Queued",
				CreatedAt: now.Add(-2 * time.Hour),
			},
			{
				Type:      routingentity.RoutedOrderActivityTypeShipmentNote,
				Actor:     "user:12",
				Message:   "Carrier assigned",
				CreatedAt: now.Add(-1 * time.Hour),
			},
		},
	})
	orders.mustSeed(routingentity.RoutedOrder{
		ID:               "ord-2",
		ProductTitle:     "Poster",
		OperatorAssignee: "ops.a",
		ActivityLog: []routingentity.RoutedOrderActivity{
			{
				Type:      routingentity.RoutedOrderActivityTypeShipmentNote,
				Actor:     "user:15",
				Message:   "Packed",
				CreatedAt: now.Add(-30 * time.Minute),
			},
		},
	})

	firstPage, err := interactor.ListRoutedOrderActivities(
		context.Background(),
		routingentity.RoutedOrderActivityFeedQuery{
			ActivityType:  routingentity.RoutedOrderActivityTypeShipmentNote,
			ActorContains: "user:",
			Since:         ptrTime(now.Add(-90 * time.Minute)),
			Limit:         1,
			IncludeSystem: false,
		},
	)
	require.NoError(t, err)
	require.Equal(t, 2, firstPage.Total)
	require.Len(t, firstPage.Entries, 1)
	require.Equal(t, "ord-2", firstPage.Entries[0].OrderID)
	require.NotNil(t, firstPage.NextCursor)
	require.NotEmpty(t, *firstPage.NextCursor)

	secondPage, err := interactor.ListRoutedOrderActivities(
		context.Background(),
		routingentity.RoutedOrderActivityFeedQuery{
			ActivityType:  routingentity.RoutedOrderActivityTypeShipmentNote,
			ActorContains: "user:",
			Since:         ptrTime(now.Add(-90 * time.Minute)),
			Limit:         1,
			After:         *firstPage.NextCursor,
			IncludeSystem: false,
		},
	)
	require.NoError(t, err)
	require.Len(t, secondPage.Entries, 1)
	require.Equal(t, "ord-1", secondPage.Entries[0].OrderID)
	require.Nil(t, secondPage.NextCursor)
}

func TestAdvanceRoutedOrderBlocksWhenExceptionActive(t *testing.T) {
	t.Parallel()

	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:              "ord-blocked",
		Status:          routingentity.RoutedOrderStatusQueued,
		ExceptionType:   "partner_delay",
		ExceptionStatus: routingentity.RoutedOrderExceptionStatusOpen,
	})

	_, err := interactor.AdvanceRoutedOrder(context.Background(), "ord-blocked")
	require.Error(t, err)
	require.Contains(t, err.Error(), "resolve the active exception")
}

func TestAdvanceRoutedOrderBlocksWhenRoutingBlocked(t *testing.T) {
	t.Parallel()

	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:                 "ord-routing-blocked",
		Status:             routingentity.RoutedOrderStatusRoutingBlocked,
		RoutingBlockCode:   "negative_margin",
		RoutingBlockReason: "all eligible partners have negative expected margin",
	})

	_, err := interactor.AdvanceRoutedOrder(context.Background(), "ord-routing-blocked")
	require.Error(t, err)
	require.Contains(t, err.Error(), "resolve the routing block")
}

func TestAdvanceRoutedOrderTransitionsQueuedToProduction(t *testing.T) {
	t.Parallel()

	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:       "ord-advance",
		Status:   routingentity.RoutedOrderStatusQueued,
		Partner:  "Print Partner A",
		Timeline: []string{"created"},
	})

	order, err := interactor.AdvanceRoutedOrder(context.Background(), "ord-advance")
	require.NoError(t, err)
	require.Equal(t, routingentity.RoutedOrderStatusInProduction, order.Status)
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "POD production")
}

func TestOpenAndResolveOrderException(t *testing.T) {
	t.Parallel()

	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:       "ord-exception",
		Timeline: []string{"created"},
	})

	opened, err := interactor.OpenOrderException(context.Background(), routinginputport.OpenOrderExceptionCmd{
		OrderID:       "ord-exception",
		ExceptionType: "reprint_request",
	})
	require.NoError(t, err)
	require.Equal(t, "reprint_request", opened.ExceptionType)
	require.Equal(t, routingentity.RoutedOrderExceptionStatusOpen, opened.ExceptionStatus)
	require.Contains(t, opened.Timeline[len(opened.Timeline)-1], "Exception opened")

	resolved, err := interactor.UpdateOrderExceptionStatus(
		context.Background(),
		routinginputport.UpdateOrderExceptionStatusCmd{
			OrderID: "ord-exception",
			Status:  routingentity.RoutedOrderExceptionStatusResolved,
		},
	)
	require.NoError(t, err)
	require.Equal(t, routingentity.RoutedOrderExceptionStatusResolved, resolved.ExceptionStatus)
	require.Contains(t, resolved.Timeline[len(resolved.Timeline)-1], "Exception resolved")
}

func TestUpdateOrderShipmentMarksInTransitAndAppendsTracking(t *testing.T) {
	t.Parallel()

	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:             "ord-ship",
		Status:         routingentity.RoutedOrderStatusQueued,
		ShipmentStatus: routingentity.RoutedOrderShipmentStatusAwaitingLabel,
		Partner:        "Print Partner A",
		Timeline:       []string{"created"},
	})

	order, err := interactor.UpdateOrderShipment(context.Background(), routinginputport.UpdateOrderShipmentCmd{
		OrderID:        "ord-ship",
		ShipmentStatus: routingentity.RoutedOrderShipmentStatusInTransit,
		Carrier:        "DHL",
		TrackingNumber: "TRACK-123",
		TrackingURL:    "https://tracking.example/TRACK-123",
		Notes:          "Handed off to carrier",
	})
	require.NoError(t, err)
	require.Equal(t, routingentity.RoutedOrderStatusShipped, order.Status)
	require.NotNil(t, order.ShippedAt)
	require.Nil(t, order.DeliveredAt)
	require.Equal(t, "DHL", order.ShipmentCarrier)
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "in transit via DHL")
	got := order.ActivityLog[len(order.ActivityLog)-1]
	require.Equal(t, routingentity.RoutedOrderActivityTypeShipmentNote, got.Type)
	require.Equal(t, "Handed off to carrier", got.Message)
	require.Equal(t, "system", got.Actor)
	require.True(t, hasActivityDetail(got.Details, "shipment_status", routingentity.RoutedOrderShipmentStatusInTransit))
}

func TestUpdateOrderShipmentMarksDelivered(t *testing.T) {
	t.Parallel()

	shippedAt := time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC)
	interactor, orders, _, _ := newOrderRoutingTestInteractor(t, nil)
	orders.mustSeed(routingentity.RoutedOrder{
		ID:             "ord-delivered",
		Status:         routingentity.RoutedOrderStatusShipped,
		ShipmentStatus: routingentity.RoutedOrderShipmentStatusInTransit,
		ShippedAt:      &shippedAt,
		Timeline:       []string{"created"},
	})

	order, err := interactor.UpdateOrderShipment(context.Background(), routinginputport.UpdateOrderShipmentCmd{
		OrderID:        "ord-delivered",
		ShipmentStatus: routingentity.RoutedOrderShipmentStatusDelivered,
		Notes:          "Delivered to customer",
	})
	require.NoError(t, err)
	require.NotNil(t, order.DeliveredAt)
	require.NotNil(t, order.ShippedAt)
	require.True(t, order.ShippedAt.Equal(shippedAt))
	require.Contains(t, order.Timeline[len(order.Timeline)-1], "marked delivered")
}
