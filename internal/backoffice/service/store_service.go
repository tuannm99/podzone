package service

import (
	"context"
	"errors"

	"github.com/tuannm99/podzone/internal/backoffice/models"
	"github.com/tuannm99/podzone/internal/backoffice/repository"
)

var ErrUnauthorized = errors.New("unauthorized access")

// StoreService handles business logic for store operations
type StoreService struct {
	storeRepo *repository.StoreRepository
}

// NewStoreService creates a new store service
func NewStoreService(storeRepo *repository.StoreRepository) *StoreService {
	return &StoreService{
		storeRepo: storeRepo,
	}
}

// CreateStore creates a new store
func (s *StoreService) CreateStore(ctx context.Context, name, description string) (*models.Store, error) {
	store := models.NewStore(name, description)

	if err := s.storeRepo.Create(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}

// GetStore retrieves a store by ID
func (s *StoreService) GetStore(ctx context.Context, id string) (*models.Store, error) {
	store, err := s.storeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return store, nil
}

// ListStores retrieves all stores for the current tenant
func (s *StoreService) ListStores(ctx context.Context) ([]*models.Store, error) {
	return s.storeRepo.List(ctx)
}

// ActivateStore activates a store
func (s *StoreService) ActivateStore(ctx context.Context, id string) (*models.Store, error) {
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
func (s *StoreService) DeactivateStore(ctx context.Context, id string) (*models.Store, error) {
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
