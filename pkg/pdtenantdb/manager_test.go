package pdtenantdb_test

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pdlogmocks "github.com/tuannm99/podzone/pkg/pdlog/mocks"
	"github.com/tuannm99/podzone/pkg/pdtenantdb"
	pdtenantdbmocks "github.com/tuannm99/podzone/pkg/pdtenantdb/mocks"
	"github.com/tuannm99/podzone/pkg/testkit"
)

func setupManager(t *testing.T, cfg *pdtenantdb.Config, placements map[string]pdtenantdb.Placement) pdtenantdb.Manager {
	t.Helper()
	info := testkit.PostgresInfo(t)
	reg := pdtenantdbmocks.NewMockClusterRegistry(t)
	reg.EXPECT().GetCluster(mock.Anything, "pg-01").Return(pdtenantdb.ClusterConfig{
		Host:     info.Host,
		Port:     info.Port,
		User:     info.User,
		Password: info.Password,
		SSLMode:  "disable",
	}, nil).Maybe()
	res := pdtenantdbmocks.NewMockPlacementResolver(t)
	res.EXPECT().
		Resolve(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, tenantID string) (pdtenantdb.Placement, error) {
			pl, ok := placements[tenantID]
			if !ok {
				return pdtenantdb.Placement{}, fmt.Errorf("missing placement for %s", tenantID)
			}
			return pl, nil
		}).
		Maybe()
	return pdtenantdb.NewManager(cfg, res, reg)
}

// PZEP-0017 / MT4: Manager no longer exposes a way to get a raw pool
// directly (DBForTenant removed from the interface) - pool reuse is now
// observed through WithTenantTx, by counting how many times the pool is
// actually opened rather than comparing *sqlx.DB pointers.
func TestManager_WithTenantTx_ReusesPoolByClusterAndDB(t *testing.T) {
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

	original := pdtenantdb.SQLXOpen
	var opens int32
	pdtenantdb.SQLXOpen = func(driverName, dataSourceName string) (*sqlx.DB, error) {
		atomic.AddInt32(&opens, 1)
		return original(driverName, dataSourceName)
	}
	t.Cleanup(func() { pdtenantdb.SQLXOpen = original })

	m := setupManager(t, cfg, placements)

	run := func() {
		err := m.WithTenantTx(context.Background(), "t1", &sql.TxOptions{}, func(tx *sqlx.Tx) error {
			var one int
			return tx.GetContext(context.Background(), &one, "SELECT 1")
		})
		require.NoError(t, err)
	}
	run()
	run()

	require.Equal(t, int32(1), atomic.LoadInt32(&opens), "pool must be opened once and reused, not once per call")
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

// PZEP-0015 (Phase 2 of MT1): with TenantRoleEnabled, WithTenantTx must
// SET LOCAL ROLE into the tenant's own role for the duration of the
// transaction; with it unset (default), behavior must stay byte-for-byte
// identical to Phase 1 (no role switch at all).
func TestManager_WithTenantTx_SetsRoleWhenTenantRoleEnabled(t *testing.T) {
	info := testkit.PostgresInfo(t)
	ctx := context.Background()

	// Hyphen forces QuoteIdent's quoting to actually matter - an unquoted
	// identifier with a hyphen is a syntax error, so this also proves the
	// SET LOCAL ROLE statement isn't just working by accident on a
	// conveniently-plain name.
	schemaName := fmt.Sprintf("t-role-flag-%d", time.Now().UnixNano())
	roleName := schemaName

	admin := testkit.PostgresDB(t)
	_, err := admin.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, pdtenantdb.QuoteIdent(schemaName)))
	require.NoError(t, err)
	_, err = admin.ExecContext(ctx, fmt.Sprintf(`CREATE ROLE %s NOLOGIN`, pdtenantdb.QuoteIdent(roleName)))
	require.NoError(t, err)
	_, err = admin.ExecContext(ctx, fmt.Sprintf(
		`GRANT %s TO %s`, pdtenantdb.QuoteIdent(roleName), pdtenantdb.QuoteIdent(info.User),
	))
	require.NoError(t, err)

	placements := map[string]pdtenantdb.Placement{
		"t-flag": {
			TenantID:    "t-flag",
			ClusterName: "pg-01",
			Mode:        pdtenantdb.ModeSchema,
			DBName:      info.DBName,
			SchemaName:  schemaName,
		},
	}

	mOff := setupManager(t, &pdtenantdb.Config{SharedDB: info.DBName}, placements)
	err = mOff.WithTenantTx(ctx, "t-flag", &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		var currentUser string
		if err := tx.GetContext(ctx, &currentUser, `SELECT current_user`); err != nil {
			return err
		}
		require.NotEqual(t, roleName, currentUser, "TenantRoleEnabled=false must not switch role")
		return nil
	})
	require.NoError(t, err)

	mOn := setupManager(t, &pdtenantdb.Config{SharedDB: info.DBName, TenantRoleEnabled: true}, placements)
	err = mOn.WithTenantTx(ctx, "t-flag", &sql.TxOptions{}, func(tx *sqlx.Tx) error {
		var currentUser string
		if err := tx.GetContext(ctx, &currentUser, `SELECT current_user`); err != nil {
			return err
		}
		require.Equal(t, roleName, currentUser, "TenantRoleEnabled=true must SET LOCAL ROLE into the tenant role")
		return nil
	})
	require.NoError(t, err)
}

// PZEP-0017 / MT10: if a schema is dropped out of band (operator action),
// the bootstrapped cache must not keep believing it exists forever - the
// next call should fail once (schema genuinely gone), and the call after
// that should succeed (evicted + re-bootstrapped), without a process
// restart.
func TestManager_WithTenantTx_RecoversFromSchemaDroppedOutOfBand(t *testing.T) {
	info := testkit.PostgresInfo(t)
	schemaName := fmt.Sprintf("t_mt10_%d", time.Now().UnixNano())

	cfg := &pdtenantdb.Config{
		SharedDB: info.DBName,
		OnSchemaBootstrap: func(ctx context.Context, db *sqlx.DB, schema string) error {
			_, err := db.ExecContext(ctx, fmt.Sprintf(
				`CREATE TABLE IF NOT EXISTS %s.probe (id INT)`, pdtenantdb.QuoteIdent(schema),
			))
			return err
		},
	}
	placements := map[string]pdtenantdb.Placement{
		"tenant-mt10": {
			TenantID:    "tenant-mt10",
			ClusterName: "pg-01",
			Mode:        pdtenantdb.ModeSchema,
			DBName:      info.DBName,
			SchemaName:  schemaName,
		},
	}
	m := setupManager(t, cfg, placements)

	query := func(tx *sqlx.Tx) error {
		var count int
		return tx.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM probe")
	}

	// Provision the schema + probe table.
	require.NoError(t, m.WithTenantTx(context.Background(), "tenant-mt10", &sql.TxOptions{}, query))

	// Simulate an operator dropping the schema out of band.
	admin := testkit.PostgresDB(t)
	_, err := admin.ExecContext(context.Background(), fmt.Sprintf(
		`DROP SCHEMA %s CASCADE`, pdtenantdb.QuoteIdent(schemaName),
	))
	require.NoError(t, err)

	// The very next call still believes the schema is bootstrapped (cached)
	// and fails against the now-missing table/schema.
	err = m.WithTenantTx(context.Background(), "tenant-mt10", &sql.TxOptions{}, query)
	require.Error(t, err)

	// The eviction from that failure means the call after it re-bootstraps
	// and succeeds - no process restart needed.
	require.NoError(t, m.WithTenantTx(context.Background(), "tenant-mt10", &sql.TxOptions{}, query))
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

	noop := func(tx *sqlx.Tx) error { return nil }

	err := m.WithTenantTx(context.Background(), "a", &sql.TxOptions{}, noop)
	require.NoError(t, err)

	err = m.WithTenantTx(context.Background(), "b", &sql.TxOptions{}, noop)
	require.Error(t, err)
	require.Contains(t, err.Error(), "dedicated pool capacity")
}

// PZEP-0017 / MT16 (observability half): dedicated-pool capacity rejections
// and placement resolution failures must be logged.
func TestManager_LogsDedicatedPoolCapacityRejection(t *testing.T) {
	log := pdlogmocks.NewMockLogger(t)
	log.EXPECT().Info(mock.Anything, mock.Anything).Return().Maybe()
	log.EXPECT().Warn("pdtenantdb: dedicated pool capacity reached", mock.Anything).Return().Once()

	cfg := &pdtenantdb.Config{
		SharedDB:          testkit.PostgresInfo(t).DBName,
		MaxDedicatedPools: 1,
		Logger:            log,
	}
	testkit.EnsurePostgresDB(t, "bo_log_a")
	testkit.EnsurePostgresDB(t, "bo_log_b")
	placements := map[string]pdtenantdb.Placement{
		"log-a": {TenantID: "log-a", ClusterName: "pg-01", Mode: pdtenantdb.ModeDatabase, DBName: "bo_log_a"},
		"log-b": {TenantID: "log-b", ClusterName: "pg-01", Mode: pdtenantdb.ModeDatabase, DBName: "bo_log_b"},
	}
	m := setupManager(t, cfg, placements)
	noop := func(tx *sqlx.Tx) error { return nil }

	require.NoError(t, m.WithTenantTx(context.Background(), "log-a", &sql.TxOptions{}, noop))
	err := m.WithTenantTx(context.Background(), "log-b", &sql.TxOptions{}, noop)
	require.Error(t, err)
}

func TestManager_LogsPlacementResolutionFailure(t *testing.T) {
	log := pdlogmocks.NewMockLogger(t)
	log.EXPECT().Warn("pdtenantdb: placement resolution failed", mock.Anything).Return().Once()

	cfg := &pdtenantdb.Config{SharedDB: testkit.PostgresInfo(t).DBName, Logger: log}
	m := setupManager(t, cfg, map[string]pdtenantdb.Placement{})

	err := m.WithTenantTx(context.Background(), "missing-tenant", &sql.TxOptions{}, func(tx *sqlx.Tx) error { return nil })
	require.Error(t, err)
}

// PZEP-0017 / MT11: CloseIdleDedicated must not close a dedicated pool that
// a caller has an active checkout on, even if its lastUsed timestamp looks
// idle by the time the janitor runs. Proven black-box: track how many times
// the pool is actually opened (SQLXOpen) rather than reaching into
// unexported state - if CloseIdleDedicated wrongly closed the pool while it
// was checked out, the very next WithTenantTx call for the same tenant
// would have to reopen it, which would show up as a second SQLXOpen call
// before the caller in flight even finished.
func TestManager_CloseIdleDedicated_DoesNotCloseCheckedOutPool(t *testing.T) {
	cfg := &pdtenantdb.Config{
		SharedDB:          testkit.PostgresInfo(t).DBName,
		MaxDedicatedPools: 10,
		DedicatedIdleTTL:  time.Nanosecond, // eligible to close almost immediately
	}
	testkit.EnsurePostgresDB(t, "bo_checkout")
	placements := map[string]pdtenantdb.Placement{
		"checkout-tenant": {
			TenantID: "checkout-tenant", ClusterName: "pg-01",
			Mode: pdtenantdb.ModeDatabase, DBName: "bo_checkout",
		},
	}

	original := pdtenantdb.SQLXOpen
	var opens int32
	pdtenantdb.SQLXOpen = func(driverName, dataSourceName string) (*sqlx.DB, error) {
		atomic.AddInt32(&opens, 1)
		return original(driverName, dataSourceName)
	}
	t.Cleanup(func() { pdtenantdb.SQLXOpen = original })

	m := setupManager(t, cfg, placements)
	noop := func(tx *sqlx.Tx) error { return nil }

	inTx := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- m.WithTenantTx(context.Background(), "checkout-tenant", &sql.TxOptions{}, func(tx *sqlx.Tx) error {
			close(inTx)
			<-release
			return nil
		})
	}()

	<-inTx
	require.Equal(t, int32(1), atomic.LoadInt32(&opens))

	// The janitor runs while the transaction above is still in flight - the
	// pool's lastUsed is already "idle" by DedicatedIdleTTL's tiny window,
	// but the checkout must keep it alive.
	m.CloseIdleDedicated(time.Now().Add(time.Hour))

	close(release)
	require.NoError(t, <-done)

	// A follow-up call must reuse the same pool - proving the janitor did
	// not close it while checked out.
	err := m.WithTenantTx(context.Background(), "checkout-tenant", &sql.TxOptions{}, noop)
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&opens), "pool must not have been reopened")

	// Now that nothing holds a checkout, the janitor can actually close it.
	m.CloseIdleDedicated(time.Now().Add(time.Hour))
	err = m.WithTenantTx(context.Background(), "checkout-tenant", &sql.TxOptions{}, noop)
	require.NoError(t, err)
	require.Equal(t, int32(2), atomic.LoadInt32(&opens), "pool must have been reopened after the idle close")
}
