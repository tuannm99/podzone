package pdtenantdb_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	pdtenantdb "github.com/tuannm99/podzone/pkg/pdtenantdb"
	kvsmocks "github.com/tuannm99/podzone/pkg/toolkit/kvstores/mocks"
)

func TestConsulClusterRegistry_CacheHit(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	reg := pdtenantdb.NewConsulClusterRegistry(kv, "podzone/postgres/clusters", 2*time.Minute)

	key := "podzone/postgres/clusters/pg-01"
	val := []byte(`{"host":"pgbouncer.svc","port":6432,"user":"u","password":"p","ssl_mode":"disable"}`)

	kv.EXPECT().Get(key).Return(val, nil).Once()

	ctx := context.Background()

	// First call hits KV
	cfg1, err := reg.GetCluster(ctx, "pg-01")
	require.NoError(t, err)
	require.Equal(t, "pgbouncer.svc", cfg1.Host)

	// Second call should be served from cache (no more KV.Get)
	cfg2, err := reg.GetCluster(ctx, "pg-01")
	require.NoError(t, err)
	require.Equal(t, cfg1, cfg2)
}

func TestConsulClusterRegistry_InvalidJSON(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	reg := pdtenantdb.NewConsulClusterRegistry(kv, "podzone/postgres/clusters", 2*time.Minute)

	key := "podzone/postgres/clusters/pg-01"
	kv.EXPECT().Get(key).Return([]byte(`{invalid`), nil).Once()

	_, err := reg.GetCluster(context.Background(), "pg-01")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid cluster config json")
}

func TestConsulClusterRegistry_MissingHostPort(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	reg := pdtenantdb.NewConsulClusterRegistry(kv, "podzone/postgres/clusters", 2*time.Minute)

	key := "podzone/postgres/clusters/pg-01"
	kv.EXPECT().Get(key).Return([]byte(`{"host":"","port":0}`), nil).Once()

	_, err := reg.GetCluster(context.Background(), "pg-01")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing host/port")
}

func TestConsulClusterRegistry_SingleflightConcurrent(t *testing.T) {
	kv := kvsmocks.NewMockKVStore(t)
	reg := pdtenantdb.NewConsulClusterRegistry(kv, "podzone/postgres/clusters", 2*time.Minute)

	key := "podzone/postgres/clusters/pg-01"
	val := []byte(`{"host":"pgbouncer.svc","port":6432,"user":"u","password":"p","ssl_mode":"disable"}`)

	// Must be called once because of singleflight
	kv.EXPECT().Get(key).Return(val, nil).Once()

	ctx := context.Background()
	var wg sync.WaitGroup

	errs := make(chan error, 20)
	for range 20 {
		wg.Go(func() {
			_, err := reg.GetCluster(ctx, "pg-01")
			errs <- err
		})
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
}
