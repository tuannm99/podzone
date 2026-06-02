package outputport

import (
	"context"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type RoutedOrderCommandRepository interface {
	Create(ctx context.Context, order routingentity.RoutedOrder) (*routingentity.RoutedOrder, error)
	Update(ctx context.Context, order routingentity.RoutedOrder) (*routingentity.RoutedOrder, error)
}
