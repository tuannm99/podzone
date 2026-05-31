package outputport

import (
	"context"

	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
)

type StoreRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindBySubdomain(ctx context.Context, subdomain string) (*storeentity.StoreRequest, error)
	Create(ctx context.Context, request storeentity.StoreRequest) (*storeentity.StoreRequest, error)
	FindByID(ctx context.Context, id string) (*storeentity.StoreRequest, error)
	List(ctx context.Context, workspaceID string) ([]storeentity.StoreRequest, error)
	ClaimNextQueued(ctx context.Context) (*storeentity.StoreRequest, error)
	UpdateStatus(ctx context.Context, id string, status storeentity.RequestStatus) error
	MarkReady(ctx context.Context, id string, storeID string) error
	MarkFailed(ctx context.Context, id string, reason string) error
}
