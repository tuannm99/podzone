package tenancy

import (
	"context"
	"fmt"
	"strings"

	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/storeaccess"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
)

type Readiness interface {
	EnsureReady(ctx context.Context, tenantID string) error
}

type Runtime interface {
	ResolveRequestScope(ctx context.Context, tenantID, storeID string) (RequestScope, error)
}

type RequestScope struct {
	TenantID  string
	Placement *pdtenantdb.Placement
	Store     *storectx.Store
}

type Service struct {
	readiness   Readiness
	storeAccess storeaccess.Access
	placements  pdtenantdb.PlacementResolver
}

var _ Runtime = (*Service)(nil)

func New(
	readiness Readiness,
	storeAccess storeaccess.Access,
	placements pdtenantdb.PlacementResolver,
) Runtime {
	return &Service{
		readiness:   readiness,
		storeAccess: storeAccess,
		placements:  placements,
	}
}

func (s *Service) ResolveRequestScope(
	ctx context.Context,
	tenantID, storeID string,
) (RequestScope, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return RequestScope{}, fmt.Errorf("tenant id is required")
	}
	if s.readiness == nil {
		return RequestScope{}, fmt.Errorf("tenant readiness runtime is not configured")
	}
	if err := s.readiness.EnsureReady(ctx, tenantID); err != nil {
		return RequestScope{}, err
	}

	scope := RequestScope{TenantID: tenantID}
	if s.placements != nil {
		placement, err := s.placements.Resolve(ctx, tenantID)
		if err != nil {
			return RequestScope{}, err
		}
		scope.Placement = &placement
	}

	storeID = strings.TrimSpace(storeID)
	if storeID == "" {
		return scope, nil
	}
	if s.storeAccess == nil {
		return RequestScope{}, fmt.Errorf("store access runtime is not configured")
	}
	store, err := s.storeAccess.ResolveStore(ctx, storeID)
	if err != nil {
		return RequestScope{}, err
	}
	scope.Store = store
	return scope, nil
}
