package operations_test

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/tuannm99/podzone/internal/backoffice/application/operations"
	catalogentity "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	catalogoutputmocks "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/mocks"
	exceptionctx "github.com/tuannm99/podzone/internal/backoffice/domain/exception"
	orderctx "github.com/tuannm99/podzone/internal/backoffice/domain/order"
	orderoutputmocks "github.com/tuannm99/podzone/internal/backoffice/domain/order/mocks"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	routingoutputmocks "github.com/tuannm99/podzone/internal/backoffice/domain/routing/mocks"
	settlementctx "github.com/tuannm99/podzone/internal/backoffice/domain/settlement"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/scope"
	"github.com/tuannm99/podzone/pkg/ddd"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

const testRoutingStoreID = "store-ops"

type (
	OrderRoutingUsecase              = operations.OrderRoutingUsecase
	PartnerRoutingProfile            = routingctx.PartnerRoutingProfile
	PartnerShippingCostRule          = routingctx.PartnerShippingCostRule
	RoutedOrder                      = routingctx.RoutedOrder
	RoutedOrderActivity              = routingctx.RoutedOrderActivity
	RoutedOrderActivityDetail        = routingctx.RoutedOrderActivityDetail
	RoutedOrderActivityFeedEntry     = routingctx.RoutedOrderActivityFeedEntry
	RoutedOrderActivityFeedPage      = routingctx.RoutedOrderActivityFeedPage
	RoutedOrderActivityFeedQuery     = routingctx.RoutedOrderActivityFeedQuery
	RoutedOrderRecommendation        = routingctx.RoutedOrderRecommendation
	RoutingPartnerOption             = routingctx.RoutingPartnerOption
	RecommendRoutedOrderPartnerQuery = routingctx.RecommendRoutedOrderPartnerQuery
	BulkUpdateRoutedOrdersCmd        = operations.BulkUpdateRoutedOrdersCmd
	CreateRoutedOrderCmd             = operations.CreateRoutedOrderCmd
	ForceRerouteBlockedOrderCmd      = routingctx.ForceRerouteBlockedOrderCmd
	OpenOrderExceptionCmd            = exceptionctx.OpenOrderExceptionCmd
	UpdateOrderExceptionStatusCmd    = operations.UpdateOrderExceptionStatusCmd
	UpdateOrderIssueHandlingCmd      = settlementctx.UpdateOrderIssueHandlingCmd
	UpdateOrderQueueControlCmd       = orderctx.UpdateOrderQueueControlCmd
	UpdateOrderSettlementCmd         = settlementctx.UpdateOrderSettlementCmd
	UpdateOrderShipmentCmd           = operations.UpdateOrderShipmentCmd
)

const (
	RoutedOrderActivityTypeIssueNote      = routingctx.RoutedOrderActivityTypeIssueNote
	RoutedOrderActivityTypeSettlementNote = routingctx.RoutedOrderActivityTypeSettlementNote
	RoutedOrderActivityTypeShipmentNote   = routingctx.RoutedOrderActivityTypeShipmentNote
	RoutedOrderActivityTypeSystem         = routingctx.RoutedOrderActivityTypeSystem

	RoutedOrderExceptionStatusOpen     = routingctx.RoutedOrderExceptionStatusOpen
	RoutedOrderExceptionStatusResolved = routingctx.RoutedOrderExceptionStatusResolved

	RoutedOrderIssueResolutionMonitor = routingctx.RoutedOrderIssueResolutionMonitor
	RoutedOrderIssueResolutionReprint = routingctx.RoutedOrderIssueResolutionReprint

	RoutedOrderSettlementStatusPaid       = routingctx.RoutedOrderSettlementStatusPaid
	RoutedOrderSettlementStatusPending    = routingctx.RoutedOrderSettlementStatusPending
	RoutedOrderSettlementStatusReconciled = routingctx.RoutedOrderSettlementStatusReconciled

	RoutedOrderShipmentStatusAwaitingLabel = routingctx.RoutedOrderShipmentStatusAwaitingLabel
	RoutedOrderShipmentStatusDelivered     = routingctx.RoutedOrderShipmentStatusDelivered
	RoutedOrderShipmentStatusInTransit     = routingctx.RoutedOrderShipmentStatusInTransit

	RoutedOrderStatusInProduction   = routingctx.RoutedOrderStatusInProduction
	RoutedOrderStatusQueued         = routingctx.RoutedOrderStatusQueued
	RoutedOrderStatusRoutingBlocked = routingctx.RoutedOrderStatusRoutingBlocked
	RoutedOrderStatusShipped        = routingctx.RoutedOrderStatusShipped
)

type testOrderRoutingHarness struct {
	orders map[string]RoutedOrder
}

func newTestOrderRoutingHarness() *testOrderRoutingHarness {
	return &testOrderRoutingHarness{orders: map[string]RoutedOrder{}}
}

func (h *testOrderRoutingHarness) mustSeed(order RoutedOrder) {
	if strings.TrimSpace(order.StoreID) == "" {
		order.StoreID = testRoutingStoreID
	}
	h.orders[order.ID] = cloneOrder(order)
}

func newOrderRoutingTestInteractor(
	t *testing.T,
	candidates map[string]catalogentity.ProductSetupCandidate,
) (
	OrderRoutingUsecase,
	*testOrderRoutingHarness,
) {
	t.Helper()

	ordersMock := routingoutputmocks.NewMockOrderRoutingRepository(t)
	customerOrdersMock := orderoutputmocks.NewMockCustomerOrderQueryRepository(t)
	productsMock := catalogoutputmocks.NewMockProductSetupRepository(t)
	partnersMock := routingoutputmocks.NewMockPartnerDirectory(t)
	orderState := newTestOrderRoutingHarness()
	productState := map[string]catalogentity.ProductSetupCandidate{}
	for id, candidate := range candidates {
		if strings.TrimSpace(candidate.StoreID) == "" {
			candidate.StoreID = testRoutingStoreID
		}
		productState[id] = candidate
	}
	partnerState := []PartnerRoutingProfile{
		{
			ID:                    "prt-1",
			Code:                  "print-partner-a",
			Name:                  "Print Partner A",
			PartnerType:           "print_on_demand",
			Status:                "active",
			SupportedProductTypes: []string{"tshirt", "hoodie", "tote"},
			SupportedRegions:      []string{"us", "eu"},
			SLADays:               3,
			RoutingPriority:       100,
			BaseFulfillmentCost:   "$9.00",
			ShippingCostRules: []PartnerShippingCostRule{
				{Region: "us", Cost: "$4.00"},
				{Region: "eu", Cost: "$5.50"},
			},
		},
		{
			ID:                    "prt-2",
			Code:                  "fulfill-fast",
			Name:                  "Fulfill Fast",
			PartnerType:           "fulfillment",
			Status:                "active",
			SupportedProductTypes: []string{"poster", "tshirt"},
			SupportedRegions:      []string{"us", "uk"},
			SLADays:               2,
			RoutingPriority:       90,
			BaseFulfillmentCost:   "$7.00",
			ShippingCostRules: []PartnerShippingCostRule{
				{Region: "us", Cost: "$2.00"},
				{Region: "uk", Cost: "$3.50"},
			},
		},
	}

	ordersMock.EXPECT().
		ListByStore(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, storeID string) ([]RoutedOrder, error) {
			orders := make([]RoutedOrder, 0, len(orderState.orders))
			for _, order := range orderState.orders {
				if order.StoreID == storeID {
					orders = append(orders, cloneOrder(order))
				}
			}
			return orders, nil
		}).Maybe()

	ordersMock.EXPECT().
		ListActivityFeed(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, query RoutedOrderActivityFeedQuery) (*RoutedOrderActivityFeedPage, error) {
			entries := make([]RoutedOrderActivityFeedEntry, 0)
			for _, order := range orderState.orders {
				if query.OrderID != "" && order.ID != strings.TrimSpace(query.OrderID) {
					continue
				}
				if query.Partner != "" &&
					!strings.Contains(strings.ToLower(order.Partner), strings.ToLower(query.Partner)) {
					continue
				}
				if query.Assignee != "" &&
					!strings.Contains(strings.ToLower(order.OperatorAssignee), strings.ToLower(query.Assignee)) {
					continue
				}
				for _, activity := range order.ActivityLog {
					if query.ActivityType == "notes" && activity.Type == RoutedOrderActivityTypeSystem {
						continue
					}
					if query.ActivityType != "" && query.ActivityType != "all" && query.ActivityType != "notes" &&
						activity.Type != query.ActivityType {
						continue
					}
					if !query.IncludeSystem && query.ActivityType != "system" && query.ActivityType != "notes" &&
						activity.Type == RoutedOrderActivityTypeSystem {
						continue
					}
					if query.Since != nil && activity.CreatedAt.Before(query.Since.UTC()) {
						continue
					}
					if query.ActorContains != "" &&
						!strings.Contains(strings.ToLower(activity.Actor), strings.ToLower(query.ActorContains)) {
						continue
					}
					entries = append(entries, RoutedOrderActivityFeedEntry{
						OrderID:          order.ID,
						ProductTitle:     order.ProductTitle,
						Partner:          order.Partner,
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
			return &RoutedOrderActivityFeedPage{
				Entries:    pageEntries,
				Total:      total,
				NextCursor: nextCursor,
			}, nil
		}).
		Maybe()

	ordersMock.EXPECT().
		GetByID(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, id string) (*RoutedOrder, error) {
			order, ok := orderState.orders[id]
			if !ok {
				return nil, nil
			}
			cloned := cloneOrder(order)
			return &cloned, nil
		}).
		Maybe()

	ordersMock.EXPECT().
		Create(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, order RoutedOrder) (*RoutedOrder, error) {
			orderState.orders[order.ID] = cloneOrder(order)
			cloned := cloneOrder(order)
			return &cloned, nil
		}).
		Maybe()

	ordersMock.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, order RoutedOrder) (*RoutedOrder, error) {
			if _, ok := orderState.orders[order.ID]; !ok {
				return nil, fmt.Errorf("order %s not found", order.ID)
			}
			orderState.orders[order.ID] = cloneOrder(order)
			cloned := cloneOrder(order)
			return &cloned, nil
		}).
		Maybe()

	customerOrdersMock.EXPECT().
		GetCustomerOrder(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, storeID string, id string) (*orderctx.CustomerOrder, error) {
			order, ok := orderState.orders[id]
			if !ok || order.StoreID != storeID {
				return nil, ddd.ErrNotFound
			}
			return rehydrateTestCustomerOrder(order)
		}).
		Maybe()

	productsMock.EXPECT().
		GetCandidateByID(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, storeID string, id string) (*catalogentity.ProductSetupCandidate, error) {
			candidate, ok := productState[id]
			if !ok || candidate.StoreID != storeID {
				return nil, nil
			}
			copyCandidate := candidate
			return &copyCandidate, nil
		}).
		Maybe()

	productsMock.EXPECT().
		GetCandidateByDraftID(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, storeID string, draftID string) (*catalogentity.ProductSetupCandidate, error) {
			for _, candidate := range productState {
				if candidate.StoreID == storeID && candidate.DraftID == draftID {
					copyCandidate := candidate
					return &copyCandidate, nil
				}
			}
			return nil, nil
		}).
		Maybe()

	productsMock.EXPECT().
		ListCandidates(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, storeID string) ([]catalogentity.ProductSetupCandidate, error) {
			candidates := make([]catalogentity.ProductSetupCandidate, 0, len(productState))
			for _, candidate := range productState {
				if candidate.StoreID == storeID {
					candidates = append(candidates, candidate)
				}
			}
			return candidates, nil
		}).
		Maybe()

	productsMock.EXPECT().
		ListDrafts(mock.Anything, mock.Anything).
		Return([]catalogentity.ProductSetupDraft(nil), nil).
		Maybe()
	productsMock.EXPECT().
		GetDraftByID(mock.Anything, mock.Anything, mock.Anything).
		Return((*catalogentity.ProductSetupDraft)(nil), nil).
		Maybe()
	productsMock.EXPECT().
		CreateDraft(mock.Anything, mock.Anything).
		Return((*catalogentity.ProductSetupDraft)(nil), fmt.Errorf("not implemented")).
		Maybe()
	productsMock.EXPECT().
		CreateCandidate(mock.Anything, mock.Anything).
		Return((*catalogentity.ProductSetupCandidate)(nil), fmt.Errorf("not implemented")).
		Maybe()
	productsMock.EXPECT().
		UpdateCandidateStatus(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return((*catalogentity.ProductSetupCandidate)(nil), fmt.Errorf("not implemented")).
		Maybe()

	partnersMock.EXPECT().
		ListActivePartners(mock.Anything, mock.Anything).
		RunAndReturn(func(context.Context, string) ([]PartnerRoutingProfile, error) {
			return append([]PartnerRoutingProfile(nil), partnerState...), nil
		}).
		Maybe()

	interactor := operations.NewOrderRoutingInteractor(
		ordersMock,
		customerOrdersMock,
		productsMock,
		partnersMock,
		ddd.EventDispatcher(nil),
		ddd.NewUUIDGenerator(),
		ddd.NewFixedClock(time.Date(2026, 6, 4, 10, 30, 0, 0, time.UTC)),
	)
	return interactor, orderState
}

func rehydrateTestCustomerOrder(order RoutedOrder) (*orderctx.CustomerOrder, error) {
	return orderctx.RehydrateCustomerOrder(orderctx.CustomerOrderSnapshot{
		ID:                 order.ID,
		Version:            order.AggregateVersion,
		StoreID:            order.StoreID,
		CandidateID:        order.CandidateID,
		ProductTitle:       order.ProductTitle,
		Quantity:           order.Quantity,
		Total:              order.Total,
		CustomerName:       order.CustomerName,
		Status:             order.Status,
		Partner:            order.Partner,
		OperatorAssignee:   order.OperatorAssignee,
		ShipmentSlaDueAt:   order.ShipmentSlaDueAt,
		IssueSlaDueAt:      order.IssueSlaDueAt,
		ExceptionStatus:    order.ExceptionStatus,
		RoutingBlockCode:   order.RoutingBlockCode,
		RoutingBlockReason: order.RoutingBlockReason,
		SettlementStatus:   order.SettlementStatus,
		UpdatedAt:          order.UpdatedAt,
	})
}

func cloneOrder(order RoutedOrder) RoutedOrder {
	cloned := order
	cloned.Timeline = append([]string(nil), order.Timeline...)
	cloned.ActivityLog = append([]RoutedOrderActivity(nil), order.ActivityLog...)
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

func testRoutingContext() context.Context {
	ctx := context.Background()
	ctx = scope.WithStoreContext(ctx, scope.StoreContext{StoreID: testRoutingStoreID})
	return ctx
}

func testTenantRoutingContext() context.Context {
	return toolkit.WithTenantID(testRoutingContext(), "t_demo")
}

func hasActivityDetail(details []RoutedOrderActivityDetail, key, value string) bool {
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

func encodeTestActivityCursor(entry RoutedOrderActivityFeedEntry) string {
	return strings.Join([]string{
		entry.OrderID,
		entry.Activity.Actor,
		entry.Activity.Type,
		entry.Activity.Message,
		entry.Activity.CreatedAt.UTC().Format(time.RFC3339Nano),
	}, "|")
}
