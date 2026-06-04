package storeaccess

import (
	"context"
	"fmt"
	"strings"

	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
)

type Access interface {
	ResolveStore(ctx context.Context, storeID string) (*storectx.Store, error)
}

type Service struct {
	repo storectx.StoreRepository
}

var _ Access = (*Service)(nil)

func New(repo storectx.StoreRepository) Access {
	return &Service{repo: repo}
}

func (s *Service) ResolveStore(ctx context.Context, storeID string) (*storectx.Store, error) {
	storeID = strings.TrimSpace(storeID)
	if storeID == "" {
		return nil, fmt.Errorf("store id is required")
	}
	store, err := s.repo.FindByID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, fmt.Errorf("store not found")
	}
	return store, nil
}
