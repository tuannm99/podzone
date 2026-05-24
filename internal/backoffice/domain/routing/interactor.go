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
	selectedOption := findSelectedRoutingOption(recommendation)

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
	status := routingentity.RoutedOrderStatusQueued
	partner := recommendation.SelectedPartner
	timelineEntry := routingTimelineEntry(routingentity.RoutedOrderStatusQueued, recommendation.SelectedPartner)
	routingBlockCode := ""
	routingBlockReason := ""
	fulfillmentCost := "TBD"
	shippingCost := "TBD"
	estimatedMargin := "TBD"
	if selectedOption != nil {
		fulfillmentCost = multiplyMoney(selectedOption.EstimatedFulfillmentCost, qty)
		shippingCost = selectedOption.EstimatedShippingCost
		estimatedMargin = calculateMargin(
			multiplyMoney(candidate.RetailPrice, qty),
			fulfillmentCost,
			shippingCost,
			"$0.00",
		)
	}
	if recommendation.SelectedPartner == "" {
		status = routingentity.RoutedOrderStatusRoutingBlocked
		partner = ""
		timelineEntry = fmt.Sprintf("Routing blocked: %s", recommendation.BlockedReason)
		routingBlockCode = recommendation.BlockedReasonCode
		routingBlockReason = recommendation.BlockedReason
	}
	order := routingentity.RoutedOrder{
		ID:                 fmt.Sprintf("ORD-%s", strings.ToUpper(uuid.NewString()[:8])),
		CandidateID:        candidate.ID,
		ProductTitle:       candidate.Title,
		Partner:            partner,
		Quantity:           qty,
		Total:              multiplyMoney(candidate.RetailPrice, qty),
		CustomerName:       customerName,
		Status:             status,
		ShipmentStatus:     routingentity.RoutedOrderShipmentStatusAwaitingLabel,
		OperatorAssignee:   "unassigned",
		RoutingBlockCode:   routingBlockCode,
		RoutingBlockReason: routingBlockReason,
		BaseCostSnapshot:   multiplyMoney(candidate.BaseCost, qty),
		FulfillmentCost:    fulfillmentCost,
		ShippingCost:       shippingCost,
		IssueCost:          "$0.00",
		IssueResolution:    routingentity.RoutedOrderIssueResolutionMonitor,
		SettlementStatus:   routingentity.RoutedOrderSettlementStatusPending,
		Timeline: []string{
			fmt.Sprintf("Order created for %s", candidate.Title),
			timelineEntry,
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
					"status", status,
					"product_type", recommendation.ProductType,
					"ship_region", recommendation.ShipRegion,
				),
			),
			newActivity(
				routingentity.RoutedOrderActivityTypeSystem,
				actor,
				timelineEntry,
				now,
				activityDetails(
					"status", status,
					"partner", partner,
					"routing_summary", recommendation.Summary,
					"candidate_partner", recommendation.CandidatePartner,
					"estimated_unit_margin", estimatedMargin,
					"routing_block_code", routingBlockCode,
					"routing_block_reason", routingBlockReason,
				),
			),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	order.RealizedMargin = calculateMargin(order.Total, order.FulfillmentCost, order.ShippingCost, order.IssueCost)
	return i.orders.Create(ctx, order)
}

func (i *OrderRoutingInteractor) ForceRerouteBlockedOrder(
	ctx context.Context,
	cmd routinginputport.ForceRerouteBlockedOrderCmd,
) (*routingentity.RoutedOrder, error) {
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("routed order not found")
	}
	if order.Status != routingentity.RoutedOrderStatusRoutingBlocked {
		return nil, fmt.Errorf("routed order is not in routing_blocked status")
	}
	preferredPartner := strings.TrimSpace(cmd.PreferredPartner)
	if preferredPartner == "" {
		return nil, fmt.Errorf("preferred partner is required")
	}

	candidate, err := i.products.GetCandidateByID(ctx, strings.TrimSpace(order.CandidateID))
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
	recommendation := buildRoutingRecommendation(
		candidate,
		partners,
		normalizeRoutingLabel(orderRoutingLabel(order, candidate)),
		normalizeRoutingLabel(shipRegionFromOrder(order)),
		preferredPartner,
	)
	if recommendation.SelectedPartner == "" {
		return nil, fmt.Errorf("preferred partner %s is not eligible for reroute", preferredPartner)
	}
	selectedOption := findSelectedRoutingOption(recommendation)
	if selectedOption == nil {
		return nil, fmt.Errorf("selected routing option not found")
	}

	now := time.Now().UTC()
	order.Status = routingentity.RoutedOrderStatusQueued
	order.Partner = recommendation.SelectedPartner
	order.RoutingBlockCode = ""
	order.RoutingBlockReason = ""
	order.FulfillmentCost = multiplyMoney(selectedOption.EstimatedFulfillmentCost, order.Quantity)
	order.ShippingCost = selectedOption.EstimatedShippingCost
	order.RealizedMargin = calculateMargin(order.Total, order.FulfillmentCost, order.ShippingCost, order.IssueCost)
	entry := fmt.Sprintf("Routing unblocked: manually rerouted to %s", recommendation.SelectedPartner)
	order.Timeline = append(order.Timeline, entry)
	order.ActivityLog = append(order.ActivityLog, newActivity(
		routingentity.RoutedOrderActivityTypeSystem,
		activityActorFromContext(ctx),
		entry,
		now,
		activityDetails(
			"status", routingentity.RoutedOrderStatusQueued,
			"partner", recommendation.SelectedPartner,
			"routing_summary", recommendation.Summary,
			"estimated_unit_margin", selectedOption.EstimatedUnitMargin,
			"manual_reroute", "true",
		),
	))
	order.UpdatedAt = now
	return i.orders.Update(ctx, *order)
}

func orderRoutingLabel(order *routingentity.RoutedOrder, candidate *catalogentity.ProductSetupCandidate) string {
	for _, activity := range order.ActivityLog {
		for _, detail := range activity.Details {
			if detail.Key == "product_type" && strings.TrimSpace(detail.Value) != "" {
				return detail.Value
			}
		}
	}
	if candidate != nil {
		return inferProductType(candidate.Title)
	}
	return ""
}

func shipRegionFromOrder(order *routingentity.RoutedOrder) string {
	for _, activity := range order.ActivityLog {
		for _, detail := range activity.Details {
			if detail.Key == "ship_region" && strings.TrimSpace(detail.Value) != "" {
				return detail.Value
			}
		}
	}
	return ""
}

func inferProductType(title string) string {
	normalized := strings.ToLower(strings.TrimSpace(title))
	switch {
	case strings.Contains(normalized, "hoodie"):
		return "hoodie"
	case strings.Contains(normalized, "poster"):
		return "poster"
	case strings.Contains(normalized, "tote"):
		return "tote"
	case strings.Contains(normalized, "tee"), strings.Contains(normalized, "shirt"):
		return "tshirt"
	default:
		return ""
	}
}
