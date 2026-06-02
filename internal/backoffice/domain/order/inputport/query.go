package inputport

import (
	"context"

	routingentity "github.com/tuannm99/podzone/internal/backoffice/domain/routing/entity"
)

type ListCustomerOrdersQuery struct {
	StoreID string
}

type CustomerOrderQueryUsecase interface {
	ListCustomerOrders(ctx context.Context, query ListCustomerOrdersQuery) ([]routingentity.RoutedOrder, error)
}
