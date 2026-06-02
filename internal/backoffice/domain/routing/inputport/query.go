package inputport

import (
	"context"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type RoutingQueryUsecase interface {
	RecommendRoutedOrderPartner(
		ctx context.Context,
		query RecommendRoutedOrderPartnerQuery,
	) (*routingentity.RoutedOrderRecommendation, error)
}
