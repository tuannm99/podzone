package operations

import (
	"context"
	"strings"

	fulfillmentctx "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment"
	orderctx "github.com/tuannm99/podzone/internal/backoffice/domain/order"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	"github.com/tuannm99/podzone/pkg/ddd"
)

type FulfillmentInteractor struct {
	orders         routingctx.OrderRoutingRepository
	customerOrders orderctx.CustomerOrderQueryRepository
	events         ddd.EventDispatcher
	clock          ddd.Clock
}

func NewFulfillmentInteractor(
	orders routingctx.OrderRoutingRepository,
	customerOrders orderctx.CustomerOrderQueryRepository,
	dispatcher ddd.EventDispatcher,
	clock ddd.Clock,
) *FulfillmentInteractor {
	return &FulfillmentInteractor{
		orders:         orders,
		customerOrders: customerOrders,
		events:         dispatcher,
		clock:          clock,
	}
}

func (i *FulfillmentInteractor) UpdateOrderShipment(
	ctx context.Context,
	cmd fulfillmentctx.UpdateOrderShipmentCmd,
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
	domainEvents, err := updateOrderShipment(
		order,
		customerOrder,
		cmd.ShipmentStatus,
		cmd.Carrier,
		cmd.TrackingNumber,
		cmd.TrackingURL,
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
