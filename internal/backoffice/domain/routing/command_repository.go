package routing

import (
	"context"
)

type RoutedOrderCommandRepository interface {
	Create(ctx context.Context, order RoutedOrder) (*RoutedOrder, error)
	Update(ctx context.Context, order RoutedOrder) (*RoutedOrder, error)
}
