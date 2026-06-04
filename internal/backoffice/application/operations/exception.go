package operations

import (
	"context"
	"strings"
	"time"

	exceptionctx "github.com/tuannm99/podzone/internal/backoffice/domain/exception"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
)

type ExceptionInteractor struct {
	orders routingctx.OrderRoutingRepository
}

func NewExceptionInteractor(orders routingctx.OrderRoutingRepository) *ExceptionInteractor {
	return &ExceptionInteractor{orders: orders}
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
	now := time.Now().UTC()
	if err := openOrderException(order, cmd.ExceptionType, routingctx.ActivityActorFromContext(ctx), now); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
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
	now := time.Now().UTC()
	if err := updateOrderExceptionStatus(order, cmd.Status, routingctx.ActivityActorFromContext(ctx), now); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}
