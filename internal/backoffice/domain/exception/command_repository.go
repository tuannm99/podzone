package exception

import (
	"context"
)

type OrderExceptionCommandRepository interface {
	SaveOrderException(ctx context.Context, exception *OrderException) error
}
