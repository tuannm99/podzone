package store

import "context"

type StoreUsecase interface {
	ListStores(ctx context.Context, query ListStoresQuery) ([]Store, error)
	GetStore(ctx context.Context, query GetStoreQuery) (*Store, error)
	CreateStoreFromCommand(ctx context.Context, cmd CreateStoreCmd) (*Store, error)
	UpdateStoreStatusFromCommand(ctx context.Context, cmd UpdateStoreStatusCmd) (*Store, error)

	// Compatibility methods for the current GraphQL surface.
	GetAllStores(ctx context.Context) ([]Store, error)
	GetStoreByID(ctx context.Context, id string) (*Store, error)
	CreateStore(ctx context.Context, name, description string) (*Store, error)
	UpdateStoreStatus(ctx context.Context, id string, active bool) (*Store, error)
}
