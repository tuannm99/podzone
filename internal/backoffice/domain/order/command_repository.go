package order

import (
	"context"
)

type CustomerOrderCommandRepository interface {
	SaveCustomerOrder(ctx context.Context, order *CustomerOrder) error
}
