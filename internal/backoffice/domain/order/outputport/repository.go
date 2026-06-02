package outputport

import (
	"context"

	orderentity "github.com/tuannm99/podzone/internal/backoffice/domain/order/entity"
)

type CustomerOrderRepository interface {
	GetCustomerOrder(ctx context.Context, storeID string, orderID string) (*orderentity.CustomerOrder, error)
	SaveCustomerOrder(ctx context.Context, order *orderentity.CustomerOrder) error
}
