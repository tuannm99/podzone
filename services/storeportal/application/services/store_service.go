package services

import (
	"context"
	"errors"

	"github.com/tuannm99/podzone/services/storeportal/domain/entities"
	"github.com/tuannm99/podzone/services/storeportal/domain/repositories"
)

var (
	ErrStoreNotFound = errors.New("store not found")
	ErrUnauthorized  = errors.New("unauthorized access")
)

// StoreService handles business logic for store operations
type StoreService struct {
	storeRepo repositories.StoreRepository
}

// NewStoreService creates a new store service
func NewStoreService(storeRepo repositories.StoreRepository) *StoreService {
	return &StoreService{
		storeRepo: storeRepo,
	}
}

// CreateStore creates a new store
func (s *StoreService) CreateStore(ctx context.Context, name, description string) (*entities.Store, error) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	store := entities.NewStore(name, description)
	store.OwnerID = tenantID

	if err := s.storeRepo.Create(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}

// GetStore retrieves a store by ID
func (s *StoreService) GetStore(ctx context.Context, id string) (*entities.Store, error) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	store, err := s.storeRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if store.OwnerID != tenantID {
		return nil, ErrUnauthorized
	}

	return store, nil
}

// ListStores retrieves all stores for the current tenant
func (s *StoreService) ListStores(ctx context.Context) ([]*entities.Store, error) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	return s.storeRepo.ListByOwnerID(ctx, tenantID)
}

// ActivateStore activates a store
func (s *StoreService) ActivateStore(ctx context.Context, id string) (*entities.Store, error) {
	store, err := s.GetStore(ctx, id)
	if err != nil {
		return nil, err
	}

	store.Activate()
	if err := s.storeRepo.Update(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}

// DeactivateStore deactivates a store
func (s *StoreService) DeactivateStore(ctx context.Context, id string) (*entities.Store, error) {
	store, err := s.GetStore(ctx, id)
	if err != nil {
		return nil, err
	}

	store.Deactivate()
	if err := s.storeRepo.Update(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}
