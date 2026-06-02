package outputport

import (
	"context"

	fulfillmententity "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment/entity"
)

type FulfillmentOrderQueryRepository interface {
	GetFulfillmentOrder(ctx context.Context, storeID string, orderID string) (*fulfillmententity.FulfillmentOrder, error)
}
