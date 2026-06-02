package interactor

import (
	"context"
	"strings"
	"time"

	fulfillmentinputport "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment/inputport"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routingoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/outputport"
	routingsupport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/support"
)

type Interactor struct {
	orders routingoutputport.OrderRoutingRepository
}

var _ fulfillmentinputport.FulfillmentCommandUsecase = (*Interactor)(nil)

func New(orders routingoutputport.OrderRoutingRepository) *Interactor {
	return &Interactor{orders: orders}
}

func (i *Interactor) UpdateOrderShipment(
	ctx context.Context,
	cmd fulfillmentinputport.UpdateOrderShipmentCmd,
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
	if err := order.UpdateShipment(
		cmd.ShipmentStatus,
		cmd.Carrier,
		cmd.TrackingNumber,
		cmd.TrackingURL,
		cmd.Notes,
		routingsupport.ActivityActorFromContext(ctx),
		now,
	); err != nil {
		return nil, err
	}
	return i.orders.Update(ctx, *order)
}
