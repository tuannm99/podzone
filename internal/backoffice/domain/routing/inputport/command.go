package inputport

import (
	"context"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type RecommendRoutedOrderPartnerQuery struct {
	StoreID          string
	CandidateID      string
	ProductType      string
	ShipRegion       string
	PreferredPartner string
}

type ForceRerouteBlockedOrderCmd struct {
	StoreID          string
	OrderID          string
	PreferredPartner string
}

type RoutingCommandUsecase interface {
	ForceRerouteBlockedOrder(ctx context.Context, cmd ForceRerouteBlockedOrderCmd) (*routingentity.RoutedOrder, error)
}
