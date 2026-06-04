package routing

import (
	"context"
)

type OrderRoutingRepository interface {
	RoutedOrderCommandRepository
	RoutedOrderReadModelRepository
}

type PartnerDirectory interface {
	ListActivePartners(ctx context.Context, tenantID string) ([]PartnerRoutingProfile, error)
}
