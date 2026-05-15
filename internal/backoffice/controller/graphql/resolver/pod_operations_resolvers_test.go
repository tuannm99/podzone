package resolver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated/model"
	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
)

func TestCreateRoutedOrderMapsInputAndOutput(t *testing.T) {
	t.Parallel()

	orderUC := &fakeOrderRoutingUsecase{
		createRoutedOrderFn: func(_ context.Context, cmd inputport.CreateRoutedOrderCmd) (*entity.RoutedOrder, error) {
			require.Equal(t, inputport.CreateRoutedOrderCmd{
				CandidateID:  "cand-1",
				CustomerName: "Alex POD",
				Quantity:     3,
			}, cmd)
			now := time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC)
			return &entity.RoutedOrder{
				ID:               "ord-1",
				CandidateID:      cmd.CandidateID,
				ProductTitle:     "Vintage Tee",
				Partner:          "Print Partner A",
				Quantity:         cmd.Quantity,
				Total:            "$60.00",
				CustomerName:     cmd.CustomerName,
				Status:           entity.RoutedOrderStatusQueued,
				Timeline:         []string{"created"},
				ShipmentStatus:   entity.RoutedOrderShipmentStatusAwaitingLabel,
				OperatorAssignee: "unassigned",
				BaseCostSnapshot: "$24.00",
				FulfillmentCost:  "$24.00",
				ShippingCost:     "$0.00",
				IssueCost:        "$0.00",
				IssueResolution:  entity.RoutedOrderIssueResolutionMonitor,
				RealizedMargin:   "$36.00",
				SettlementStatus: entity.RoutedOrderSettlementStatusPending,
				CreatedAt:        now,
				UpdatedAt:        now,
			}, nil
		},
	}

	resolver := &mutationResolver{&Resolver{
		ProductSetupUsecase: &fakeProductSetupUsecase{},
		OrderRoutingUsecase: orderUC,
	}}

	got, err := resolver.CreateRoutedOrder(context.Background(), model.CreateRoutedOrderInput{
		CandidateID:  "cand-1",
		CustomerName: "Alex POD",
		Quantity:     3,
	})
	require.NoError(t, err)
	require.Equal(t, "ord-1", got.ID)
	require.Equal(t, "cand-1", got.CandidateID)
	require.Equal(t, "Vintage Tee", got.ProductTitle)
	require.Equal(t, "$36.00", got.RealizedMargin)
	require.Equal(t, entity.RoutedOrderSettlementStatusPending, got.SettlementStatus)
}

func TestBulkUpdateRoutedOrdersMapsPointersAndList(t *testing.T) {
	t.Parallel()

	sla := time.Date(2026, 5, 15, 18, 30, 0, 0, time.UTC)
	owner := "ops.lead"
	status := entity.RoutedOrderSettlementStatusPaid
	orderUC := &fakeOrderRoutingUsecase{
		bulkUpdateRoutedOrdersFn: func(_ context.Context, cmd inputport.BulkUpdateRoutedOrdersCmd) ([]entity.RoutedOrder, error) {
			require.Equal(t, []string{"ord-1", "ord-2"}, cmd.OrderIDs)
			require.NotNil(t, cmd.OperatorAssignee)
			require.Equal(t, owner, *cmd.OperatorAssignee)
			require.NotNil(t, cmd.ShipmentSlaDueAt)
			require.True(t, cmd.ShipmentSlaDueAt.Equal(sla))
			require.NotNil(t, cmd.SettlementStatus)
			require.Equal(t, status, *cmd.SettlementStatus)
			return []entity.RoutedOrder{
				{ID: "ord-1", OperatorAssignee: owner, ShipmentSlaDueAt: &sla, SettlementStatus: status},
				{ID: "ord-2", OperatorAssignee: owner, ShipmentSlaDueAt: &sla, SettlementStatus: status},
			}, nil
		},
	}

	resolver := &mutationResolver{&Resolver{
		ProductSetupUsecase: &fakeProductSetupUsecase{},
		OrderRoutingUsecase: orderUC,
	}}

	got, err := resolver.BulkUpdateRoutedOrders(context.Background(), model.BulkUpdateRoutedOrdersInput{
		OrderIds:         []string{"ord-1", "ord-2"},
		OperatorAssignee: &owner,
		ShipmentSLADueAt: &sla,
		SettlementStatus: &status,
	})
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "ord-1", got[0].ID)
	require.Equal(t, owner, got[0].OperatorAssignee)
	require.NotNil(t, got[0].ShipmentSLADueAt)
	require.True(t, got[0].ShipmentSLADueAt.Equal(sla))
	require.Equal(t, status, got[0].SettlementStatus)
}

func TestProductSetupSnapshotMapsNestedGraphQLFields(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 15, 9, 0, 0, 0, time.UTC)
	productUC := &fakeProductSetupUsecase{
		getSnapshotFn: func(context.Context) (*entity.ProductSetupSnapshot, error) {
			return &entity.ProductSetupSnapshot{
				Drafts: []entity.ProductSetupDraft{
					{
						ID:          "draft-1",
						Name:        "Vintage Tee Draft",
						Partner:     "Print Partner A",
						BaseCost:    "$8.00",
						RetailPrice: "$20.00",
						Status:      entity.ProductSetupDraftStatusReadyForReview,
						Notes:       "ready",
						CreatedAt:   now,
						UpdatedAt:   now,
					},
				},
				Candidates: []entity.ProductSetupCandidate{
					{
						ID:              "cand-1",
						DraftID:         "draft-1",
						Title:           "Vintage Tee",
						SKU:             "TEE-001",
						Partner:         "Print Partner A",
						BaseCost:        "$8.00",
						RetailPrice:     "$20.00",
						EstimatedMargin: "$12.00",
						Status:          entity.ProductSetupCandidateStatusPublishedMock,
						Channel:         "shopify",
						UpdatedAt:       now,
						Variants: []entity.ProductSetupVariant{
							{ID: "v1", Label: "Black / M", Color: "Black", Size: "M", Status: "ready"},
						},
						ArtworkChecklist: entity.ProductSetupArtworkChecklist{
							FrontArtwork:     true,
							BackArtwork:      false,
							MockupReady:      true,
							PrintSpecChecked: true,
						},
						MerchandisingNotes: "launch ready",
					},
				},
			}, nil
		},
	}

	resolver := &queryResolver{&Resolver{
		ProductSetupUsecase: productUC,
		OrderRoutingUsecase: &fakeOrderRoutingUsecase{},
	}}

	got, err := resolver.ProductSetupSnapshot(context.Background())
	require.NoError(t, err)
	require.Len(t, got.Drafts, 1)
	require.Len(t, got.Candidates, 1)
	require.Equal(t, "draft-1", got.Drafts[0].ID)
	require.Equal(t, "TEE-001", got.Candidates[0].Sku)
	require.NotNil(t, got.Candidates[0].ArtworkChecklist)
	require.True(t, got.Candidates[0].ArtworkChecklist.FrontArtwork)
	require.Equal(t, "Black", got.Candidates[0].Variants[0].Color)
}

func TestUpdateOrderShipmentPropagatesUsecaseErrors(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("invalid shipment status")
	orderUC := &fakeOrderRoutingUsecase{
		updateOrderShipmentFn: func(_ context.Context, cmd inputport.UpdateOrderShipmentCmd) (*entity.RoutedOrder, error) {
			require.Equal(t, "ord-1", cmd.OrderID)
			require.Equal(t, "bad-status", cmd.ShipmentStatus)
			return nil, wantErr
		},
	}

	resolver := &mutationResolver{&Resolver{
		ProductSetupUsecase: &fakeProductSetupUsecase{},
		OrderRoutingUsecase: orderUC,
	}}

	got, err := resolver.UpdateOrderShipment(context.Background(), model.UpdateOrderShipmentInput{
		OrderID:        "ord-1",
		ShipmentStatus: "bad-status",
		Carrier:        "DHL",
		TrackingNumber: "TRACK-1",
		TrackingURL:    "https://tracking.example/TRACK-1",
		Notes:          "bad input",
	})
	require.ErrorIs(t, err, wantErr)
	require.Nil(t, got)
}

func TestRoutedOrderActivitiesMapsQueryAndResponse(t *testing.T) {
	t.Parallel()

	since := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	orderUC := &fakeOrderRoutingUsecase{
		listRoutedOrderActivitiesFn: func(_ context.Context, query inputport.ListRoutedOrderActivitiesQuery) (*entity.RoutedOrderActivityFeedPage, error) {
			require.Equal(t, "shipment_note", query.ActivityType)
			require.Equal(t, "user:12", query.ActorContains)
			require.NotNil(t, query.Since)
			require.True(t, query.Since.Equal(since))
			require.Equal(t, 25, query.Limit)
			require.Equal(t, "cursor-1", query.After)
			require.True(t, query.IncludeSystem)
			nextCursor := "cursor-2"
			return &entity.RoutedOrderActivityFeedPage{
				Entries: []entity.RoutedOrderActivityFeedEntry{
					{
						OrderID:          "ord-1",
						ProductTitle:     "Vintage Tee",
						OperatorAssignee: "ops.lead",
						Activity: entity.RoutedOrderActivity{
							Type:      entity.RoutedOrderActivityTypeShipmentNote,
							Actor:     "user:12",
							Message:   "Handed off to carrier",
							Details:   []entity.RoutedOrderActivityDetail{{Key: "carrier", Value: "DHL"}},
							CreatedAt: since,
						},
					},
				},
				Total:      80,
				NextCursor: &nextCursor,
			}, nil
		},
	}

	resolver := &queryResolver{&Resolver{
		ProductSetupUsecase: &fakeProductSetupUsecase{},
		OrderRoutingUsecase: orderUC,
	}}

	got, err := resolver.RoutedOrderActivities(context.Background(), &model.RoutedOrderActivityFeedInput{
		ActivityType:  ptrString("shipment_note"),
		ActorContains: ptrString("user:12"),
		Since:         &since,
		Limit:         ptrInt(25),
		After:         ptrString("cursor-1"),
		IncludeSystem: ptrBool(true),
	})
	require.NoError(t, err)
	require.Len(t, got.Entries, 1)
	require.Equal(t, "ord-1", got.Entries[0].OrderID)
	require.Equal(t, "DHL", got.Entries[0].Activity.Details[0].Value)
	require.Equal(t, 80, got.Total)
	require.NotNil(t, got.NextCursor)
	require.Equal(t, "cursor-2", *got.NextCursor)
}

type fakeProductSetupUsecase struct {
	getSnapshotFn           func(ctx context.Context) (*entity.ProductSetupSnapshot, error)
	createDraftFn           func(ctx context.Context, cmd inputport.CreateProductSetupDraftCmd) (*entity.ProductSetupDraft, error)
	promoteCandidateFn      func(ctx context.Context, cmd inputport.PromoteProductSetupCandidateCmd) (*entity.ProductSetupCandidate, error)
	updateCandidateStatusFn func(ctx context.Context, id, status string) (*entity.ProductSetupCandidate, error)
}

func (f *fakeProductSetupUsecase) GetSnapshot(ctx context.Context) (*entity.ProductSetupSnapshot, error) {
	if f.getSnapshotFn != nil {
		return f.getSnapshotFn(ctx)
	}
	return &entity.ProductSetupSnapshot{}, nil
}

func (f *fakeProductSetupUsecase) CreateDraft(ctx context.Context, cmd inputport.CreateProductSetupDraftCmd) (*entity.ProductSetupDraft, error) {
	if f.createDraftFn != nil {
		return f.createDraftFn(ctx, cmd)
	}
	return nil, errors.New("unexpected CreateDraft call")
}

func (f *fakeProductSetupUsecase) PromoteCandidate(ctx context.Context, cmd inputport.PromoteProductSetupCandidateCmd) (*entity.ProductSetupCandidate, error) {
	if f.promoteCandidateFn != nil {
		return f.promoteCandidateFn(ctx, cmd)
	}
	return nil, errors.New("unexpected PromoteCandidate call")
}

func (f *fakeProductSetupUsecase) UpdateCandidateStatus(ctx context.Context, id, status string) (*entity.ProductSetupCandidate, error) {
	if f.updateCandidateStatusFn != nil {
		return f.updateCandidateStatusFn(ctx, id, status)
	}
	return nil, errors.New("unexpected UpdateCandidateStatus call")
}

type fakeOrderRoutingUsecase struct {
	listRoutedOrdersFn          func(ctx context.Context) ([]entity.RoutedOrder, error)
	listRoutedOrderActivitiesFn func(ctx context.Context, query inputport.ListRoutedOrderActivitiesQuery) (*entity.RoutedOrderActivityFeedPage, error)
	createRoutedOrderFn         func(ctx context.Context, cmd inputport.CreateRoutedOrderCmd) (*entity.RoutedOrder, error)
	advanceRoutedOrderFn        func(ctx context.Context, orderID string) (*entity.RoutedOrder, error)
	openOrderExceptionFn        func(ctx context.Context, cmd inputport.OpenOrderExceptionCmd) (*entity.RoutedOrder, error)
	updateExceptionStatusFn     func(ctx context.Context, cmd inputport.UpdateOrderExceptionStatusCmd) (*entity.RoutedOrder, error)
	updateOrderShipmentFn       func(ctx context.Context, cmd inputport.UpdateOrderShipmentCmd) (*entity.RoutedOrder, error)
	updateOrderSettlementFn     func(ctx context.Context, cmd inputport.UpdateOrderSettlementCmd) (*entity.RoutedOrder, error)
	updateOrderIssueHandlingFn  func(ctx context.Context, cmd inputport.UpdateOrderIssueHandlingCmd) (*entity.RoutedOrder, error)
	updateOrderQueueControlFn   func(ctx context.Context, cmd inputport.UpdateOrderQueueControlCmd) (*entity.RoutedOrder, error)
	bulkUpdateRoutedOrdersFn    func(ctx context.Context, cmd inputport.BulkUpdateRoutedOrdersCmd) ([]entity.RoutedOrder, error)
}

func (f *fakeOrderRoutingUsecase) ListRoutedOrders(ctx context.Context) ([]entity.RoutedOrder, error) {
	if f.listRoutedOrdersFn != nil {
		return f.listRoutedOrdersFn(ctx)
	}
	return nil, nil
}

func (f *fakeOrderRoutingUsecase) ListRoutedOrderActivities(ctx context.Context, query inputport.ListRoutedOrderActivitiesQuery) (*entity.RoutedOrderActivityFeedPage, error) {
	if f.listRoutedOrderActivitiesFn != nil {
		return f.listRoutedOrderActivitiesFn(ctx, query)
	}
	return nil, errors.New("unexpected ListRoutedOrderActivities call")
}

func (f *fakeOrderRoutingUsecase) CreateRoutedOrder(ctx context.Context, cmd inputport.CreateRoutedOrderCmd) (*entity.RoutedOrder, error) {
	if f.createRoutedOrderFn != nil {
		return f.createRoutedOrderFn(ctx, cmd)
	}
	return nil, errors.New("unexpected CreateRoutedOrder call")
}

func (f *fakeOrderRoutingUsecase) AdvanceRoutedOrder(ctx context.Context, orderID string) (*entity.RoutedOrder, error) {
	if f.advanceRoutedOrderFn != nil {
		return f.advanceRoutedOrderFn(ctx, orderID)
	}
	return nil, errors.New("unexpected AdvanceRoutedOrder call")
}

func (f *fakeOrderRoutingUsecase) OpenOrderException(ctx context.Context, cmd inputport.OpenOrderExceptionCmd) (*entity.RoutedOrder, error) {
	if f.openOrderExceptionFn != nil {
		return f.openOrderExceptionFn(ctx, cmd)
	}
	return nil, errors.New("unexpected OpenOrderException call")
}

func (f *fakeOrderRoutingUsecase) UpdateOrderExceptionStatus(ctx context.Context, cmd inputport.UpdateOrderExceptionStatusCmd) (*entity.RoutedOrder, error) {
	if f.updateExceptionStatusFn != nil {
		return f.updateExceptionStatusFn(ctx, cmd)
	}
	return nil, errors.New("unexpected UpdateOrderExceptionStatus call")
}

func (f *fakeOrderRoutingUsecase) UpdateOrderShipment(ctx context.Context, cmd inputport.UpdateOrderShipmentCmd) (*entity.RoutedOrder, error) {
	if f.updateOrderShipmentFn != nil {
		return f.updateOrderShipmentFn(ctx, cmd)
	}
	return nil, errors.New("unexpected UpdateOrderShipment call")
}

func (f *fakeOrderRoutingUsecase) UpdateOrderSettlement(ctx context.Context, cmd inputport.UpdateOrderSettlementCmd) (*entity.RoutedOrder, error) {
	if f.updateOrderSettlementFn != nil {
		return f.updateOrderSettlementFn(ctx, cmd)
	}
	return nil, errors.New("unexpected UpdateOrderSettlement call")
}

func (f *fakeOrderRoutingUsecase) UpdateOrderIssueHandling(ctx context.Context, cmd inputport.UpdateOrderIssueHandlingCmd) (*entity.RoutedOrder, error) {
	if f.updateOrderIssueHandlingFn != nil {
		return f.updateOrderIssueHandlingFn(ctx, cmd)
	}
	return nil, errors.New("unexpected UpdateOrderIssueHandling call")
}

func (f *fakeOrderRoutingUsecase) UpdateOrderQueueControl(ctx context.Context, cmd inputport.UpdateOrderQueueControlCmd) (*entity.RoutedOrder, error) {
	if f.updateOrderQueueControlFn != nil {
		return f.updateOrderQueueControlFn(ctx, cmd)
	}
	return nil, errors.New("unexpected UpdateOrderQueueControl call")
}

func (f *fakeOrderRoutingUsecase) BulkUpdateRoutedOrders(ctx context.Context, cmd inputport.BulkUpdateRoutedOrdersCmd) ([]entity.RoutedOrder, error) {
	if f.bulkUpdateRoutedOrdersFn != nil {
		return f.bulkUpdateRoutedOrdersFn(ctx, cmd)
	}
	return nil, errors.New("unexpected BulkUpdateRoutedOrders call")
}

func ptrString(value string) *string {
	return &value
}

func ptrInt(value int) *int {
	return &value
}

func ptrBool(value bool) *bool {
	return &value
}
