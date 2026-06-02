package routing

import (
	"context"
	"strings"
	"time"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routinginputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/inputport"
)

func (i *OrderRoutingInteractor) AdvanceRoutedOrder(
	ctx context.Context,
	storeID string,
	orderID string,
) (*routingentity.RoutedOrder, error) {
	storeID, err := requiredStoreScope(ctx, storeID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(orderID))
	if err != nil {
		return nil, err
	}
	if err := ensureOrderStore(order, storeID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := order.Advance(activityActorFromContext(ctx), now); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) OpenOrderException(
	ctx context.Context,
	cmd routinginputport.OpenOrderExceptionCmd,
) (*routingentity.RoutedOrder, error) {
	storeID, err := requiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := ensureOrderStore(order, storeID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := order.OpenException(cmd.ExceptionType, activityActorFromContext(ctx), now); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}

func (i *OrderRoutingInteractor) UpdateOrderExceptionStatus(
	ctx context.Context,
	cmd routinginputport.UpdateOrderExceptionStatusCmd,
) (*routingentity.RoutedOrder, error) {
	storeID, err := requiredStoreScope(ctx, cmd.StoreID)
	if err != nil {
		return nil, err
	}
	order, err := i.orders.GetByID(ctx, strings.TrimSpace(cmd.OrderID))
	if err != nil {
		return nil, err
	}
	if err := ensureOrderStore(order, storeID); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := order.UpdateExceptionStatus(cmd.Status, activityActorFromContext(ctx), now); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}
