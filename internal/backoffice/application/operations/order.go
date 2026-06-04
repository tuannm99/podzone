package operations

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	catalogctx "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	orderctx "github.com/tuannm99/podzone/internal/backoffice/domain/order"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type OrderInteractor struct {
	orders   routingctx.OrderRoutingRepository
	products catalogctx.ProductSetupRepository
	partners routingctx.PartnerDirectory
}

func NewOrderInteractor(
	orders routingctx.OrderRoutingRepository,
	products catalogctx.ProductSetupRepository,
	partners routingctx.PartnerDirectory,
) *OrderInteractor {
	return &OrderInteractor{orders: orders, products: products, partners: partners}
}

func (i *OrderInteractor) ListCustomerOrders(
	ctx context.Context,
	query orderctx.ListCustomerOrdersQuery,
) ([]routingctx.RoutedOrder, error) {
	storeID, err := routingctx.RequiredStoreScope(ctx, query.StoreID)
	if err != nil {
		return nil, err
	}
	return i.orders.ListByStore(ctx, storeID)
}

func (i *OrderInteractor) CreateCustomerOrder(
	ctx context.Context,
	cmd orderctx.CreateCustomerOrderCmd,
) (*routingctx.RoutedOrder, error) {
	storeID, err := routingctx.RequiredStoreScope(ctx, cmd.StoreID)
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
	if candidate == nil || candidate.Status != catalogctx.ProductSetupCandidateStatusPublishedMock {
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
	recommendation := routingctx.BuildRoutingRecommendation(
		candidate,
		partners,
		routingctx.NormalizeRoutingLabel(cmd.ProductType),
		routingctx.NormalizeRoutingLabel(cmd.ShipRegion),
		strings.TrimSpace(cmd.PreferredPartner),
	)
	selectedOption := routingctx.FindSelectedRoutingOption(recommendation)

	qty := cmd.Quantity
	if qty < 1 {
		qty = 1
	}
	now := time.Now().UTC()
	actor := routingctx.ActivityActorFromContext(ctx)
	partner := recommendation.SelectedPartner
	routingBlockCode := ""
	routingBlockReason := ""
	fulfillmentCost := "TBD"
	shippingCost := "TBD"
	estimatedMargin := "TBD"
	if selectedOption != nil {
		fulfillmentCost = routingctx.MultiplyMoney(selectedOption.EstimatedFulfillmentCost, qty)
		shippingCost = selectedOption.EstimatedShippingCost
		estimatedMargin = routingctx.CalculateMargin(
			routingctx.MultiplyMoney(candidate.RetailPrice, qty),
			fulfillmentCost,
			shippingCost,
			"$0.00",
		)
	}
	if recommendation.SelectedPartner == "" {
		partner = ""
		routingBlockCode = recommendation.BlockedReasonCode
		routingBlockReason = recommendation.BlockedReason
	}
	total := routingctx.MultiplyMoney(candidate.RetailPrice, qty)
	customerOrder, changes, err := orderctx.ReceiveCustomerOrder(orderctx.ReceiveCustomerOrderInput{
		ID:                 fmt.Sprintf("ORD-%s", strings.ToUpper(uuid.NewString()[:8])),
		StoreID:            storeID,
		CandidateID:        candidate.ID,
		ProductTitle:       candidate.Title,
		Quantity:           qty,
		Total:              total,
		CustomerName:       cmd.CustomerName,
		Partner:            partner,
		RoutingBlockCode:   routingBlockCode,
		RoutingBlockReason: routingBlockReason,
		Now:                now,
	})
	if err != nil {
		return nil, err
	}
	orderSnapshot := customerOrder.Snapshot()

	timeline := make([]string, 0, len(changes))
	for _, change := range changes {
		timeline = append(timeline, change.Message)
	}
	routingDetails := routingctx.ActivityDetails(
		"status", orderSnapshot.Status,
		"partner", orderSnapshot.Partner,
		"routing_summary", recommendation.Summary,
		"candidate_partner", recommendation.CandidatePartner,
		"estimated_unit_margin", estimatedMargin,
		"routing_block_code", routingBlockCode,
		"routing_block_reason", routingBlockReason,
	)
	for _, detail := range changes[1].Details {
		routingDetails = append(routingDetails, routingctx.RoutedOrderActivityDetail{
			Key:   detail.Key,
			Value: detail.Value,
		})
	}

	order := routingctx.RoutedOrder{
		ID:                 orderSnapshot.ID,
		StoreID:            orderSnapshot.StoreID,
		CandidateID:        orderSnapshot.CandidateID,
		ProductTitle:       orderSnapshot.ProductTitle,
		Partner:            orderSnapshot.Partner,
		Quantity:           orderSnapshot.Quantity,
		Total:              orderSnapshot.Total,
		CustomerName:       orderSnapshot.CustomerName,
		Status:             orderSnapshot.Status,
		ShipmentStatus:     routingctx.RoutedOrderShipmentStatusAwaitingLabel,
		OperatorAssignee:   orderSnapshot.OperatorAssignee,
		RoutingBlockCode:   orderSnapshot.RoutingBlockCode,
		RoutingBlockReason: orderSnapshot.RoutingBlockReason,
		BaseCostSnapshot:   routingctx.MultiplyMoney(candidate.BaseCost, qty),
		FulfillmentCost:    fulfillmentCost,
		ShippingCost:       shippingCost,
		IssueCost:          "$0.00",
		IssueResolution:    routingctx.RoutedOrderIssueResolutionMonitor,
		SettlementStatus:   orderSnapshot.SettlementStatus,
		Timeline:           timeline,
		ActivityLog: []routingctx.RoutedOrderActivity{
			routingctx.NewActivity(
				routingctx.RoutedOrderActivityTypeSystem,
				actor,
				changes[0].Message,
				now,
				routingctx.ActivityDetails(
					"candidate_id", candidate.ID,
					"quantity", fmt.Sprintf("%d", qty),
					"status", orderSnapshot.Status,
					"product_type", recommendation.ProductType,
					"ship_region", recommendation.ShipRegion,
				),
			),
			routingctx.NewActivity(
				routingctx.RoutedOrderActivityTypeSystem,
				actor,
				changes[1].Message,
				now,
				routingDetails,
			),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	order.RealizedMargin = routingctx.CalculateMargin(
		order.Total,
		order.FulfillmentCost,
		order.ShippingCost,
		order.IssueCost,
	)
	return i.orders.Create(ctx, order)
}

func (i *OrderInteractor) AdvanceCustomerOrder(
	ctx context.Context,
	cmd orderctx.AdvanceCustomerOrderCmd,
) (*routingctx.RoutedOrder, error) {
	storeID, err := routingctx.RequiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := routingctx.EnsureOrderStore(order, storeID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := advanceRoutedOrder(order, routingctx.ActivityActorFromContext(ctx), now); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}

func (i *OrderInteractor) UpdateOrderQueueControl(
	ctx context.Context,
	cmd orderctx.UpdateOrderQueueControlCmd,
) (*routingctx.RoutedOrder, error) {
	storeID, err := routingctx.RequiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := routingctx.EnsureOrderStore(order, storeID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	updateOrderQueueControl(order,
		cmd.OperatorAssignee,
		cmd.ShipmentSlaDueAt,
		cmd.IssueSlaDueAt,
		routingctx.ActivityActorFromContext(ctx),
		now,
	)
	return i.orders.Update(ctx, *order)
}

func (i *OrderInteractor) BulkUpdateOrders(
	ctx context.Context,
	cmd orderctx.BulkUpdateOrdersCmd,
) ([]routingctx.RoutedOrder, error) {
	storeID, err := routingctx.RequiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	if len(cmd.OrderIDs) == 0 {
		return nil, fmt.Errorf("at least one routed order id is required")
	}
	if cmd.OperatorAssignee == nil && cmd.ShipmentSlaDueAt == nil && cmd.SettlementStatus == nil {
		return nil, fmt.Errorf("at least one bulk update field is required")
	}

	updated := make([]routingctx.RoutedOrder, 0, len(cmd.OrderIDs))
	for _, rawID := range cmd.OrderIDs {
		orderID := strings.TrimSpace(rawID)
		if orderID == "" {
			continue
		}
		order, err := i.orders.GetByID(ctx, orderID)
		if err != nil {
			return nil, err
		}
		if err := routingctx.EnsureOrderStore(order, storeID); err != nil {
			return nil, err
		}

		now := time.Now().UTC()
		if err := applyBulkQueueControl(order,
			cmd.OperatorAssignee,
			cmd.ShipmentSlaDueAt,
			cmd.SettlementStatus,
			routingctx.ActivityActorFromContext(ctx),
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
