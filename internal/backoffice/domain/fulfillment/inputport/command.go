package inputport

import (
	"context"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type UpdateOrderShipmentCmd struct {
	StoreID        string
	OrderID        string
	ShipmentStatus string
	Carrier        string
	TrackingNumber string
	TrackingURL    string
	Notes          string
}

type FulfillmentCommandUsecase interface {
	UpdateOrderShipment(ctx context.Context, cmd UpdateOrderShipmentCmd) (*routingentity.RoutedOrder, error)
}
