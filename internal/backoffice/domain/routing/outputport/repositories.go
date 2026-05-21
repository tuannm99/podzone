package outputport

import (
	"context"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routingusecase "github.com/tuannm99/podzone/internal/backoffice/domain/routing/usecase"
)

type OrderRoutingRepository interface {
	List(ctx context.Context) ([]routingentity.RoutedOrder, error)
	ListActivityFeed(
		ctx context.Context,
		query routingusecase.ListRoutedOrderActivitiesQuery,
	) (*routingentity.RoutedOrderActivityFeedPage, error)
	GetByID(ctx context.Context, id string) (*routingentity.RoutedOrder, error)
	Create(ctx context.Context, order routingentity.RoutedOrder) (*routingentity.RoutedOrder, error)
	Update(ctx context.Context, order routingentity.RoutedOrder) (*routingentity.RoutedOrder, error)
}

type PartnerDirectory interface {
	ListActivePartners(ctx context.Context, tenantID string) ([]routingentity.PartnerRoutingProfile, error)
}
