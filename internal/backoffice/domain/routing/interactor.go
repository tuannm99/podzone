package routing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	catalogentity "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/entity"
	catalogoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/outputport"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routinginputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/inputport"
	routingoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/outputport"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type OrderRoutingInteractor struct {
	orders   routingoutputport.OrderRoutingRepository
	products catalogoutputport.ProductSetupRepository
	partners routingoutputport.PartnerDirectory
}

var _ routinginputport.OrderRoutingUsecase = (*OrderRoutingInteractor)(nil)

func NewOrderRoutingInteractor(
	orders routingoutputport.OrderRoutingRepository,
	products catalogoutputport.ProductSetupRepository,
	partners routingoutputport.PartnerDirectory,
) routinginputport.OrderRoutingUsecase {
	return &OrderRoutingInteractor{orders: orders, products: products, partners: partners}
}

func (i *OrderRoutingInteractor) ListRoutedOrders(ctx context.Context) ([]routingentity.RoutedOrder, error) {
	return i.orders.List(ctx)
}

func (i *OrderRoutingInteractor) ListRoutedOrderActivities(
	ctx context.Context,
	query routingentity.RoutedOrderActivityFeedQuery,
) (*routingentity.RoutedOrderActivityFeedPage, error) {
	return i.orders.ListActivityFeed(ctx, query)
}

func (i *OrderRoutingInteractor) RecommendRoutedOrderPartner(
	ctx context.Context,
	query routinginputport.RecommendRoutedOrderPartnerQuery,
) (*routingentity.RoutedOrderRecommendation, error) {
	candidateID := strings.TrimSpace(query.CandidateID)
	if candidateID == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	candidate, err := i.products.GetCandidateByID(ctx, candidateID)
	if err != nil {
		return nil, err
	}
	if candidate == nil {
		return nil, fmt.Errorf("product candidate not found")
	}
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}
	partners, err := i.partners.ListActivePartners(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return buildRoutingRecommendation(
		candidate,
		partners,
		normalizeRoutingLabel(query.ProductType),
		normalizeRoutingLabel(query.ShipRegion),
		strings.TrimSpace(query.PreferredPartner),
	), nil
}

func (i *OrderRoutingInteractor) CreateRoutedOrder(
	ctx context.Context,
	cmd routinginputport.CreateRoutedOrderCmd,
) (*routingentity.RoutedOrder, error) {
	candidateID := strings.TrimSpace(cmd.CandidateID)
	if candidateID == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	candidate, err := i.products.GetCandidateByID(ctx, candidateID)
	if err != nil {
		return nil, err
	}
	if candidate == nil || candidate.Status != catalogentity.ProductSetupCandidateStatusPublishedMock {
		return nil, fmt.Errorf("published mock product candidate is required")
	}
	tenantID, err := toolkit.GetTenantID(ctx)
	if err != nil {
		return nil, err
	}
	partners, err := i.partners.ListActivePartners(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	recommendation := buildRoutingRecommendation(
		candidate,
		partners,
		normalizeRoutingLabel(cmd.ProductType),
		normalizeRoutingLabel(cmd.ShipRegion),
		strings.TrimSpace(cmd.PreferredPartner),
	)
	if recommendation.SelectedPartner == "" {
		return nil, fmt.Errorf(
			"no eligible active partner found for %s in %s",
			fallbackRoutingLabel(recommendation.ProductType, "product type"),
			fallbackRoutingLabel(recommendation.ShipRegion, "target region"),
		)
	}

	qty := cmd.Quantity
	if qty < 1 {
		qty = 1
	}
	customerName := strings.TrimSpace(cmd.CustomerName)
	if customerName == "" {
		customerName = "Sample customer"
	}
	now := time.Now().UTC()
	actor := activityActorFromContext(ctx)
	order := routingentity.RoutedOrder{
		ID:               fmt.Sprintf("ORD-%s", strings.ToUpper(uuid.NewString()[:8])),
		CandidateID:      candidate.ID,
		ProductTitle:     candidate.Title,
		Partner:          recommendation.SelectedPartner,
		Quantity:         qty,
		Total:            multiplyMoney(candidate.RetailPrice, qty),
		CustomerName:     customerName,
		Status:           routingentity.RoutedOrderStatusQueued,
		ShipmentStatus:   routingentity.RoutedOrderShipmentStatusAwaitingLabel,
		OperatorAssignee: "unassigned",
		BaseCostSnapshot: multiplyMoney(candidate.BaseCost, qty),
		FulfillmentCost:  multiplyMoney(candidate.BaseCost, qty),
		ShippingCost:     "$0.00",
		IssueCost:        "$0.00",
		IssueResolution:  routingentity.RoutedOrderIssueResolutionMonitor,
		SettlementStatus: routingentity.RoutedOrderSettlementStatusPending,
		Timeline: []string{
			fmt.Sprintf("Order created for %s", candidate.Title),
			routingTimelineEntry(routingentity.RoutedOrderStatusQueued, candidate.Partner),
		},
		ActivityLog: []routingentity.RoutedOrderActivity{
			newActivity(
				routingentity.RoutedOrderActivityTypeSystem,
				actor,
				fmt.Sprintf("Order created for %s", candidate.Title),
				now,
				activityDetails(
					"candidate_id", candidate.ID,
					"quantity", fmt.Sprintf("%d", qty),
					"status", routingentity.RoutedOrderStatusQueued,
					"product_type", recommendation.ProductType,
					"ship_region", recommendation.ShipRegion,
				),
			),
			newActivity(
				routingentity.RoutedOrderActivityTypeSystem,
				actor,
				routingTimelineEntry(routingentity.RoutedOrderStatusQueued, recommendation.SelectedPartner),
				now,
				activityDetails(
					"status", routingentity.RoutedOrderStatusQueued,
					"partner", recommendation.SelectedPartner,
					"routing_summary", recommendation.Summary,
					"candidate_partner", recommendation.CandidatePartner,
				),
			),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	order.RealizedMargin = calculateMargin(order.Total, order.FulfillmentCost, order.ShippingCost, order.IssueCost)
	return i.orders.Create(ctx, order)
}
