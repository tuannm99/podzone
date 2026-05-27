package storeaccess

import (
	"context"
	"fmt"
	"strings"

	storeentity "github.com/tuannm99/podzone/internal/backoffice/domain/store/entity"
	storeoutputport "github.com/tuannm99/podzone/internal/backoffice/domain/store/outputport"
)

type Access interface {
	ResolveStore(ctx context.Context, storeID string) (*storeentity.Store, error)
}

type Service struct {
	repo storeoutputport.StoreRepository
}

var _ Access = (*Service)(nil)

func New(repo storeoutputport.StoreRepository) Access {
	return &Service{repo: repo}
}

func (s *Service) ResolveStore(ctx context.Context, storeID string) (*storeentity.Store, error) {
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
