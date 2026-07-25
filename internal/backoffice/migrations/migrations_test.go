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

// PZEP-0009 / MT1 Phase 1: migration 0016 must provision backoffice_app as
// a genuinely least-privilege role - no createdb/createrole/superuser.
func TestApplyTxCreatesLeastPrivilegeAppRole(t *testing.T) {
	if _, ok := os.LookupEnv("XDG_RUNTIME_DIR"); !ok {
		t.Skip("docker-backed integration test requires XDG_RUNTIME_DIR")
	}
	info := testkit.PostgresInfo(t)
	tenantID := "tenant-app-role"
	schemaName := fmt.Sprintf("t_app_role_%d", time.Now().UnixNano())
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
		&pdtenantdb.Config{SharedDB: info.DBName, OnSchemaBootstrap: ApplyToSchema},
		resolver,
		registry,
	)
	t.Cleanup(func() { _ = manager.CloseAll() })

	ctx := context.Background()
	err := manager.WithTenantTx(ctx, tenantID, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		var exists bool
		return tx.GetContext(ctx, &exists, `SELECT true`)
	})
	require.NoError(t, err)

	admin := testkit.PostgresDB(t)
	var role struct {
		CanLogin    bool `db:"rolcanlogin"`
		Super       bool `db:"rolsuper"`
		CreateDB    bool `db:"rolcreatedb"`
		CreateRole  bool `db:"rolcreaterole"`
		HasPassword bool `db:"has_password"`
	}
	err = admin.GetContext(ctx, &role, `
		SELECT rolcanlogin, rolsuper, rolcreatedb, rolcreaterole, rolpassword IS NOT NULL AS has_password
		FROM pg_authid WHERE rolname = 'backoffice_app'
	`)
	require.NoError(t, err)
	require.True(t, role.CanLogin)
	require.False(t, role.Super)
	require.False(t, role.CreateDB)
	require.False(t, role.CreateRole)
	require.False(t, role.HasPassword, "migration must not set a password - see PZEP-0009 Rollout")

	// SET ROLE needs no password (superuser privilege-de-escalation, not
	// authentication) - use it to prove backoffice_app genuinely cannot do
	// admin-only DDL, not just that its pg_authid flags look right.
	tx, err := admin.BeginTxx(ctx, nil)
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()
	_, err = tx.ExecContext(ctx, `SET ROLE backoffice_app`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `CREATE SCHEMA should_not_be_permitted`)
	require.Error(t, err, "backoffice_app must not be able to CREATE SCHEMA")
}
