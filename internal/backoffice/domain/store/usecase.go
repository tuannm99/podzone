package store

import (
	"context"

	"github.com/tuannm99/podzone/pkg/collection"
)

type StoreUsecase interface {
	ListStores(ctx context.Context, query ListStoresQuery) (collection.Page[Store], error)
	GetStore(ctx context.Context, query GetStoreQuery) (*Store, error)
	CreateStoreFromCommand(ctx context.Context, cmd CreateStoreCmd) (*Store, error)
	BootstrapStore(ctx context.Context, cmd BootstrapStoreCmd) (*Store, error)
	UpdateStoreStatusFromCommand(ctx context.Context, cmd UpdateStoreStatusCmd) (*Store, error)

	GetStoreByID(ctx context.Context, id string) (*Store, error)
	CreateStore(ctx context.Context, name, description string) (*Store, error)
	UpdateStoreStatus(ctx context.Context, id string, active bool) (*Store, error)
}
