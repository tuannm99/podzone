package exception

import (
	"context"
)

type OrderExceptionQueryRepository interface {
	GetOrderException(ctx context.Context, storeID string, orderID string) (*OrderException, error)
}
