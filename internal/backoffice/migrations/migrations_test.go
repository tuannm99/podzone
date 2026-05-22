package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	pdtenantdbmocks "github.com/tuannm99/podzone/pkg/pdtenantdb/mocks"
	"github.com/tuannm99/podzone/pkg/testkit"
)

func TestApplyTxConcurrentForFreshTenant(t *testing.T) {
	if _, ok := os.LookupEnv("XDG_RUNTIME_DIR"); !ok {
		t.Skip("docker-backed integration test requires XDG_RUNTIME_DIR")
	}
	info := testkit.PostgresInfo(t)
	tenantID := "tenant-migration-race"
	schemaName := fmt.Sprintf("t_migration_race_%d", time.Now().UnixNano())
	resolver := pdtenantdbmocks.NewMockPlacementResolver(t)
	resolver.EXPECT().Resolve(mock.Anything, tenantID).Return(pdtenantdb.Placement{
		TenantID:    tenantID,
		ClusterName: "pg-01",
		Mode:        pdtenantdb.ModeSchema,
		DBName:      info.DBName,
		SchemaName:  schemaName,
	}, nil).Maybe()
	registry := pdtenantdbmocks.NewMockClusterRegistry(t)
	registry.EXPECT().GetCluster(mock.Anything, "pg-01").Return(pdtenantdb.ClusterConfig{
		Host:     info.Host,
		Port:     info.Port,
		User:     info.User,
		Password: info.Password,
		SSLMode:  "disable",
	}, nil).Maybe()

	manager := pdtenantdb.NewManager(
		&pdtenantdb.Config{SharedDB: info.DBName},
		resolver,
		registry,
	)
	t.Cleanup(func() { _ = manager.CloseAll() })

	ctx := context.Background()
	errs := make(chan error, 8)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- manager.WithTenantTx(ctx, tenantID, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
				return ApplyTx(ctx, tx)
			})
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}

	err := manager.WithTenantTx(ctx, tenantID, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		var versions []string
		if err := tx.SelectContext(ctx, &versions, `SELECT version FROM backoffice_schema_migrations ORDER BY version`); err != nil {
			return err
		}

		entries, err := migrationsFS.ReadDir("sql")
		if err != nil {
			return err
		}
		expected := make([]string, 0, len(entries))
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
				continue
			}
			expected = append(expected, strings.TrimSuffix(entry.Name(), ".sql"))
		}

		require.Equal(t, expected, versions)

		var count int
		if err := tx.GetContext(ctx, &count, `SELECT COUNT(*) FROM routed_order_activities`); err != nil {
			return err
		}
		require.Equal(t, 0, count)
		return nil
	})
	require.NoError(t, err)
}
