package outputport

import (
	"context"

	fulfillmententity "github.com/tuannm99/podzone/internal/backoffice/domain/fulfillment/entity"
)

type FulfillmentOrderCommandRepository interface {
	SaveFulfillmentOrder(ctx context.Context, order *fulfillmententity.FulfillmentOrder) error
}
