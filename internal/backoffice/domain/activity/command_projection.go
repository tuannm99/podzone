package activity

import (
	"context"
)

type ActivityCommandProjection interface {
	AppendActivity(ctx context.Context, entry ActivityEntry) error
}
