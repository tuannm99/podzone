package fulfillment

import (
	"context"
)

type FulfillmentOrderQueryRepository interface {
	GetFulfillmentOrder(ctx context.Context, storeID string, orderID string) (*FulfillmentOrder, error)
}
