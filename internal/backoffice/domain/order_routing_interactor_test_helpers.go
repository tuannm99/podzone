package interactor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
	outputmocks "github.com/tuannm99/podzone/internal/backoffice/domain/outputport/mocks"
)

type testOrderRoutingHarness struct {
	orders map[string]entity.RoutedOrder
}

func newTestOrderRoutingHarness() *testOrderRoutingHarness {
	return &testOrderRoutingHarness{orders: map[string]entity.RoutedOrder{}}
}

func (h *testOrderRoutingHarness) mustSeed(order entity.RoutedOrder) {
	h.orders[order.ID] = cloneOrder(order)
}

func newOrderRoutingTestInteractor(
	t *testing.T,
	candidates map[string]entity.ProductSetupCandidate,
) (*OrderRoutingInteractor, *testOrderRoutingHarness, *outputmocks.MockProductSetupRepository, *outputmocks.MockPartnerDirectory) {
	t.Helper()

	ordersMock := outputmocks.NewMockOrderRoutingRepository(t)
	productsMock := outputmocks.NewMockProductSetupRepository(t)
	partnersMock := outputmocks.NewMockPartnerDirectory(t)
	orderState := newTestOrderRoutingHarness()
	productState := map[string]entity.ProductSetupCandidate{}
	for id, candidate := range candidates {
		productState[id] = candidate
	}
	partnerState := []entity.PartnerRoutingProfile{
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
		},
	}

	ordersMock.EXPECT().List(mock.Anything).RunAndReturn(func(context.Context) ([]entity.RoutedOrder, error) {
		orders := make([]entity.RoutedOrder, 0, len(orderState.orders))
		for _, order := range orderState.orders {
			orders = append(orders, cloneOrder(order))
		}
		return orders, nil
	}).Maybe()

	ordersMock.EXPECT().
		ListActivityFeed(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, query inputport.ListRoutedOrderActivitiesQuery) (*entity.RoutedOrderActivityFeedPage, error) {
			entries := make([]entity.RoutedOrderActivityFeedEntry, 0)
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
					if query.ActivityType == "notes" && activity.Type == entity.RoutedOrderActivityTypeSystem {
						continue
					}
					if query.ActivityType != "" && query.ActivityType != "all" && query.ActivityType != "notes" &&
						activity.Type != query.ActivityType {
						continue
					}
					if !query.IncludeSystem && query.ActivityType != "system" && query.ActivityType != "notes" &&
						activity.Type == entity.RoutedOrderActivityTypeSystem {
						continue
					}
					if query.Since != nil && activity.CreatedAt.Before(query.Since.UTC()) {
						continue
					}
					if query.ActorContains != "" &&
						!strings.Contains(strings.ToLower(activity.Actor), strings.ToLower(query.ActorContains)) {
						continue
					}
					entries = append(entries, entity.RoutedOrderActivityFeedEntry{
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
			return &entity.RoutedOrderActivityFeedPage{
				Entries:    pageEntries,
				Total:      total,
				NextCursor: nextCursor,
			}, nil
		}).
		Maybe()

	ordersMock.EXPECT().
		GetByID(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, id string) (*entity.RoutedOrder, error) {
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
		RunAndReturn(func(_ context.Context, order entity.RoutedOrder) (*entity.RoutedOrder, error) {
			orderState.orders[order.ID] = cloneOrder(order)
			cloned := cloneOrder(order)
			return &cloned, nil
		}).
		Maybe()

	ordersMock.EXPECT().
		Update(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, order entity.RoutedOrder) (*entity.RoutedOrder, error) {
			if _, ok := orderState.orders[order.ID]; !ok {
				return nil, fmt.Errorf("order %s not found", order.ID)
			}
			orderState.orders[order.ID] = cloneOrder(order)
			cloned := cloneOrder(order)
			return &cloned, nil
		}).
		Maybe()

	productsMock.EXPECT().
		GetCandidateByID(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, id string) (*entity.ProductSetupCandidate, error) {
			candidate, ok := productState[id]
			if !ok {
				return nil, nil
			}
			copyCandidate := candidate
			return &copyCandidate, nil
		}).
		Maybe()

	productsMock.EXPECT().
		GetCandidateByDraftID(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, draftID string) (*entity.ProductSetupCandidate, error) {
			for _, candidate := range productState {
				if candidate.DraftID == draftID {
					copyCandidate := candidate
					return &copyCandidate, nil
				}
			}
			return nil, nil
		}).
		Maybe()

	productsMock.EXPECT().
		ListCandidates(mock.Anything).
		RunAndReturn(func(context.Context) ([]entity.ProductSetupCandidate, error) {
			candidates := make([]entity.ProductSetupCandidate, 0, len(productState))
			for _, candidate := range productState {
				candidates = append(candidates, candidate)
			}
			return candidates, nil
		}).
		Maybe()

	productsMock.EXPECT().ListDrafts(mock.Anything).Return([]entity.ProductSetupDraft(nil), nil).Maybe()
	productsMock.EXPECT().
		GetDraftByID(mock.Anything, mock.Anything).
		Return((*entity.ProductSetupDraft)(nil), nil).
		Maybe()
	productsMock.EXPECT().
		CreateDraft(mock.Anything, mock.Anything).
		Return((*entity.ProductSetupDraft)(nil), fmt.Errorf("not implemented")).
		Maybe()
	productsMock.EXPECT().
		CreateCandidate(mock.Anything, mock.Anything).
		Return((*entity.ProductSetupCandidate)(nil), fmt.Errorf("not implemented")).
		Maybe()
	productsMock.EXPECT().
		UpdateCandidateStatus(mock.Anything, mock.Anything, mock.Anything).
		Return((*entity.ProductSetupCandidate)(nil), fmt.Errorf("not implemented")).
		Maybe()

	partnersMock.EXPECT().
		ListActivePartners(mock.Anything, mock.Anything).
		RunAndReturn(func(context.Context, string) ([]entity.PartnerRoutingProfile, error) {
			return append([]entity.PartnerRoutingProfile(nil), partnerState...), nil
		}).
		Maybe()

	return &OrderRoutingInteractor{orders: ordersMock, products: productsMock, partners: partnersMock}, orderState, productsMock, partnersMock
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
