package resolver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/internal/backoffice/controller/graphql/generated/model"
	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
	inputmocks "github.com/tuannm99/podzone/internal/backoffice/domain/inputport/mocks"
)

func TestCreateRoutedOrderMapsInputAndOutput(t *testing.T) {
	t.Parallel()

	orderUC := inputmocks.NewMockOrderRoutingUsecase(t)
	productUC := inputmocks.NewMockProductSetupUsecase(t)
	orderUC.EXPECT().
		CreateRoutedOrder(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, cmd inputport.CreateRoutedOrderCmd) (*entity.RoutedOrder, error) {
			require.Equal(t, inputport.CreateRoutedOrderCmd{
				CandidateID:      "cand-1",
				CustomerName:     "Alex POD",
				Quantity:         3,
				ProductType:      "tshirt",
				ShipRegion:       "us",
				PreferredPartner: "Print Partner A",
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
		}).
		Once()

	resolver := &mutationResolver{&Resolver{
		ProductSetupUsecase: productUC,
		OrderRoutingUsecase: orderUC,
	}}

	got, err := resolver.CreateRoutedOrder(context.Background(), model.CreateRoutedOrderInput{
		CandidateID:       "cand-1",
		CustomerName:      "Alex POD",
		Quantity:          3,
		ProductType:       "tshirt",
		ShipRegion:        "us",
		PreferredPartner:  ptrString("Print Partner A"),
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
	orderUC := inputmocks.NewMockOrderRoutingUsecase(t)
	productUC := inputmocks.NewMockProductSetupUsecase(t)
	orderUC.EXPECT().
		BulkUpdateRoutedOrders(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, cmd inputport.BulkUpdateRoutedOrdersCmd) ([]entity.RoutedOrder, error) {
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
		}).
		Once()

	resolver := &mutationResolver{&Resolver{
		ProductSetupUsecase: productUC,
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
	productUC := inputmocks.NewMockProductSetupUsecase(t)
	orderUC := inputmocks.NewMockOrderRoutingUsecase(t)
	productUC.EXPECT().GetSnapshot(mock.Anything).Return(&entity.ProductSetupSnapshot{
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
	}, nil).Once()

	resolver := &queryResolver{&Resolver{
		ProductSetupUsecase: productUC,
		OrderRoutingUsecase: orderUC,
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
	orderUC := inputmocks.NewMockOrderRoutingUsecase(t)
	productUC := inputmocks.NewMockProductSetupUsecase(t)
	orderUC.EXPECT().
		UpdateOrderShipment(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, cmd inputport.UpdateOrderShipmentCmd) (*entity.RoutedOrder, error) {
			require.Equal(t, "ord-1", cmd.OrderID)
			require.Equal(t, "bad-status", cmd.ShipmentStatus)
			return nil, wantErr
		}).
		Once()

	resolver := &mutationResolver{&Resolver{
		ProductSetupUsecase: productUC,
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
	orderUC := inputmocks.NewMockOrderRoutingUsecase(t)
	productUC := inputmocks.NewMockProductSetupUsecase(t)
	orderUC.EXPECT().
		ListRoutedOrderActivities(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, query inputport.ListRoutedOrderActivitiesQuery) (*entity.RoutedOrderActivityFeedPage, error) {
			require.Equal(t, "shipment_note", query.ActivityType)
			require.Equal(t, "user:12", query.ActorContains)
			require.Equal(t, "ord-1", query.OrderID)
			require.Equal(t, "print partner a", query.Partner)
			require.Equal(t, "ops.lead", query.Assignee)
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
						Partner:          "Print Partner A",
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
		}).
		Once()

	resolver := &queryResolver{&Resolver{
		ProductSetupUsecase: productUC,
		OrderRoutingUsecase: orderUC,
	}}

	got, err := resolver.RoutedOrderActivities(context.Background(), &model.RoutedOrderActivityFeedInput{
		ActivityType:  ptrString("shipment_note"),
		ActorContains: ptrString("user:12"),
		OrderID:       ptrString("ord-1"),
		Partner:       ptrString("print partner a"),
		Assignee:      ptrString("ops.lead"),
		Since:         &since,
		Limit:         ptrInt(25),
		After:         ptrString("cursor-1"),
		IncludeSystem: ptrBool(true),
	})
	require.NoError(t, err)
	require.Len(t, got.Entries, 1)
	require.Equal(t, "ord-1", got.Entries[0].OrderID)
	require.Equal(t, "Print Partner A", got.Entries[0].Partner)
	require.Equal(t, "DHL", got.Entries[0].Activity.Details[0].Value)
	require.Equal(t, 80, got.Total)
	require.NotNil(t, got.NextCursor)
	require.Equal(t, "cursor-2", *got.NextCursor)
}

func TestRoutedOrderRecommendationMapsQueryAndResponse(t *testing.T) {
	t.Parallel()

	orderUC := inputmocks.NewMockOrderRoutingUsecase(t)
	productUC := inputmocks.NewMockProductSetupUsecase(t)
	orderUC.EXPECT().
		RecommendRoutedOrderPartner(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, query inputport.RecommendRoutedOrderPartnerQuery) (*entity.RoutedOrderRecommendation, error) {
			require.Equal(t, inputport.RecommendRoutedOrderPartnerQuery{
				CandidateID:      "cand-1",
				ProductType:      "tshirt",
				ShipRegion:       "us",
				PreferredPartner: "Fulfill Fast",
			}, query)
			return &entity.RoutedOrderRecommendation{
				CandidateID:      "cand-1",
				ProductTitle:     "Vintage Tee",
				CandidatePartner: "Print Partner A",
				ProductType:      "tshirt",
				ShipRegion:       "us",
				SelectedPartner:  "Fulfill Fast",
				Summary:          "Preferred partner selected.",
				Options: []entity.RoutingPartnerOption{
					{
						Partner: entity.PartnerRoutingProfile{
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
						Eligible: true,
						Reason:   "preferred partner matched the routing request",
					},
				},
			}, nil
		}).
		Once()

	resolver := &queryResolver{&Resolver{
		ProductSetupUsecase: productUC,
		OrderRoutingUsecase: orderUC,
	}}

	got, err := resolver.RoutedOrderRecommendation(context.Background(), model.RoutedOrderRecommendationInput{
		CandidateID:      "cand-1",
		ProductType:      "tshirt",
		ShipRegion:       "us",
		PreferredPartner: ptrString("Fulfill Fast"),
	})
	require.NoError(t, err)
	require.Equal(t, "Fulfill Fast", got.SelectedPartner)
	require.Equal(t, "Print Partner A", got.CandidatePartner)
	require.Len(t, got.Options, 1)
	require.Equal(t, "Fulfill Fast", got.Options[0].Partner.Name)
	require.Equal(t, 90, got.Options[0].Partner.RoutingPriority)
}

func ptrString(value string) *string { return &value }
func ptrInt(value int) *int          { return &value }
func ptrBool(value bool) *bool       { return &value }
