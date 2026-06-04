package fulfillment

import (
	"context"
)

type FulfillmentOrderCommandRepository interface {
	SaveFulfillmentOrder(ctx context.Context, order *FulfillmentOrder) error
}
