package operations

import (
	"context"
	"fmt"
	"strings"

	catalogctx "github.com/tuannm99/podzone/internal/backoffice/domain/catalog"
	orderctx "github.com/tuannm99/podzone/internal/backoffice/domain/order"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	"github.com/tuannm99/podzone/pkg/collection"
	"github.com/tuannm99/podzone/pkg/ddd"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type OrderInteractor struct {
	orders         routingctx.OrderRoutingRepository
	customerOrders orderctx.CustomerOrderQueryRepository
	products       catalogctx.ProductSetupRepository
	partners       routingctx.PartnerDirectory
	events         ddd.EventDispatcher
	ids            ddd.IDGenerator
	clock          ddd.Clock
}

func NewOrderInteractor(
	orders routingctx.OrderRoutingRepository,
	customerOrders orderctx.CustomerOrderQueryRepository,
	products catalogctx.ProductSetupRepository,
	partners routingctx.PartnerDirectory,
	dispatcher ddd.EventDispatcher,
	ids ddd.IDGenerator,
	clock ddd.Clock,
) *OrderInteractor {
	return &OrderInteractor{
		orders:         orders,
		customerOrders: customerOrders,
		products:       products,
		partners:       partners,
		events:         dispatcher,
		ids:            ids,
		clock:          clock,
	}
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

func (i *OrderInteractor) ListCustomerOrderPage(
	ctx context.Context,
	query orderctx.ListCustomerOrderPageQuery,
) (collection.Page[routingctx.RoutedOrder], error) {
	storeID, err := routingctx.RequiredStoreScope(ctx, query.StoreID)
	if err != nil {
		return collection.Page[routingctx.RoutedOrder]{}, err
	}
	return i.orders.ListPageByStore(ctx, storeID, query.Collection.Normalize())
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
		i.clock.Now(),
	)
	selectedOption := routingctx.FindSelectedRoutingOption(recommendation)

	qty := cmd.Quantity
	if qty < 1 {
		qty = 1
	}
	now := i.clock.Now()
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
	orderID, err := i.ids.NewID("order")
	if err != nil {
		return nil, err
	}
	customerOrder, changes, err := orderctx.ReceiveCustomerOrder(orderctx.ReceiveCustomerOrderInput{
		ID:                 orderID.String(),
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
	domainEvents := collectDomainEvents(customerOrder.PullEvents())
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
	saved, err := i.orders.Create(ctx, order)
	if err != nil {
		return nil, err
	}
	if err := dispatchDomainEvents(ctx, i.events, domainEvents); err != nil {
		return nil, err
	}
	return saved, nil
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
	customerOrder, err := i.customerOrders.GetCustomerOrder(ctx, storeID, order.ID)
	if err != nil {
		return nil, err
	}
	now := i.clock.Now()
	domainEvents, err := advanceRoutedOrder(
		order,
		customerOrder,
		routingctx.ActivityActorFromContext(ctx),
		now,
	)
	if err != nil {
		return nil, err
	}
	saved, err := i.orders.Update(ctx, *order)
	if err != nil {
		return nil, err
	}
	if err := dispatchDomainEvents(ctx, i.events, domainEvents); err != nil {
		return nil, err
	}
	return saved, nil
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
	customerOrder, err := i.customerOrders.GetCustomerOrder(ctx, storeID, order.ID)
	if err != nil {
		return nil, err
	}

	now := i.clock.Now()
	domainEvents := updateOrderQueueControl(
		order,
		customerOrder,
		cmd.OperatorAssignee,
		cmd.ShipmentSlaDueAt,
		cmd.IssueSlaDueAt,
		routingctx.ActivityActorFromContext(ctx),
		now,
	)
	saved, err := i.orders.Update(ctx, *order)
	if err != nil {
		return nil, err
	}
	if err := dispatchDomainEvents(ctx, i.events, domainEvents); err != nil {
		return nil, err
	}
	return saved, nil
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
		customerOrder, err := i.customerOrders.GetCustomerOrder(ctx, storeID, order.ID)
		if err != nil {
			return nil, err
		}

		now := i.clock.Now()
		domainEvents, err := applyBulkQueueControl(
			order,
			customerOrder,
			cmd.OperatorAssignee,
			cmd.ShipmentSlaDueAt,
			cmd.SettlementStatus,
			routingctx.ActivityActorFromContext(ctx),
			now,
		)
		if err != nil {
			return nil, err
		}
		saved, err := i.orders.Update(ctx, *order)
		if err != nil {
			return nil, err
		}
		if err := dispatchDomainEvents(ctx, i.events, domainEvents); err != nil {
			return nil, err
		}
		updated = append(updated, *saved)
	}
	return updated, nil
}
