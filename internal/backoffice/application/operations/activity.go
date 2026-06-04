package operations

import (
	"context"

	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
)

type ActivityInteractor struct {
	orders routingctx.RoutedOrderReadModelRepository
}

func NewActivityInteractor(orders routingctx.RoutedOrderReadModelRepository) *ActivityInteractor {
	return &ActivityInteractor{orders: orders}
}

func (i *ActivityInteractor) ListRoutedOrderActivities(
	ctx context.Context,
	query routingctx.RoutedOrderActivityFeedQuery,
) (*routingctx.RoutedOrderActivityFeedPage, error) {
	storeID, err := routingctx.RequiredStoreScope(ctx, query.StoreID)
	if err != nil {
		return nil, err
	}
	query.StoreID = storeID
	return i.orders.ListActivityFeed(ctx, query)
}
