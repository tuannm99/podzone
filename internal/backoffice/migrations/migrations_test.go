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

// PZEP-0017 / MT10: InvalidateScope is pdtenantdb.Config.OnScopeInvalidated's
// wiring target - it must evict the matching appliedScopes entry so the
// next ApplyTx call for that scope re-verifies instead of trusting a
// permanently-cached "done" bit.
func TestInvalidateScope_EvictsMatchingEntry(t *testing.T) {
	appliedScopes.Store("some_db|t_some_schema", struct{}{})

	InvalidateScope("some_db", "t_some_schema")

	_, stillCached := appliedScopes.Load("some_db|t_some_schema")
	require.False(t, stillCached)
}

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

// PZEP-0015: MT1 Phase 2. Migration 0018 must provision a role per tenant
// schema, scoped to exactly that schema, with backoffice_app as a member
// (so WithTenantTx can SET LOCAL ROLE into it) - and two tenants' roles
// must not be able to see each other's data. This is the concrete "mutual
// invisibility" test MT16 flagged as missing.
func TestApplyTxCreatesPerTenantRoleWithMutualIsolation(t *testing.T) {
	if _, ok := os.LookupEnv("XDG_RUNTIME_DIR"); !ok {
		t.Skip("docker-backed integration test requires XDG_RUNTIME_DIR")
	}
	info := testkit.PostgresInfo(t)
	registry := pdtenantdbmocks.NewMockClusterRegistry(t)
	registry.EXPECT().GetCluster(mock.Anything, "pg-01").Return(pdtenantdb.ClusterConfig{
		Host:     info.Host,
		Port:     info.Port,
		User:     info.User,
		Password: info.Password,
		SSLMode:  "disable",
	}, nil).Maybe()

	ctx := context.Background()

	provision := func(tenantID string) string {
		schemaName := fmt.Sprintf("t_role_%s_%d", tenantID, time.Now().UnixNano())
		resolver := pdtenantdbmocks.NewMockPlacementResolver(t)
		resolver.EXPECT().Resolve(mock.Anything, tenantID).Return(pdtenantdb.Placement{
			TenantID:    tenantID,
			ClusterName: "pg-01",
			Mode:        pdtenantdb.ModeSchema,
			DBName:      info.DBName,
			SchemaName:  schemaName,
		}, nil).Maybe()

		manager := pdtenantdb.NewManager(
			&pdtenantdb.Config{SharedDB: info.DBName, OnSchemaBootstrap: ApplyToSchema},
			resolver,
			registry,
		)
		t.Cleanup(func() { _ = manager.CloseAll() })

		err := manager.WithTenantTx(ctx, tenantID, &sql.TxOptions{}, func(tx *sqlx.Tx) error {
			_, err := tx.ExecContext(ctx, `
				INSERT INTO stores (id, name, owner_id, status, created_at, updated_at)
				VALUES ($1, 'x', 'owner', 'active', now(), now())
			`, tenantID+"-store")
			return err
		})
		require.NoError(t, err)
		return schemaName
	}

	schemaA := provision("tenant-role-a")
	schemaB := provision("tenant-role-b")

	admin := testkit.PostgresDB(t)

	// AC-1/AC-2 (role provisioning): role exists per schema, least-privilege,
	// and backoffice_app is a member (so it can SET LOCAL ROLE into it).
	for _, schema := range []string{schemaA, schemaB} {
		var role struct {
			CanLogin   bool `db:"rolcanlogin"`
			Super      bool `db:"rolsuper"`
			CreateDB   bool `db:"rolcreatedb"`
			CreateRole bool `db:"rolcreaterole"`
		}
		err := admin.GetContext(ctx, &role, `
			SELECT rolcanlogin, rolsuper, rolcreatedb, rolcreaterole
			FROM pg_authid WHERE rolname = $1
		`, schema)
		require.NoError(t, err)
		require.False(t, role.CanLogin, "tenant role must be NOLOGIN")
		require.False(t, role.Super)
		require.False(t, role.CreateDB)
		require.False(t, role.CreateRole)

		var isMember bool
		err = admin.GetContext(ctx, &isMember, `
			SELECT EXISTS (
				SELECT 1 FROM pg_auth_members m
				JOIN pg_roles r ON r.oid = m.roleid
				JOIN pg_roles g ON g.oid = m.member
				WHERE r.rolname = $1 AND g.rolname = 'backoffice_app'
			)
		`, schema)
		require.NoError(t, err)
		require.True(t, isMember, "backoffice_app must be a member of the tenant role")
	}

	// AC-4: the tenant role can perform CRUD against its own schema.
	func() {
		tx, err := admin.BeginTxx(ctx, nil)
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()
		_, err = tx.ExecContext(ctx, fmt.Sprintf(`SET LOCAL ROLE %s`, pdtenantdb.QuoteIdent(schemaA)))
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx, fmt.Sprintf(`SET LOCAL search_path TO %s`, pdtenantdb.QuoteIdent(schemaA)))
		require.NoError(t, err)
		var count int
		err = tx.GetContext(ctx, &count, `SELECT COUNT(*) FROM stores`)
		require.NoError(t, err, "tenant role must be able to SELECT its own schema's tables")
		require.Equal(t, 1, count)
	}()

	// AC-3: mutual invisibility - tenant A's role must not be able to read
	// tenant B's schema, even though backoffice_app itself (Phase 1) has
	// grants on both.
	func() {
		tx, err := admin.BeginTxx(ctx, nil)
		require.NoError(t, err)
		defer func() { _ = tx.Rollback() }()
		_, err = tx.ExecContext(ctx, fmt.Sprintf(`SET LOCAL ROLE %s`, pdtenantdb.QuoteIdent(schemaA)))
		require.NoError(t, err)
		_, err = tx.ExecContext(ctx, fmt.Sprintf(`SET LOCAL search_path TO %s`, pdtenantdb.QuoteIdent(schemaB)))
		require.NoError(t, err)
		var count int
		err = tx.GetContext(ctx, &count, `SELECT COUNT(*) FROM stores`)
		require.Error(t, err, "tenant A's role must not be able to read tenant B's schema")
	}()
}
