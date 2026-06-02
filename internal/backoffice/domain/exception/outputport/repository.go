package outputport

import (
	"context"

	exceptionentity "github.com/tuannm99/podzone/internal/backoffice/domain/exception/entity"
)

type OrderExceptionRepository interface {
	GetOrderException(ctx context.Context, storeID string, orderID string) (*exceptionentity.OrderException, error)
	SaveOrderException(ctx context.Context, exception *exceptionentity.OrderException) error
}
