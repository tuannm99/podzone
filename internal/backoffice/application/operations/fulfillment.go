package operations

import (
	"context"
	"strings"
	"time"

	fulfillmentctx "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
)

type FulfillmentInteractor struct {
	orders routingctx.OrderRoutingRepository
}

func NewFulfillmentInteractor(orders routingctx.OrderRoutingRepository) *FulfillmentInteractor {
	return &FulfillmentInteractor{orders: orders}
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
	now := time.Now().UTC()
	if err := updateOrderShipment(order,
		cmd.ShipmentStatus,
		cmd.Carrier,
		cmd.TrackingNumber,
		cmd.TrackingURL,
		cmd.Notes,
		routingctx.ActivityActorFromContext(ctx),
		now,
	); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}
