package routing

import "context"

type ForceRerouteBlockedOrderCmd struct {
	StoreID          string
	OrderID          string
	PreferredPartner string
}

type RoutingCommandUsecase interface {
	ForceRerouteBlockedOrder(ctx context.Context, cmd ForceRerouteBlockedOrderCmd) (*RoutedOrder, error)
}
