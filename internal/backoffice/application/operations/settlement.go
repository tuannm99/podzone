package operations

import (
	"context"
	"strings"

	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	settlementctx "github.com/tuannm99/podzone/internal/backoffice/domain/settlement"
	"github.com/tuannm99/podzone/pkg/ddd"
)

type SettlementInteractor struct {
	orders routingctx.OrderRoutingRepository
	events ddd.EventDispatcher
	clock  ddd.Clock
}

func NewSettlementInteractor(
	orders routingctx.OrderRoutingRepository,
	dispatcher ddd.EventDispatcher,
	clock ddd.Clock,
) *SettlementInteractor {
	return &SettlementInteractor{orders: orders, events: dispatcher, clock: clock}
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

	now := i.clock.Now()
	domainEvents, err := updateOrderSettlement(order,
		cmd.FulfillmentCost,
		cmd.ShippingCost,
		cmd.SettlementStatus,
		cmd.Notes,
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

	now := i.clock.Now()
	domainEvents, err := updateOrderIssueHandling(order,
		cmd.IssueCost,
		cmd.IssueResolution,
		cmd.Notes,
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
