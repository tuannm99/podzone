package outputport

import (
	"context"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type RoutedOrderReadModelRepository interface {
	List(ctx context.Context) ([]routingentity.RoutedOrder, error)
	ListByStore(ctx context.Context, storeID string) ([]routingentity.RoutedOrder, error)
	ListActivityFeed(
		ctx context.Context,
		query routingentity.RoutedOrderActivityFeedQuery,
	) (*routingentity.RoutedOrderActivityFeedPage, error)
	GetByID(ctx context.Context, id string) (*routingentity.RoutedOrder, error)
}
