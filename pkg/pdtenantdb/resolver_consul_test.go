package pdtenantdb_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	"github.com/tuannm99/podzone/pkg/toolkit/kvstores"
	kvsmocks "github.com/tuannm99/podzone/pkg/toolkit/kvstores/mocks"
)

func TestConsulPlacementResolver_CacheHit(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	r := pdtenantdb.NewConsulPlacementResolver(kv)

	key := "podzone/tenants/tenant-abc/placement"
	val := []byte(`{"cluster_name":"pg-01","mode":"schema","db_name":"backoffice","schema_name":"t_tenant_abc"}`)

	kv.EXPECT().Get(key).Return(val, nil).Once()

	ctx := context.Background()

	pl1, err := r.Resolve(ctx, "tenant-abc")
	require.NoError(t, err)
	require.Equal(t, "pg-01", pl1.ClusterName)
	require.Equal(t, pdtenantdb.ModeSchema, pl1.Mode)
	require.Equal(t, "backoffice", pl1.DBName)
	require.Equal(t, "t_tenant_abc", pl1.SchemaName)

	// Second call must be served from cache — no additional KV.Get.
	pl2, err := r.Resolve(ctx, "tenant-abc")
	require.NoError(t, err)
	require.Equal(t, pl1, pl2)
}

func TestConsulPlacementResolver_DatabaseMode(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	r := pdtenantdb.NewConsulPlacementResolver(kv)

	key := "podzone/tenants/big-tenant/placement"
	val := []byte(`{"cluster_name":"pg-02","mode":"database","db_name":"bo_big_tenant"}`)

	kv.EXPECT().Get(key).Return(val, nil).Once()

	pl, err := r.Resolve(context.Background(), "big-tenant")
	require.NoError(t, err)
	require.Equal(t, pdtenantdb.ModeDatabase, pl.Mode)
	require.Equal(t, "bo_big_tenant", pl.DBName)
	require.Empty(t, pl.SchemaName)
}

func TestConsulPlacementResolver_NotFound(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	r := pdtenantdb.NewConsulPlacementResolver(kv)

	key := "podzone/tenants/unknown/placement"
	kv.EXPECT().Get(key).Return(nil, kvstores.ErrKeyNotFound).Once()

	_, err := r.Resolve(context.Background(), "unknown")
	require.Error(t, err)
	require.ErrorIs(t, err, pdtenantdb.ErrPlacementNotFound)
}

func TestConsulPlacementResolver_BackendError(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	r := pdtenantdb.NewConsulPlacementResolver(kv)

	key := "podzone/tenants/down/placement"
	kv.EXPECT().Get(key).Return(nil, context.DeadlineExceeded).Once()

	_, err := r.Resolve(context.Background(), "down")
	require.Error(t, err)
	require.ErrorIs(t, err, pdtenantdb.ErrPlacementBackend)
}

func TestConsulPlacementResolver_InvalidJSON(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	r := pdtenantdb.NewConsulPlacementResolver(kv)

	key := "podzone/tenants/bad/placement"
	kv.EXPECT().Get(key).Return([]byte(`{invalid`), nil).Once()

	_, err := r.Resolve(context.Background(), "bad")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid placement json")
}

func TestConsulPlacementResolver_MissingFields(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	r := pdtenantdb.NewConsulPlacementResolver(kv)

	key := "podzone/tenants/partial/placement"
	kv.EXPECT().Get(key).Return([]byte(`{"mode":"schema"}`), nil).Once()

	_, err := r.Resolve(context.Background(), "partial")
	require.Error(t, err)
	require.Contains(t, err.Error(), "incomplete placement")
}

func TestConsulPlacementResolver_SchemaModeMissingSchemaName(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	r := pdtenantdb.NewConsulPlacementResolver(kv)

	key := "podzone/tenants/no-schema/placement"
	kv.EXPECT().Get(key).Return([]byte(`{"cluster_name":"pg-01","mode":"schema","db_name":"backoffice"}`), nil).Once()

	_, err := r.Resolve(context.Background(), "no-schema")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing schema_name")
}

func TestConsulPlacementResolver_SingleflightConcurrent(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	r := pdtenantdb.NewConsulPlacementResolver(kv)

	key := "podzone/tenants/concurrent/placement"
	val := []byte(`{"cluster_name":"pg-01","mode":"schema","db_name":"backoffice","schema_name":"t_concurrent"}`)

	// Must be called only once despite concurrent resolves.
	kv.EXPECT().Get(key).Return(val, nil).Once()

	ctx := context.Background()
	var wg sync.WaitGroup
	errs := make(chan error, 20)

	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := r.Resolve(ctx, "concurrent")
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
}

func TestConsulPlacementResolver_TTLExpiry(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)

	// Use a very short TTL so we can test expiry without sleeping long.
	r := pdtenantdb.NewConsulPlacementResolverWithTTL(kv, "podzone/tenants", 50*time.Millisecond)

	key := "podzone/tenants/ttl-tenant/placement"
	val := []byte(`{"cluster_name":"pg-01","mode":"schema","db_name":"backoffice","schema_name":"t_ttl"}`)

	// Called twice: once on first resolve, once after TTL expires.
	kv.EXPECT().Get(key).Return(val, nil).Twice()

	ctx := context.Background()

	_, err := r.Resolve(ctx, "ttl-tenant")
	require.NoError(t, err)

	time.Sleep(60 * time.Millisecond)

	_, err = r.Resolve(ctx, "ttl-tenant")
	require.NoError(t, err)
}
