package interactor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
)

func TestCreateRoutedOrderSnapshotsCostsAndMargin(t *testing.T) {
	t.Parallel()

	products := &fakeProductSetupRepository{
		candidates: map[string]entity.ProductSetupCandidate{
			"cand-1": {
				ID:          "cand-1",
				Title:       "Vintage Tee",
				Partner:     "Print Partner A",
				BaseCost:    "$8.00",
				RetailPrice: "$20.00",
				Status:      entity.ProductSetupCandidateStatusPublishedMock,
			},
		},
	}
	orders := newFakeOrderRoutingRepository()
	interactor := &OrderRoutingInteractor{orders: orders, products: products}

	order, err := interactor.CreateRoutedOrder(context.Background(), inputport.CreateRoutedOrderCmd{
		CandidateID:  "cand-1",
		CustomerName: "Alex POD",
		Quantity:     2,
	})
	if err != nil {
		t.Fatalf("CreateRoutedOrder() error = %v", err)
	}

	if order.Status != entity.RoutedOrderStatusQueued {
		t.Fatalf("Status = %q, want %q", order.Status, entity.RoutedOrderStatusQueued)
	}
	if order.ShipmentStatus != entity.RoutedOrderShipmentStatusAwaitingLabel {
		t.Fatalf("ShipmentStatus = %q, want %q", order.ShipmentStatus, entity.RoutedOrderShipmentStatusAwaitingLabel)
	}
	if order.OperatorAssignee != "unassigned" {
		t.Fatalf("OperatorAssignee = %q, want unassigned", order.OperatorAssignee)
	}
	if order.BaseCostSnapshot != "$16.00" {
		t.Fatalf("BaseCostSnapshot = %q, want $16.00", order.BaseCostSnapshot)
	}
	if order.FulfillmentCost != "$16.00" {
		t.Fatalf("FulfillmentCost = %q, want $16.00", order.FulfillmentCost)
	}
	if order.Total != "$40.00" {
		t.Fatalf("Total = %q, want $40.00", order.Total)
	}
	if order.RealizedMargin != "$24.00" {
		t.Fatalf("RealizedMargin = %q, want $24.00", order.RealizedMargin)
	}
	if order.SettlementStatus != entity.RoutedOrderSettlementStatusPending {
		t.Fatalf("SettlementStatus = %q, want %q", order.SettlementStatus, entity.RoutedOrderSettlementStatusPending)
	}
	if len(order.Timeline) < 2 {
		t.Fatalf("Timeline length = %d, want at least 2", len(order.Timeline))
	}
	if len(order.ActivityLog) < 2 {
		t.Fatalf("ActivityLog length = %d, want at least 2", len(order.ActivityLog))
	}
	if order.ActivityLog[0].Actor != "system" {
		t.Fatalf("ActivityLog[0].Actor = %q, want system", order.ActivityLog[0].Actor)
	}
	if len(order.ActivityLog[0].Details) == 0 {
		t.Fatal("ActivityLog[0].Details is empty, want structured details")
	}
}

func TestUpdateOrderSettlementRecalculatesMarginIncludingIssueCost(t *testing.T) {
	t.Parallel()

	orders := newFakeOrderRoutingRepository()
	orders.mustSeed(entity.RoutedOrder{
		ID:               "ord-1",
		Total:            "$40.00",
		FulfillmentCost:  "$10.00",
		ShippingCost:     "$4.00",
		IssueCost:        "$3.00",
		SettlementStatus: entity.RoutedOrderSettlementStatusPending,
		Timeline:         []string{"created"},
	})
	interactor := &OrderRoutingInteractor{orders: orders, products: &fakeProductSetupRepository{}}

	order, err := interactor.UpdateOrderSettlement(context.Background(), inputport.UpdateOrderSettlementCmd{
		OrderID:          "ord-1",
		FulfillmentCost:  "$12.00",
		ShippingCost:     "$5.50",
		SettlementStatus: entity.RoutedOrderSettlementStatusReconciled,
		Notes:            "Supplier invoice matched",
	})
	if err != nil {
		t.Fatalf("UpdateOrderSettlement() error = %v", err)
	}

	if order.FulfillmentCost != "$12.00" {
		t.Fatalf("FulfillmentCost = %q, want $12.00", order.FulfillmentCost)
	}
	if order.ShippingCost != "$5.50" {
		t.Fatalf("ShippingCost = %q, want $5.50", order.ShippingCost)
	}
	if order.RealizedMargin != "$19.50" {
		t.Fatalf("RealizedMargin = %q, want $19.50", order.RealizedMargin)
	}
	if order.SettlementStatus != entity.RoutedOrderSettlementStatusReconciled {
		t.Fatalf("SettlementStatus = %q, want %q", order.SettlementStatus, entity.RoutedOrderSettlementStatusReconciled)
	}
	if order.SettlementNotes != "Supplier invoice matched" {
		t.Fatalf("SettlementNotes = %q, want Supplier invoice matched", order.SettlementNotes)
	}
	if !strings.Contains(order.Timeline[len(order.Timeline)-1], "Settlement") {
		t.Fatalf("Timeline tail = %q, want settlement entry", order.Timeline[len(order.Timeline)-1])
	}
	if got := order.ActivityLog[len(order.ActivityLog)-1]; got.Type != entity.RoutedOrderActivityTypeSettlementNote || got.Message != "Supplier invoice matched" || got.Actor != "system" {
		t.Fatalf("ActivityLog tail = %#v, want settlement note entry", got)
	}
	if got := order.ActivityLog[len(order.ActivityLog)-1]; !hasActivityDetail(got.Details, "settlement_status", entity.RoutedOrderSettlementStatusReconciled) {
		t.Fatalf("ActivityLog details = %#v, want settlement_status detail", got.Details)
	}
}

func TestUpdateOrderIssueHandlingRequiresActiveIssue(t *testing.T) {
	t.Parallel()

	orders := newFakeOrderRoutingRepository()
	orders.mustSeed(entity.RoutedOrder{
		ID:               "ord-no-issue",
		Total:            "$40.00",
		FulfillmentCost:  "$10.00",
		ShippingCost:     "$5.00",
		IssueCost:        "$0.00",
		ShipmentStatus:   entity.RoutedOrderShipmentStatusAwaitingLabel,
		SettlementStatus: entity.RoutedOrderSettlementStatusPending,
	})
	interactor := &OrderRoutingInteractor{orders: orders, products: &fakeProductSetupRepository{}}

	_, err := interactor.UpdateOrderIssueHandling(context.Background(), inputport.UpdateOrderIssueHandlingCmd{
		OrderID:         "ord-no-issue",
		IssueCost:       "$6.00",
		IssueResolution: entity.RoutedOrderIssueResolutionReprint,
		Notes:           "Needs reprint",
	})
	if err == nil || !strings.Contains(err.Error(), "active exception or delivery issue") {
		t.Fatalf("UpdateOrderIssueHandling() error = %v, want active issue validation", err)
	}
}

func TestUpdateOrderIssueHandlingRecalculatesMargin(t *testing.T) {
	t.Parallel()

	orders := newFakeOrderRoutingRepository()
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
	interactor := &OrderRoutingInteractor{orders: orders, products: &fakeProductSetupRepository{}}

	order, err := interactor.UpdateOrderIssueHandling(context.Background(), inputport.UpdateOrderIssueHandlingCmd{
		OrderID:         "ord-issue",
		IssueCost:       "$6.00",
		IssueResolution: entity.RoutedOrderIssueResolutionReprint,
		Notes:           "Reprint approved",
	})
	if err != nil {
		t.Fatalf("UpdateOrderIssueHandling() error = %v", err)
	}

	if order.IssueCost != "$6.00" {
		t.Fatalf("IssueCost = %q, want $6.00", order.IssueCost)
	}
	if order.IssueResolution != entity.RoutedOrderIssueResolutionReprint {
		t.Fatalf("IssueResolution = %q, want %q", order.IssueResolution, entity.RoutedOrderIssueResolutionReprint)
	}
	if order.RealizedMargin != "$19.00" {
		t.Fatalf("RealizedMargin = %q, want $19.00", order.RealizedMargin)
	}
	if !strings.Contains(order.Timeline[len(order.Timeline)-1], "Issue handling") {
		t.Fatalf("Timeline tail = %q, want issue handling entry", order.Timeline[len(order.Timeline)-1])
	}
	if got := order.ActivityLog[len(order.ActivityLog)-1]; got.Type != entity.RoutedOrderActivityTypeIssueNote || got.Message != "Reprint approved" || got.Actor != "system" {
		t.Fatalf("ActivityLog tail = %#v, want issue note entry", got)
	}
	if got := order.ActivityLog[len(order.ActivityLog)-1]; !hasActivityDetail(got.Details, "issue_resolution", entity.RoutedOrderIssueResolutionReprint) {
		t.Fatalf("ActivityLog details = %#v, want issue_resolution detail", got.Details)
	}
}

func TestUpdateOrderQueueControlNormalizesAssigneeAndPersistsSLA(t *testing.T) {
	t.Parallel()

	shipmentSLA := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	issueSLA := shipmentSLA.Add(4 * time.Hour)
	orders := newFakeOrderRoutingRepository()
	orders.mustSeed(entity.RoutedOrder{
		ID:       "ord-queue",
		Timeline: []string{"created"},
	})
	interactor := &OrderRoutingInteractor{orders: orders, products: &fakeProductSetupRepository{}}

	order, err := interactor.UpdateOrderQueueControl(context.Background(), inputport.UpdateOrderQueueControlCmd{
		OrderID:          "ord-queue",
		OperatorAssignee: "  ",
		ShipmentSlaDueAt: &shipmentSLA,
		IssueSlaDueAt:    &issueSLA,
	})
	if err != nil {
		t.Fatalf("UpdateOrderQueueControl() error = %v", err)
	}

	if order.OperatorAssignee != "unassigned" {
		t.Fatalf("OperatorAssignee = %q, want unassigned", order.OperatorAssignee)
	}
	if order.ShipmentSlaDueAt == nil || !order.ShipmentSlaDueAt.Equal(shipmentSLA) {
		t.Fatalf("ShipmentSlaDueAt = %v, want %v", order.ShipmentSlaDueAt, shipmentSLA)
	}
	if order.IssueSlaDueAt == nil || !order.IssueSlaDueAt.Equal(issueSLA) {
		t.Fatalf("IssueSlaDueAt = %v, want %v", order.IssueSlaDueAt, issueSLA)
	}
	if !strings.Contains(order.Timeline[len(order.Timeline)-1], "Queue ownership updated") {
		t.Fatalf("Timeline tail = %q, want queue ownership entry", order.Timeline[len(order.Timeline)-1])
	}
}

func TestBulkUpdateRoutedOrdersUpdatesSelectedOrders(t *testing.T) {
	t.Parallel()

	shipmentSLA := time.Date(2026, 5, 15, 18, 30, 0, 0, time.UTC)
	assignee := "ops.lead"
	settlementStatus := entity.RoutedOrderSettlementStatusPaid
	orders := newFakeOrderRoutingRepository()
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
	interactor := &OrderRoutingInteractor{orders: orders, products: &fakeProductSetupRepository{}}

	updated, err := interactor.BulkUpdateRoutedOrders(context.Background(), inputport.BulkUpdateRoutedOrdersCmd{
		OrderIDs:         []string{"ord-1", " ord-2 "},
		OperatorAssignee: &assignee,
		ShipmentSlaDueAt: &shipmentSLA,
		SettlementStatus: &settlementStatus,
	})
	if err != nil {
		t.Fatalf("BulkUpdateRoutedOrders() error = %v", err)
	}
	if len(updated) != 2 {
		t.Fatalf("updated length = %d, want 2", len(updated))
	}

	for _, order := range updated {
		if order.OperatorAssignee != "ops.lead" {
			t.Fatalf("OperatorAssignee = %q, want ops.lead", order.OperatorAssignee)
		}
		if order.ShipmentSlaDueAt == nil || !order.ShipmentSlaDueAt.Equal(shipmentSLA) {
			t.Fatalf("ShipmentSlaDueAt = %v, want %v", order.ShipmentSlaDueAt, shipmentSLA)
		}
		if order.SettlementStatus != entity.RoutedOrderSettlementStatusPaid {
			t.Fatalf("SettlementStatus = %q, want %q", order.SettlementStatus, entity.RoutedOrderSettlementStatusPaid)
		}
		if !strings.Contains(order.Timeline[len(order.Timeline)-1], "Bulk queue update applied") {
			t.Fatalf("Timeline tail = %q, want bulk queue update entry", order.Timeline[len(order.Timeline)-1])
		}
	}
}

func TestListRoutedOrderActivitiesFiltersAndSorts(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	orders := newFakeOrderRoutingRepository()
	orders.mustSeed(entity.RoutedOrder{
		ID:               "ord-1",
		ProductTitle:     "Vintage Tee",
		OperatorAssignee: "ops.lead",
		ActivityLog: []entity.RoutedOrderActivity{
			{
				Type:      entity.RoutedOrderActivityTypeSystem,
				Actor:     "system",
				Message:   "Queued",
				CreatedAt: now.Add(-2 * time.Hour),
			},
			{
				Type:      entity.RoutedOrderActivityTypeShipmentNote,
				Actor:     "user:12",
				Message:   "Carrier assigned",
				CreatedAt: now.Add(-1 * time.Hour),
			},
		},
	})
	orders.mustSeed(entity.RoutedOrder{
		ID:               "ord-2",
		ProductTitle:     "Poster",
		OperatorAssignee: "ops.a",
		ActivityLog: []entity.RoutedOrderActivity{
			{
				Type:      entity.RoutedOrderActivityTypeShipmentNote,
				Actor:     "user:15",
				Message:   "Packed",
				CreatedAt: now.Add(-30 * time.Minute),
			},
		},
	})
	interactor := &OrderRoutingInteractor{orders: orders, products: &fakeProductSetupRepository{}}

	firstPage, err := interactor.ListRoutedOrderActivities(context.Background(), inputport.ListRoutedOrderActivitiesQuery{
		ActivityType:  entity.RoutedOrderActivityTypeShipmentNote,
		ActorContains: "user:",
		Since:         ptrTime(now.Add(-90 * time.Minute)),
		Limit:         1,
		IncludeSystem: false,
	})
	if err != nil {
		t.Fatalf("ListRoutedOrderActivities() error = %v", err)
	}
	if firstPage.Total != 2 {
		t.Fatalf("firstPage.Total = %d, want 2", firstPage.Total)
	}
	if len(firstPage.Entries) != 1 {
		t.Fatalf("entries length = %d, want 1", len(firstPage.Entries))
	}
	if firstPage.Entries[0].OrderID != "ord-2" {
		t.Fatalf("entries[0].OrderID = %q, want ord-2", firstPage.Entries[0].OrderID)
	}
	if firstPage.NextCursor == nil || *firstPage.NextCursor == "" {
		t.Fatalf("firstPage.NextCursor = %v, want populated cursor", firstPage.NextCursor)
	}

	secondPage, err := interactor.ListRoutedOrderActivities(context.Background(), inputport.ListRoutedOrderActivitiesQuery{
		ActivityType:  entity.RoutedOrderActivityTypeShipmentNote,
		ActorContains: "user:",
		Since:         ptrTime(now.Add(-90 * time.Minute)),
		Limit:         1,
		After:         *firstPage.NextCursor,
		IncludeSystem: false,
	})
	if err != nil {
		t.Fatalf("ListRoutedOrderActivities() second page error = %v", err)
	}
	if len(secondPage.Entries) != 1 {
		t.Fatalf("secondPage entries length = %d, want 1", len(secondPage.Entries))
	}
	if secondPage.Entries[0].OrderID != "ord-1" {
		t.Fatalf("secondPage.entries[0].OrderID = %q, want ord-1", secondPage.Entries[0].OrderID)
	}
	if secondPage.NextCursor != nil {
		t.Fatalf("secondPage.NextCursor = %v, want nil on last page", secondPage.NextCursor)
	}
}

func TestAdvanceRoutedOrderBlocksWhenExceptionActive(t *testing.T) {
	t.Parallel()

	orders := newFakeOrderRoutingRepository()
	orders.mustSeed(entity.RoutedOrder{
		ID:              "ord-blocked",
		Status:          entity.RoutedOrderStatusQueued,
		ExceptionType:   "partner_delay",
		ExceptionStatus: entity.RoutedOrderExceptionStatusOpen,
	})
	interactor := &OrderRoutingInteractor{orders: orders, products: &fakeProductSetupRepository{}}

	_, err := interactor.AdvanceRoutedOrder(context.Background(), "ord-blocked")
	if err == nil || !strings.Contains(err.Error(), "resolve the active exception") {
		t.Fatalf("AdvanceRoutedOrder() error = %v, want active exception validation", err)
	}
}

func TestAdvanceRoutedOrderTransitionsQueuedToProduction(t *testing.T) {
	t.Parallel()

	orders := newFakeOrderRoutingRepository()
	orders.mustSeed(entity.RoutedOrder{
		ID:       "ord-advance",
		Status:   entity.RoutedOrderStatusQueued,
		Partner:  "Print Partner A",
		Timeline: []string{"created"},
	})
	interactor := &OrderRoutingInteractor{orders: orders, products: &fakeProductSetupRepository{}}

	order, err := interactor.AdvanceRoutedOrder(context.Background(), "ord-advance")
	if err != nil {
		t.Fatalf("AdvanceRoutedOrder() error = %v", err)
	}
	if order.Status != entity.RoutedOrderStatusInProduction {
		t.Fatalf("Status = %q, want %q", order.Status, entity.RoutedOrderStatusInProduction)
	}
	if !strings.Contains(order.Timeline[len(order.Timeline)-1], "POD production") {
		t.Fatalf("Timeline tail = %q, want production routing entry", order.Timeline[len(order.Timeline)-1])
	}
}

func TestOpenAndResolveOrderException(t *testing.T) {
	t.Parallel()

	orders := newFakeOrderRoutingRepository()
	orders.mustSeed(entity.RoutedOrder{
		ID:       "ord-exception",
		Timeline: []string{"created"},
	})
	interactor := &OrderRoutingInteractor{orders: orders, products: &fakeProductSetupRepository{}}

	opened, err := interactor.OpenOrderException(context.Background(), inputport.OpenOrderExceptionCmd{
		OrderID:       "ord-exception",
		ExceptionType: "reprint_request",
	})
	if err != nil {
		t.Fatalf("OpenOrderException() error = %v", err)
	}
	if opened.ExceptionType != "reprint_request" {
		t.Fatalf("ExceptionType = %q, want reprint_request", opened.ExceptionType)
	}
	if opened.ExceptionStatus != entity.RoutedOrderExceptionStatusOpen {
		t.Fatalf("ExceptionStatus = %q, want %q", opened.ExceptionStatus, entity.RoutedOrderExceptionStatusOpen)
	}
	if !strings.Contains(opened.Timeline[len(opened.Timeline)-1], "Exception opened") {
		t.Fatalf("Timeline tail = %q, want exception opened entry", opened.Timeline[len(opened.Timeline)-1])
	}

	resolved, err := interactor.UpdateOrderExceptionStatus(context.Background(), inputport.UpdateOrderExceptionStatusCmd{
		OrderID: "ord-exception",
		Status:  entity.RoutedOrderExceptionStatusResolved,
	})
	if err != nil {
		t.Fatalf("UpdateOrderExceptionStatus() error = %v", err)
	}
	if resolved.ExceptionStatus != entity.RoutedOrderExceptionStatusResolved {
		t.Fatalf("ExceptionStatus = %q, want %q", resolved.ExceptionStatus, entity.RoutedOrderExceptionStatusResolved)
	}
	if !strings.Contains(resolved.Timeline[len(resolved.Timeline)-1], "Exception resolved") {
		t.Fatalf("Timeline tail = %q, want exception resolved entry", resolved.Timeline[len(resolved.Timeline)-1])
	}
}

func TestUpdateOrderShipmentMarksInTransitAndAppendsTracking(t *testing.T) {
	t.Parallel()

	orders := newFakeOrderRoutingRepository()
	orders.mustSeed(entity.RoutedOrder{
		ID:             "ord-ship",
		Status:         entity.RoutedOrderStatusQueued,
		ShipmentStatus: entity.RoutedOrderShipmentStatusAwaitingLabel,
		Partner:        "Print Partner A",
		Timeline:       []string{"created"},
	})
	interactor := &OrderRoutingInteractor{orders: orders, products: &fakeProductSetupRepository{}}

	order, err := interactor.UpdateOrderShipment(context.Background(), inputport.UpdateOrderShipmentCmd{
		OrderID:        "ord-ship",
		ShipmentStatus: entity.RoutedOrderShipmentStatusInTransit,
		Carrier:        "DHL",
		TrackingNumber: "TRACK-123",
		TrackingURL:    "https://tracking.example/TRACK-123",
		Notes:          "Handed off to carrier",
	})
	if err != nil {
		t.Fatalf("UpdateOrderShipment() error = %v", err)
	}
	if order.Status != entity.RoutedOrderStatusShipped {
		t.Fatalf("Status = %q, want %q", order.Status, entity.RoutedOrderStatusShipped)
	}
	if order.ShippedAt == nil {
		t.Fatal("ShippedAt = nil, want timestamp")
	}
	if order.DeliveredAt != nil {
		t.Fatalf("DeliveredAt = %v, want nil", order.DeliveredAt)
	}
	if order.ShipmentCarrier != "DHL" {
		t.Fatalf("ShipmentCarrier = %q, want DHL", order.ShipmentCarrier)
	}
	if !strings.Contains(order.Timeline[len(order.Timeline)-1], "in transit via DHL") {
		t.Fatalf("Timeline tail = %q, want in-transit tracking entry", order.Timeline[len(order.Timeline)-1])
	}
	if got := order.ActivityLog[len(order.ActivityLog)-1]; got.Type != entity.RoutedOrderActivityTypeShipmentNote || got.Message != "Handed off to carrier" || got.Actor != "system" {
		t.Fatalf("ActivityLog tail = %#v, want shipment note entry", got)
	}
	if got := order.ActivityLog[len(order.ActivityLog)-1]; !hasActivityDetail(got.Details, "shipment_status", entity.RoutedOrderShipmentStatusInTransit) {
		t.Fatalf("ActivityLog details = %#v, want shipment_status detail", got.Details)
	}
}

func TestUpdateOrderShipmentMarksDelivered(t *testing.T) {
	t.Parallel()

	shippedAt := time.Date(2026, 5, 15, 8, 0, 0, 0, time.UTC)
	orders := newFakeOrderRoutingRepository()
	orders.mustSeed(entity.RoutedOrder{
		ID:             "ord-delivered",
		Status:         entity.RoutedOrderStatusShipped,
		ShipmentStatus: entity.RoutedOrderShipmentStatusInTransit,
		ShippedAt:      &shippedAt,
		Timeline:       []string{"created"},
	})
	interactor := &OrderRoutingInteractor{orders: orders, products: &fakeProductSetupRepository{}}

	order, err := interactor.UpdateOrderShipment(context.Background(), inputport.UpdateOrderShipmentCmd{
		OrderID:        "ord-delivered",
		ShipmentStatus: entity.RoutedOrderShipmentStatusDelivered,
		Notes:          "Delivered to customer",
	})
	if err != nil {
		t.Fatalf("UpdateOrderShipment() error = %v", err)
	}
	if order.DeliveredAt == nil {
		t.Fatal("DeliveredAt = nil, want timestamp")
	}
	if order.ShippedAt == nil || !order.ShippedAt.Equal(shippedAt) {
		t.Fatalf("ShippedAt = %v, want %v", order.ShippedAt, shippedAt)
	}
	if !strings.Contains(order.Timeline[len(order.Timeline)-1], "marked delivered") {
		t.Fatalf("Timeline tail = %q, want delivered entry", order.Timeline[len(order.Timeline)-1])
	}
}

type fakeOrderRoutingRepository struct {
	orders map[string]entity.RoutedOrder
}

func newFakeOrderRoutingRepository() *fakeOrderRoutingRepository {
	return &fakeOrderRoutingRepository{orders: map[string]entity.RoutedOrder{}}
}

func (r *fakeOrderRoutingRepository) mustSeed(order entity.RoutedOrder) {
	r.orders[order.ID] = cloneOrder(order)
}

func (r *fakeOrderRoutingRepository) List(_ context.Context) ([]entity.RoutedOrder, error) {
	orders := make([]entity.RoutedOrder, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, cloneOrder(order))
	}
	return orders, nil
}

func (r *fakeOrderRoutingRepository) ListActivityFeed(_ context.Context, query inputport.ListRoutedOrderActivitiesQuery) (*entity.RoutedOrderActivityFeedPage, error) {
	entries := make([]entity.RoutedOrderActivityFeedEntry, 0)
	for _, order := range r.orders {
		for _, activity := range order.ActivityLog {
			if query.ActivityType == "notes" && activity.Type == entity.RoutedOrderActivityTypeSystem {
				continue
			}
			if query.ActivityType != "" && query.ActivityType != "all" && query.ActivityType != "notes" && activity.Type != query.ActivityType {
				continue
			}
			if !query.IncludeSystem && query.ActivityType != "system" && query.ActivityType != "notes" && activity.Type == entity.RoutedOrderActivityTypeSystem {
				continue
			}
			if query.Since != nil && activity.CreatedAt.Before(query.Since.UTC()) {
				continue
			}
			if query.ActorContains != "" && !strings.Contains(strings.ToLower(activity.Actor), strings.ToLower(query.ActorContains)) {
				continue
			}
			entries = append(entries, entity.RoutedOrderActivityFeedEntry{
				OrderID:          order.ID,
				ProductTitle:     order.ProductTitle,
				OperatorAssignee: order.OperatorAssignee,
				Activity:         activity,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Activity.CreatedAt.After(entries[j].Activity.CreatedAt)
	})
	total := len(entries)
	start := 0
	if query.After != "" {
		for idx, entry := range entries {
			if encodeTestActivityCursor(entry) == query.After {
				start = idx + 1
				break
			}
		}
	}
	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}
	end := start + limit
	if end > total {
		end = total
	}
	pageEntries := entries[start:end]
	var nextCursor *string
	if end < total && len(pageEntries) > 0 {
		value := encodeTestActivityCursor(pageEntries[len(pageEntries)-1])
		nextCursor = &value
	}
	return &entity.RoutedOrderActivityFeedPage{
		Entries:    pageEntries,
		Total:      total,
		NextCursor: nextCursor,
	}, nil
}

func (r *fakeOrderRoutingRepository) GetByID(_ context.Context, id string) (*entity.RoutedOrder, error) {
	order, ok := r.orders[id]
	if !ok {
		return nil, nil
	}
	cloned := cloneOrder(order)
	return &cloned, nil
}

func (r *fakeOrderRoutingRepository) Create(_ context.Context, order entity.RoutedOrder) (*entity.RoutedOrder, error) {
	r.orders[order.ID] = cloneOrder(order)
	cloned := cloneOrder(order)
	return &cloned, nil
}

func (r *fakeOrderRoutingRepository) Update(_ context.Context, order entity.RoutedOrder) (*entity.RoutedOrder, error) {
	if _, ok := r.orders[order.ID]; !ok {
		return nil, fmt.Errorf("order %s not found", order.ID)
	}
	r.orders[order.ID] = cloneOrder(order)
	cloned := cloneOrder(order)
	return &cloned, nil
}

type fakeProductSetupRepository struct {
	candidates map[string]entity.ProductSetupCandidate
}

func (r *fakeProductSetupRepository) ListDrafts(context.Context) ([]entity.ProductSetupDraft, error) {
	return nil, nil
}

func (r *fakeProductSetupRepository) GetDraftByID(context.Context, string) (*entity.ProductSetupDraft, error) {
	return nil, nil
}

func (r *fakeProductSetupRepository) CreateDraft(context.Context, entity.ProductSetupDraft) (*entity.ProductSetupDraft, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *fakeProductSetupRepository) ListCandidates(context.Context) ([]entity.ProductSetupCandidate, error) {
	candidates := make([]entity.ProductSetupCandidate, 0, len(r.candidates))
	for _, candidate := range r.candidates {
		candidates = append(candidates, candidate)
	}
	return candidates, nil
}

func (r *fakeProductSetupRepository) GetCandidateByID(_ context.Context, id string) (*entity.ProductSetupCandidate, error) {
	candidate, ok := r.candidates[id]
	if !ok {
		return nil, nil
	}
	copyCandidate := candidate
	return &copyCandidate, nil
}

func (r *fakeProductSetupRepository) GetCandidateByDraftID(_ context.Context, draftID string) (*entity.ProductSetupCandidate, error) {
	for _, candidate := range r.candidates {
		if candidate.DraftID == draftID {
			copyCandidate := candidate
			return &copyCandidate, nil
		}
	}
	return nil, nil
}

func (r *fakeProductSetupRepository) CreateCandidate(context.Context, entity.ProductSetupCandidate) (*entity.ProductSetupCandidate, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *fakeProductSetupRepository) UpdateCandidateStatus(context.Context, string, string) (*entity.ProductSetupCandidate, error) {
	return nil, fmt.Errorf("not implemented")
}

func cloneOrder(order entity.RoutedOrder) entity.RoutedOrder {
	cloned := order
	cloned.Timeline = append([]string(nil), order.Timeline...)
	cloned.ActivityLog = append([]entity.RoutedOrderActivity(nil), order.ActivityLog...)
	if order.ShipmentSlaDueAt != nil {
		value := *order.ShipmentSlaDueAt
		cloned.ShipmentSlaDueAt = &value
	}
	if order.IssueSlaDueAt != nil {
		value := *order.IssueSlaDueAt
		cloned.IssueSlaDueAt = &value
	}
	if order.ShippedAt != nil {
		value := *order.ShippedAt
		cloned.ShippedAt = &value
	}
	if order.DeliveredAt != nil {
		value := *order.DeliveredAt
		cloned.DeliveredAt = &value
	}
	return cloned
}

func hasActivityDetail(details []entity.RoutedOrderActivityDetail, key, value string) bool {
	for _, detail := range details {
		if detail.Key == key && detail.Value == value {
			return true
		}
	}
	return false
}

func ptrTime(value time.Time) *time.Time {
	return &value
}

func encodeTestActivityCursor(entry entity.RoutedOrderActivityFeedEntry) string {
	return strings.Join([]string{
		entry.OrderID,
		entry.Activity.Actor,
		entry.Activity.Type,
		entry.Activity.Message,
		entry.Activity.CreatedAt.UTC().Format(time.RFC3339Nano),
	}, "|")
}
