package interactor

import (
	"context"

	activityinputport "github.com/tuannm99/podzone/internal/backoffice/domain/activity/inputport"
	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
	routingoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/outputport"
	routingsupport "github.com/tuannm99/podzone/internal/backoffice/domain/routing/support"
)

type Interactor struct {
	orders routingoutputport.RoutedOrderReadModelRepository
}

var _ activityinputport.ActivityFeedQueryUsecase = (*Interactor)(nil)

func New(orders routingoutputport.RoutedOrderReadModelRepository) *Interactor {
	return &Interactor{orders: orders}
}

func (i *Interactor) ListRoutedOrderActivities(
	ctx context.Context,
	query routingentity.RoutedOrderActivityFeedQuery,
) (*routingentity.RoutedOrderActivityFeedPage, error) {
	storeID, err := routingsupport.RequiredStoreScope(ctx, query.StoreID)
	if err != nil {
		return nil, err
	}
	query.StoreID = storeID
	return i.orders.ListActivityFeed(ctx, query)
}
