package pdtenantdb_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	"github.com/tuannm99/podzone/pkg/testkit"
)

type staticResolver struct {
	mu         sync.Mutex
	placements map[string]pdtenantdb.Placement
}

func (r *staticResolver) Resolve(ctx context.Context, tenantID string) (pdtenantdb.Placement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pl, ok := r.placements[tenantID]
	if !ok {
		return pdtenantdb.Placement{}, fmt.Errorf("missing placement for %s", tenantID)
	}
	return pl, nil
}

type staticRegistry struct {
	cfg pdtenantdb.ClusterConfig
}

func (r *staticRegistry) GetCluster(ctx context.Context, clusterName string) (pdtenantdb.ClusterConfig, error) {
	return r.cfg, nil
}

func setupManager(t *testing.T, cfg *pdtenantdb.Config, placements map[string]pdtenantdb.Placement) pdtenantdb.Manager {
	t.Helper()
	info := testkit.PostgresInfo(t)
	reg := &staticRegistry{cfg: pdtenantdb.ClusterConfig{
		Host:     info.Host,
		Port:     info.Port,
		User:     info.User,
		Password: info.Password,
		SSLMode:  "disable",
	}}
	res := &staticResolver{placements: placements}
	return pdtenantdb.NewManager(cfg, res, reg)
}

func TestManager_DBForTenant_ReusesPoolByClusterAndDB(t *testing.T) {
	cfg := &pdtenantdb.Config{
		SharedDB:          testkit.PostgresInfo(t).DBName,
		ConnMaxLifetime:   1 * time.Minute,
		MaxOpenConns:      10,
		MaxIdleConns:      2,
		MaxDedicatedPools: 200,
		DedicatedIdleTTL:  30 * time.Minute,
	}

	placements := map[string]pdtenantdb.Placement{
		"t1": {
			TenantID:    "t1",
			ClusterName: "pg-01",
			Mode:        pdtenantdb.ModeSchema,
			DBName:      cfg.SharedDB,
			SchemaName:  "t_t1",
		},
	}

	m := setupManager(t, cfg, placements)

	db1, pl1, err := m.DBForTenant(context.Background(), "t1")
	require.NoError(t, err)
	require.NotNil(t, db1)
	require.Equal(t, cfg.SharedDB, pl1.DBName)

	db2, pl2, err := m.DBForTenant(context.Background(), "t1")
	require.NoError(t, err)
	require.Same(t, db1, db2)
	require.Equal(t, pl1.DBName, pl2.DBName)
}

func TestManager_WithTenantTx_SetsSearchPathForSchemaMode(t *testing.T) {
	cfg := &pdtenantdb.Config{SharedDB: testkit.PostgresInfo(t).DBName}

	placements := map[string]pdtenantdb.Placement{
		"t1": {
			TenantID:    "t1",
			ClusterName: "pg-01",
			Mode:        pdtenantdb.ModeSchema,
			DBName:      cfg.SharedDB,
			SchemaName:  "t_t1",
		},
	}

	m := setupManager(t, cfg, placements)

	err := m.WithTenantTx(context.Background(), "t1", &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		var searchPath string
		if err := tx.Get(&searchPath, "SHOW search_path"); err != nil {
			return err
		}
		if searchPath == "" {
			return fmt.Errorf("empty search_path")
		}
		if !strings.Contains(searchPath, "t_t1") {
			return fmt.Errorf("search_path missing schema: %s", searchPath)
		}
		return nil
	})
	require.NoError(t, err)
}

func TestManager_DedicatedPoolCapacity(t *testing.T) {
	cfg := &pdtenantdb.Config{
		SharedDB:          testkit.PostgresInfo(t).DBName,
		MaxDedicatedPools: 1,
	}

	testkit.EnsurePostgresDB(t, "bo_a")
	testkit.EnsurePostgresDB(t, "bo_b")

	placements := map[string]pdtenantdb.Placement{
		"a": {TenantID: "a", ClusterName: "pg-01", Mode: pdtenantdb.ModeDatabase, DBName: "bo_a"},
		"b": {TenantID: "b", ClusterName: "pg-01", Mode: pdtenantdb.ModeDatabase, DBName: "bo_b"},
	}

	m := setupManager(t, cfg, placements)

	_, _, err := m.DBForTenant(context.Background(), "a")
	require.NoError(t, err)

	_, _, err = m.DBForTenant(context.Background(), "b")
	require.Error(t, err)
	require.Contains(t, err.Error(), "dedicated pool capacity")
}
