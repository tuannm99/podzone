package outputport

import (
	"context"

	activityentity "github.com/tuannm99/podzone/internal/backoffice/domain/activity/entity"
)

type ActivityQueryRepository interface {
	ListOrderActivity(ctx context.Context, storeID string, orderID string) ([]activityentity.ActivityEntry, error)
}
