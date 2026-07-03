package outputport

import (
	"context"

	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

type AccessAuthorizer interface {
	AuthorizeStoreRequest(ctx context.Context, workspaceID string, requestedBy string) error
	AuthorizeStoreRead(ctx context.Context, workspaceID string, requestedBy string) error
	AuthorizeStoreApproval(ctx context.Context, requestedBy string) error
}

type OperationalStoreFinalizer interface {
	FinalizeStore(ctx context.Context, request storeentity.StoreRequest) error
}

type StoreRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindBySubdomain(ctx context.Context, subdomain string) (*storeentity.StoreRequest, error)
	Create(ctx context.Context, request storeentity.StoreRequest) (*storeentity.StoreRequest, error)
	FindByID(ctx context.Context, id string) (*storeentity.StoreRequest, error)
	ListPage(
		ctx context.Context,
		workspaceID string,
		query collection.Query,
	) (collection.Page[storeentity.StoreRequest], error)
	ClaimNextQueued(ctx context.Context) (*storeentity.StoreRequest, error)
	FindNextProvisioning(ctx context.Context) (*storeentity.StoreRequest, error)
	UpdateStatus(ctx context.Context, id string, status storeentity.RequestStatus) error
	MarkReady(ctx context.Context, id string, storeID string) error
	MarkFailed(ctx context.Context, id string, reason string) error
	MarkBlocked(ctx context.Context, id string, status storeentity.RequestStatus, reason string) error
	RecordTransition(ctx context.Context, transition storeentity.StoreRequestTransition) error
}
