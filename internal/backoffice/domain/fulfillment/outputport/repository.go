package outputport

import (
	"context"

	fulfillmententity "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment/entity"
)

type FulfillmentOrderRepository interface {
	GetFulfillmentOrder(ctx context.Context, storeID string, orderID string) (*fulfillmententity.FulfillmentOrder, error)
	SaveFulfillmentOrder(ctx context.Context, order *fulfillmententity.FulfillmentOrder) error
}
