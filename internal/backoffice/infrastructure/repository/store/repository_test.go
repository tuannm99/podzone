package store

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	storectx "github.com/tuannm99/podzone/internal/backoffice/domain/store"
	"github.com/tuannm99/podzone/internal/backoffice/migrations"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	pdtenantdbmocks "github.com/tuannm99/podzone/pkg/pdtenantdb/mocks"
	"github.com/tuannm99/podzone/pkg/testkit"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func skipIfDockerUnavailable(t *testing.T) {
	t.Helper()
	if _, ok := os.LookupEnv("XDG_RUNTIME_DIR"); !ok {
		t.Skip("docker-backed integration test requires XDG_RUNTIME_DIR")
	}
}

func newRepositoryTestManager(
	t *testing.T,
	info testkit.PostgresConnInfo,
	placements map[string]pdtenantdb.Placement,
) pdtenantdb.Manager {
	t.Helper()

	registry := pdtenantdbmocks.NewMockClusterRegistry(t)
	registry.EXPECT().GetCluster(mock.Anything, "pg-01").Return(pdtenantdb.ClusterConfig{
		Host:     info.Host,
		Port:     info.Port,
		User:     info.User,
		Password: info.Password,
		SSLMode:  "disable",
	}, nil).Maybe()
	resolver := pdtenantdbmocks.NewMockPlacementResolver(t)
	resolver.EXPECT().
		Resolve(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, tenantID string) (pdtenantdb.Placement, error) {
			pl, ok := placements[tenantID]
			if !ok {
				return pdtenantdb.Placement{}, fmt.Errorf("missing placement for %s", tenantID)
			}
			return pl, nil
		}).
		Maybe()
	return pdtenantdb.NewManager(&pdtenantdb.Config{
		SharedDB:          info.DBName,
		OnSchemaBootstrap: migrations.ApplyToSchema,
	}, resolver, registry)
}

// PZEP-0018 / MT7: FindByID and UpdateStatus must be authorization-neutral -
// access is decided by IAM's RequirePermission before a resolver reaches
// the repository, not by comparing owner_id to the calling user. A tenant
// member who is not a store's owner_id must still be able to find and
// update it.
func TestStoreRepository_FindByIDAndUpdateStatus_IgnoreOwnerID(t *testing.T) {
	skipIfDockerUnavailable(t)
	info := testkit.PostgresInfo(t)
	tenantID := "tenant-store-authz"
	manager := newRepositoryTestManager(t, info, map[string]pdtenantdb.Placement{
		tenantID: {
			TenantID:    tenantID,
			ClusterName: "pg-01",
			Mode:        pdtenantdb.ModeSchema,
			DBName:      info.DBName,
			SchemaName:  fmt.Sprintf("t_store_authz_%d", time.Now().UnixNano()),
		},
	})
	t.Cleanup(func() { _ = manager.CloseAll() })

	repo := New(manager)

	creatorCtx := toolkit.WithTenantID(context.Background(), tenantID)
	creatorCtx = toolkit.WithUserID(creatorCtx, "owner-user")
	store := storectx.Store{
		ID:        "store-1",
		Name:      "Urban Finds",
		OwnerID:   "owner-user",
		Status:    storectx.StoreStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	require.NoError(t, repo.Create(creatorCtx, store))

	// A different tenant member (not owner-user) must still be able to find
	// and update the store - IAM's RequirePermission is what would have
	// gated this call before it ever reached the repository; the repository
	// itself must not re-impose an owner_id check.
	memberCtx := toolkit.WithTenantID(context.Background(), tenantID)
	memberCtx = toolkit.WithUserID(memberCtx, "other-member")

	found, err := repo.FindByID(memberCtx, "store-1")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "store-1", found.ID)
	require.Equal(t, "owner-user", found.OwnerID)

	require.NoError(t, repo.UpdateStatus(memberCtx, "store-1", storectx.StoreStatusInactive))

	updated, err := repo.FindByID(memberCtx, "store-1")
	require.NoError(t, err)
	require.Equal(t, storectx.StoreStatusInactive, updated.Status)
}
