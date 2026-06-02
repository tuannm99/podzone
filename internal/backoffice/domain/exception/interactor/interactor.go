package interactor

import (
	"context"
	"strings"
	"time"

	exceptioninputport "github.com/tuannm99/podzone/internal/backoffice/domain/exception/inputport"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routingoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/outputport"
	routingsupport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/support"
)

type Interactor struct {
	orders routingoutputport.OrderRoutingRepository
}

var _ exceptioninputport.ExceptionCommandUsecase = (*Interactor)(nil)

func New(orders routingoutputport.OrderRoutingRepository) *Interactor {
	return &Interactor{orders: orders}
}

func (i *Interactor) OpenOrderException(
	ctx context.Context,
	cmd exceptioninputport.OpenOrderExceptionCmd,
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
	if err := order.OpenException(cmd.ExceptionType, routingsupport.ActivityActorFromContext(ctx), now); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}

func (i *Interactor) UpdateOrderExceptionStatus(
	ctx context.Context,
	cmd exceptioninputport.UpdateOrderExceptionStatusCmd,
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
	if err := order.UpdateExceptionStatus(cmd.Status, routingsupport.ActivityActorFromContext(ctx), now); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}
