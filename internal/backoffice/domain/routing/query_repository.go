package routing

import (
	"context"
)

type RoutedOrderReadModelRepository interface {
	List(ctx context.Context) ([]RoutedOrder, error)
	ListByStore(ctx context.Context, storeID string) ([]RoutedOrder, error)
	ListActivityFeed(
		ctx context.Context,
		query RoutedOrderActivityFeedQuery,
	) (*RoutedOrderActivityFeedPage, error)
	GetByID(ctx context.Context, id string) (*RoutedOrder, error)
}
