package inputport

import (
	"context"

	storeentity "github.com/tuannm99/podzone/internal/backoffice/domain/store/entity"
)

type StoreUsecase interface {
	GetAllStores(ctx context.Context) ([]storeentity.Store, error)
	GetStoreByID(ctx context.Context, id string) (*storeentity.Store, error)
	CreateStore(ctx context.Context, name, description string) (*storeentity.Store, error)
	UpdateStoreStatus(ctx context.Context, id string, active bool) (*storeentity.Store, error)
}
