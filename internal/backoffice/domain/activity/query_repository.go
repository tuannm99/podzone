package activity

import (
	"context"
)

type ActivityQueryRepository interface {
	ListOrderActivity(ctx context.Context, storeID string, orderID string) ([]ActivityEntry, error)
}
