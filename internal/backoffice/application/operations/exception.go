package operations

import (
	"context"
	"strings"

	exceptionctx "github.com/tuannm99/podzone/internal/backoffice/domain/exception"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	"github.com/tuannm99/podzone/pkg/ddd"
)

type ExceptionInteractor struct {
	orders routingctx.OrderRoutingRepository
	events ddd.EventDispatcher
	clock  ddd.Clock
}

func NewExceptionInteractor(
	orders routingctx.OrderRoutingRepository,
	dispatcher ddd.EventDispatcher,
	clock ddd.Clock,
) *ExceptionInteractor {
	return &ExceptionInteractor{orders: orders, events: dispatcher, clock: clock}
}

func (i *ExceptionInteractor) OpenOrderException(
	ctx context.Context,
	cmd exceptionctx.OpenOrderExceptionCmd,
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
	domainEvents, err := openOrderException(order, cmd.ExceptionType, routingctx.ActivityActorFromContext(ctx), now)
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

func (i *ExceptionInteractor) UpdateOrderExceptionStatus(
	ctx context.Context,
	cmd exceptionctx.UpdateOrderExceptionStatusCmd,
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
	domainEvents, err := updateOrderExceptionStatus(order, cmd.Status, routingctx.ActivityActorFromContext(ctx), now)
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
