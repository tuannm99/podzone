package routing

import (
	"context"
)

type RecommendRoutedOrderPartnerQuery struct {
	StoreID          string
	CandidateID      string
	ProductType      string
	ShipRegion       string
	PreferredPartner string
}

type RoutingQueryUsecase interface {
	RecommendRoutedOrderPartner(
		ctx context.Context,
		query RecommendRoutedOrderPartnerQuery,
	) (*RoutedOrderRecommendation, error)
}
