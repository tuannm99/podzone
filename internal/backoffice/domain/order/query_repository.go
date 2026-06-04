package order

import (
	"context"
)

type CustomerOrderQueryRepository interface {
	GetCustomerOrder(ctx context.Context, storeID string, orderID string) (*CustomerOrder, error)
}
