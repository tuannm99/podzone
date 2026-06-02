package outputport

import (
	"context"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type OrderRoutingRepository interface {
	RoutedOrderCommandRepository
	RoutedOrderReadModelRepository
}

type PartnerDirectory interface {
	ListActivePartners(ctx context.Context, tenantID string) ([]routingentity.PartnerRoutingProfile, error)
}
