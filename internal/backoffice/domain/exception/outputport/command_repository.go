package outputport

import (
	"context"

	exceptionentity "github.com/tuannm99/podzone/internal/backoffice/domain/exception/entity"
)

type OrderExceptionCommandRepository interface {
	SaveOrderException(ctx context.Context, exception *exceptionentity.OrderException) error
}
