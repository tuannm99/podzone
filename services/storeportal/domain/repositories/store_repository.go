package repositories

import (
	"context"

	"github.com/tuannm99/podzone/services/storeportal/domain/entities"
)

// StoreRepository defines the interface for store persistence
type StoreRepository interface {
	// FindByID retrieves a store by its ID
	FindByID(ctx context.Context, id string) (*entities.Store, error)

	// Save persists a store
	Save(ctx context.Context, store *entities.Store) error

	// List retrieves all stores
	List(ctx context.Context) ([]*entities.Store, error)

	// Delete removes a store
	Delete(ctx context.Context, id string) error
}
