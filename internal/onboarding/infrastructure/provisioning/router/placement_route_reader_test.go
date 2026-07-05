package router

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
	kvsmocks "github.com/tuannm99/podzone/pkg/toolkit/kvstores/mocks"
)

func TestPlacementRouteReader_IsPlacementRouteReady(t *testing.T) {
	t.Parallel()

	kv := kvsmocks.NewMockKVStore(t)
	reader := NewPlacementRouteReader(PlacementRouteReaderParams{KV: kv})

	kv.EXPECT().
		Get("podzone/tenants/tenant-1/placement").
		Return([]byte(`{"cluster_name":"pg-default"}`), nil)

	ready, err := reader.IsPlacementRouteReady(context.Background(), "tenant-1")
	require.NoError(t, err)
	require.True(t, ready)
}

func TestPlacementRouteReader_IsPlacementRouteReadyReturnsFalseWhenMissing(t *testing.T) {
	t.Parallel()

	kv := kvsmocks.NewMockKVStore(t)
	reader := NewPlacementRouteReader(PlacementRouteReaderParams{KV: kv})

	kv.EXPECT().
		Get("podzone/tenants/tenant-1/placement").
		Return(nil, kvstores.ErrKeyNotFound)

	ready, err := reader.IsPlacementRouteReady(context.Background(), "tenant-1")
	require.NoError(t, err)
	require.False(t, ready)
}

func TestPlacementRouteReader_IsPlacementRouteReadyWrapsBackendError(t *testing.T) {
	t.Parallel()

	kv := kvsmocks.NewMockKVStore(t)
	reader := NewPlacementRouteReader(PlacementRouteReaderParams{KV: kv})
	backendErr := errors.New("kv store down")

	kv.EXPECT().
		Get("podzone/tenants/tenant-1/placement").
		Return(nil, backendErr)

	ready, err := reader.IsPlacementRouteReady(context.Background(), "tenant-1")
	require.ErrorIs(t, err, backendErr)
	require.False(t, ready)
}

func TestPlacementRouteReader_PublishPlacementRoute(t *testing.T) {
	t.Parallel()

	kv := kvsmocks.NewMockKVStore(t)
	reader := NewPlacementRouteReader(PlacementRouteReaderParams{KV: kv})

	kv.EXPECT().
		Put(
			"podzone/tenants/tenant-1/placement",
			[]byte(`{"cluster_name":"pg-default","db_name":"podzone_tenants","mode":"schema","schema_name":"t_tenant_1"}`),
		).
		Return(nil)

	err := reader.PublishPlacementRoute(context.Background(), "tenant-1", entity.PlacementAllocation{
		ClusterName: "pg-default",
		Mode:        "schema",
		DBName:      "podzone_tenants",
		SchemaName:  "t_tenant_1",
	})
	require.NoError(t, err)
}

func TestPlacementRouteReader_PublishPlacementRoutePublishesClusterRegistry(t *testing.T) {
	t.Parallel()

	kv := kvsmocks.NewMockKVStore(t)
	reader := NewPlacementRouteReader(PlacementRouteReaderParams{
		KV: kv,
		Config: onboardingconfig.StoreProvisioningConfig{
			AdminDSN: "postgres://postgres:secret@postgres:5432/postgres?sslmode=disable",
		},
	})

	kv.EXPECT().
		Put(
			"podzone/tenants/tenant-1/placement",
			[]byte(`{"cluster_name":"pg-default","db_name":"podzone_tenants","mode":"schema","schema_name":"t_tenant_1"}`),
		).
		Return(nil)
	kv.EXPECT().
		Put(
			"podzone/postgres/clusters/pg-default",
			[]byte(`{"host":"postgres","password":"secret","port":5432,"ssl_mode":"disable","user":"postgres"}`),
		).
		Return(nil)

	err := reader.PublishPlacementRoute(context.Background(), "tenant-1", entity.PlacementAllocation{
		ClusterName: "pg-default",
		Mode:        "schema",
		DBName:      "podzone_tenants",
		SchemaName:  "t_tenant_1",
	})
	require.NoError(t, err)
}
