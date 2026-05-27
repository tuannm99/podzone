package tenancy

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	storeentity "github.com/tuannm99/podzone/internal/backoffice/domain/store/entity"
	"github.com/tuannm99/podzone/internal/backoffice/runtime/storeaccess"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
)

type stubReadiness struct {
	calls     []string
	returnErr error
}

func (s *stubReadiness) EnsureReady(_ context.Context, tenantID string) error {
	s.calls = append(s.calls, tenantID)
	return s.returnErr
}

type stubStoreAccess struct {
	lastStoreID string
	store       *storeentity.Store
	returnErr   error
}

func (s *stubStoreAccess) ResolveStore(_ context.Context, storeID string) (*storeentity.Store, error) {
	s.lastStoreID = storeID
	return s.store, s.returnErr
}

type stubPlacementResolver struct {
	lastTenantID string
	placement    pdtenantdb.Placement
	returnErr    error
}

func (s *stubPlacementResolver) Resolve(_ context.Context, tenantID string) (pdtenantdb.Placement, error) {
	s.lastTenantID = tenantID
	return s.placement, s.returnErr
}

var (
	_ Readiness                    = (*stubReadiness)(nil)
	_ storeaccess.Access           = (*stubStoreAccess)(nil)
	_ pdtenantdb.PlacementResolver = (*stubPlacementResolver)(nil)
)

func TestResolveRequestScopeIncludesPlacementAndStore(t *testing.T) {
	t.Parallel()

	readiness := &stubReadiness{}
	stores := &stubStoreAccess{
		store: &storeentity.Store{ID: "store-ops", Name: "Ops"},
	}
	placements := &stubPlacementResolver{
		placement: pdtenantdb.Placement{
			TenantID:    "tenant-ops",
			ClusterName: "pg-default",
			Mode:        pdtenantdb.ModeSchema,
			DBName:      "backoffice",
			SchemaName:  "t_tenant_ops",
		},
	}

	runtime := New(readiness, stores, placements)
	scope, err := runtime.ResolveRequestScope(context.Background(), "tenant-ops", "store-ops")
	require.NoError(t, err)
	require.Equal(t, []string{"tenant-ops"}, readiness.calls)
	require.Equal(t, "tenant-ops", placements.lastTenantID)
	require.Equal(t, "store-ops", stores.lastStoreID)
	require.Equal(t, "tenant-ops", scope.TenantID)
	require.NotNil(t, scope.Placement)
	require.Equal(t, "pg-default", scope.Placement.ClusterName)
	require.NotNil(t, scope.Store)
	require.Equal(t, "store-ops", scope.Store.ID)
}

func TestResolveRequestScopeReturnsReadinessError(t *testing.T) {
	t.Parallel()

	readiness := &stubReadiness{returnErr: errors.New("tenant not ready")}
	runtime := New(readiness, nil, nil)

	_, err := runtime.ResolveRequestScope(context.Background(), "tenant-ops", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "tenant not ready")
}
