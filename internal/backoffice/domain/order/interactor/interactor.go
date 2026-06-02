package interactor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	catalogentity "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/entity"
	catalogoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/catalog/outputport"
	orderinputport "github.com/tuannm99/podzone/internal/backoffice/domain/order/inputport"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routingoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/outputport"
	routingpolicy "github.com/tuannm99/podzone/internal/backoffice/domain/routing/policy"
	routingsupport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/support"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type Interactor struct {
	orders   routingoutputport.OrderRoutingRepository
	products catalogoutputport.ProductSetupRepository
	partners routingoutputport.PartnerDirectory
}

var (
	_ orderinputport.CustomerOrderCommandUsecase = (*Interactor)(nil)
	_ orderinputport.CustomerOrderQueryUsecase   = (*Interactor)(nil)
)

func New(
	orders routingoutputport.OrderRoutingRepository,
	products catalogoutputport.ProductSetupRepository,
	partners routingoutputport.PartnerDirectory,
) *Interactor {
	return &Interactor{orders: orders, products: products, partners: partners}
}

func (i *Interactor) ListCustomerOrders(
	ctx context.Context,
	query orderinputport.ListCustomerOrdersQuery,
) ([]routingentity.RoutedOrder, error) {
	storeID, err := routingsupport.RequiredStoreScope(ctx, query.StoreID)
	if err != nil {
		return nil, err
	}
	return i.orders.ListByStore(ctx, storeID)
}

func (i *Interactor) CreateCustomerOrder(
	ctx context.Context,
	cmd orderinputport.CreateCustomerOrderCmd,
) (*routingentity.RoutedOrder, error) {
	storeID, err := routingsupport.RequiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	candidateID := strings.TrimSpace(cmd.CandidateID)
	if candidateID == "" {
		return nil, fmt.Errorf("candidate id is required")
	}
	candidate, err := i.products.GetCandidateByID(ctx, storeID, candidateID)
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
	recommendation := routingpolicy.BuildRoutingRecommendation(
		candidate,
		partners,
		routingpolicy.NormalizeRoutingLabel(cmd.ProductType),
		routingpolicy.NormalizeRoutingLabel(cmd.ShipRegion),
		strings.TrimSpace(cmd.PreferredPartner),
	)
	selectedOption := routingpolicy.FindSelectedRoutingOption(recommendation)

	qty := cmd.Quantity
	if qty < 1 {
		qty = 1
	}
	customerName := strings.TrimSpace(cmd.CustomerName)
	if customerName == "" {
		customerName = "Sample customer"
	}
	now := time.Now().UTC()
	actor := routingsupport.ActivityActorFromContext(ctx)
	status := routingentity.RoutedOrderStatusQueued
	partner := recommendation.SelectedPartner
	timelineEntry := routingpolicy.RoutingTimelineEntry(
		routingentity.RoutedOrderStatusQueued,
		recommendation.SelectedPartner,
	)
	routingBlockCode := ""
	routingBlockReason := ""
	fulfillmentCost := "TBD"
	shippingCost := "TBD"
	estimatedMargin := "TBD"
	if selectedOption != nil {
		fulfillmentCost = routingpolicy.MultiplyMoney(selectedOption.EstimatedFulfillmentCost, qty)
		shippingCost = selectedOption.EstimatedShippingCost
		estimatedMargin = routingpolicy.CalculateMargin(
			routingpolicy.MultiplyMoney(candidate.RetailPrice, qty),
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
		StoreID:            storeID,
		CandidateID:        candidate.ID,
		ProductTitle:       candidate.Title,
		Partner:            partner,
		Quantity:           qty,
		Total:              routingpolicy.MultiplyMoney(candidate.RetailPrice, qty),
		CustomerName:       customerName,
		Status:             status,
		ShipmentStatus:     routingentity.RoutedOrderShipmentStatusAwaitingLabel,
		OperatorAssignee:   "unassigned",
		RoutingBlockCode:   routingBlockCode,
		RoutingBlockReason: routingBlockReason,
		BaseCostSnapshot:   routingpolicy.MultiplyMoney(candidate.BaseCost, qty),
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
			routingpolicy.NewActivity(
				routingentity.RoutedOrderActivityTypeSystem,
				actor,
				fmt.Sprintf("Order created for %s", candidate.Title),
				now,
				routingpolicy.ActivityDetails(
					"candidate_id", candidate.ID,
					"quantity", fmt.Sprintf("%d", qty),
					"status", status,
					"product_type", recommendation.ProductType,
					"ship_region", recommendation.ShipRegion,
				),
			),
			routingpolicy.NewActivity(
				routingentity.RoutedOrderActivityTypeSystem,
				actor,
				timelineEntry,
				now,
				routingpolicy.ActivityDetails(
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
	order.RealizedMargin = routingpolicy.CalculateMargin(
		order.Total,
		order.FulfillmentCost,
		order.ShippingCost,
		order.IssueCost,
	)
	return i.orders.Create(ctx, order)
}

func (i *Interactor) AdvanceCustomerOrder(
	ctx context.Context,
	cmd orderinputport.AdvanceCustomerOrderCmd,
) (*routingentity.RoutedOrder, error) {
	storeID, err := routingsupport.RequiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := routingsupport.EnsureOrderStore(order, storeID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := order.Advance(routingsupport.ActivityActorFromContext(ctx), now); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}

func (i *Interactor) UpdateOrderQueueControl(
	ctx context.Context,
	cmd orderinputport.UpdateOrderQueueControlCmd,
) (*routingentity.RoutedOrder, error) {
	storeID, err := routingsupport.RequiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := routingsupport.EnsureOrderStore(order, storeID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	order.UpdateQueueControl(
		cmd.OperatorAssignee,
		cmd.ShipmentSlaDueAt,
		cmd.IssueSlaDueAt,
		routingsupport.ActivityActorFromContext(ctx),
		now,
	)
	return i.orders.Update(ctx, *order)
}

func (i *Interactor) BulkUpdateOrders(
	ctx context.Context,
	cmd orderinputport.BulkUpdateOrdersCmd,
) ([]routingentity.RoutedOrder, error) {
	storeID, err := routingsupport.RequiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	if len(cmd.OrderIDs) == 0 {
		return nil, fmt.Errorf("at least one routed order id is required")
	}
	if cmd.OperatorAssignee == nil && cmd.ShipmentSlaDueAt == nil && cmd.SettlementStatus == nil {
		return nil, fmt.Errorf("at least one bulk update field is required")
	}

	updated := make([]routingentity.RoutedOrder, 0, len(cmd.OrderIDs))
	for _, rawID := range cmd.OrderIDs {
		orderID := strings.TrimSpace(rawID)
		if orderID == "" {
			continue
		}
		order, err := i.orders.GetByID(ctx, orderID)
		if err != nil {
			return nil, err
		}
		if err := routingsupport.EnsureOrderStore(order, storeID); err != nil {
			return nil, err
		}

		now := time.Now().UTC()
		if err := order.ApplyBulkQueueControl(
			cmd.OperatorAssignee,
			cmd.ShipmentSlaDueAt,
			cmd.SettlementStatus,
			routingsupport.ActivityActorFromContext(ctx),
			now,
		); err != nil {
			return nil, err
		}
		saved, err := i.orders.Update(ctx, *order)
		if err != nil {
			return nil, err
		}
		updated = append(updated, *saved)
	}
	return updated, nil
}
