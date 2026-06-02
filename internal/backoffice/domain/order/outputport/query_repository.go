package outputport

import (
	"context"

	orderentity "github.com/tuannm99/podzone/internal/backoffice/domain/order/entity"
)

type CustomerOrderQueryRepository interface {
	GetCustomerOrder(ctx context.Context, storeID string, orderID string) (*orderentity.CustomerOrder, error)
}
