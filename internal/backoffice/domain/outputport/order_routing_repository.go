package outputport

import (
	"context"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/inputport"
)

type OrderRoutingRepository interface {
	List(ctx context.Context) ([]entity.RoutedOrder, error)
	ListActivityFeed(ctx context.Context, query inputport.ListRoutedOrderActivitiesQuery) (*entity.RoutedOrderActivityFeedPage, error)
	GetByID(ctx context.Context, id string) (*entity.RoutedOrder, error)
	Create(ctx context.Context, order entity.RoutedOrder) (*entity.RoutedOrder, error)
	Update(ctx context.Context, order entity.RoutedOrder) (*entity.RoutedOrder, error)
}
