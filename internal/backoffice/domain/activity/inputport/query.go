package inputport

import (
	"context"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type ActivityFeedQueryUsecase interface {
	ListRoutedOrderActivities(
		ctx context.Context,
		query routingentity.RoutedOrderActivityFeedQuery,
	) (*routingentity.RoutedOrderActivityFeedPage, error)
}
