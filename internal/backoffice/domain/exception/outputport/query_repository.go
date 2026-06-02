package outputport

import (
	"context"

	exceptionentity "github.com/tuannm99/podzone/internal/backoffice/domain/exception/entity"
)

type OrderExceptionQueryRepository interface {
	GetOrderException(ctx context.Context, storeID string, orderID string) (*exceptionentity.OrderException, error)
}
