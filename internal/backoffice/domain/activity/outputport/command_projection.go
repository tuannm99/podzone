package outputport

import (
	"context"

	activityentity "github.com/tuannm99/podzone/internal/backoffice/domain/activity/entity"
)

type ActivityCommandProjection interface {
	AppendActivity(ctx context.Context, entry activityentity.ActivityEntry) error
}
