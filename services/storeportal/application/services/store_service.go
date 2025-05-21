package services

import (
	"context"
	"errors"

	"github.com/tuannm99/podzone/services/storeportal/domain/entities"
	"github.com/tuannm99/podzone/services/storeportal/domain/repositories"
)

var (
	ErrStoreNotFound = errors.New("store not found")
	ErrInvalidStore  = errors.New("invalid store data")
)

// StoreService handles store-related business logic
type StoreService struct {
	storeRepo repositories.StoreRepository
}

// CreateStore creates a new store
func (s *StoreService) CreateStore(ctx context.Context, name, description string) (*entities.Store, error) {
	if name == "" {
		return nil, ErrInvalidStore
	}

	store := entities.NewStore(name, description)
	if err := s.storeRepo.Save(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}

// GetStore retrieves a store by ID
func (s *StoreService) GetStore(ctx context.Context, id string) (*entities.Store, error) {
	store, err := s.storeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, ErrStoreNotFound
	}
	return store, nil
}

// ListStores retrieves all stores
func (s *StoreService) ListStores(ctx context.Context) ([]*entities.Store, error) {
	return s.storeRepo.List(ctx)
}

// ActivateStore activates a store
func (s *StoreService) ActivateStore(ctx context.Context, id string) error {
	store, err := s.GetStore(ctx, id)
	if err != nil {
		return err
	}

	store.Activate()
	return s.storeRepo.Save(ctx, store)
}

// DeactivateStore deactivates a store
func (s *StoreService) DeactivateStore(ctx context.Context, id string) error {
	store, err := s.GetStore(ctx, id)
	if err != nil {
		return err
	}

	store.Deactivate()
	return s.storeRepo.Save(ctx, store)
}
