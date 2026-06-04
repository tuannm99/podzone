package operations

import (
	"context"
	"strings"
	"time"

	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	settlementctx "github.com/tuannm99/podzone/internal/backoffice/domain/settlement"
)

type SettlementInteractor struct {
	orders routingctx.OrderRoutingRepository
}

func NewSettlementInteractor(orders routingctx.OrderRoutingRepository) *SettlementInteractor {
	return &SettlementInteractor{orders: orders}
}

func (i *SettlementInteractor) UpdateOrderSettlement(
	ctx context.Context,
	cmd settlementctx.UpdateOrderSettlementCmd,
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
	if err := updateOrderSettlement(order,
		cmd.FulfillmentCost,
		cmd.ShippingCost,
		cmd.SettlementStatus,
		cmd.Notes,
		routingctx.ActivityActorFromContext(ctx),
		now,
	); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}

func (i *SettlementInteractor) UpdateOrderIssueHandling(
	ctx context.Context,
	cmd settlementctx.UpdateOrderIssueHandlingCmd,
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
	if err := updateOrderIssueHandling(order,
		cmd.IssueCost,
		cmd.IssueResolution,
		cmd.Notes,
		routingctx.ActivityActorFromContext(ctx),
		now,
	); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}
