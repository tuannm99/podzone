package outputport

import (
	"context"

	storeentity "github.com/tuannm99/podzone/internal/backoffice/domain/store/entity"
)

type StoreRepository interface {
	FindAll(ctx context.Context) ([]storeentity.Store, error)
	FindByID(ctx context.Context, id string) (*storeentity.Store, error)
	Create(ctx context.Context, store storeentity.Store) error
	UpdateStatus(ctx context.Context, id string, status string) error
}
