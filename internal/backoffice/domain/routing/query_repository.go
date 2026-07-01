package routing

import (
	"context"

	"github.com/tuannm99/podzone/pkg/collection"
)

type RoutedOrderReadModelRepository interface {
	List(ctx context.Context) ([]RoutedOrder, error)
	ListByStore(ctx context.Context, storeID string) ([]RoutedOrder, error)
	ListPageByStore(
		ctx context.Context,
		storeID string,
		query collection.Query,
	) (collection.Page[RoutedOrder], error)
	ListActivityFeed(
		ctx context.Context,
		query RoutedOrderActivityFeedQuery,
	) (*RoutedOrderActivityFeedPage, error)
	GetByID(ctx context.Context, id string) (*RoutedOrder, error)
}
