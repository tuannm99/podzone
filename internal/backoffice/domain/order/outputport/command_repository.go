package outputport

import (
	"context"

	orderentity "github.com/tuannm99/podzone/internal/backoffice/domain/order/entity"
)

type CustomerOrderCommandRepository interface {
	SaveCustomerOrder(ctx context.Context, order *orderentity.CustomerOrder) error
}
