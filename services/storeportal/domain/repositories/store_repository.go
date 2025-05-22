package repositories

import (
	"context"

	"github.com/tuannm99/podzone/services/storeportal/domain/entities"
)

// StoreRepository defines the interface for store persistence operations
type StoreRepository interface {
	// Create creates a new store
	Create(ctx context.Context, store *entities.Store) error

	// Get retrieves a store by ID
	Get(ctx context.Context, id string) (*entities.Store, error)

	// List retrieves all stores
	List(ctx context.Context) ([]*entities.Store, error)

	// ListByOwnerID retrieves all stores for a specific owner
	ListByOwnerID(ctx context.Context, ownerID string) ([]*entities.Store, error)

	// Update updates an existing store
	Update(ctx context.Context, store *entities.Store) error

	// Delete deletes a store by ID
	Delete(ctx context.Context, id string) error
}
