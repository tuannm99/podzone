package outputport

import (
	"context"

	activityentity "github.com/tuannm99/podzone/internal/backoffice/domain/activity/entity"
)

type ActivityProjection interface {
	AppendActivity(ctx context.Context, entry activityentity.ActivityEntry) error
	ListOrderActivity(ctx context.Context, storeID string, orderID string) ([]activityentity.ActivityEntry, error)
}
