package pdtenantdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"golang.org/x/sync/singleflight"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

var SQLXOpen = sqlx.Open

type Manager interface {
	// Note: no method here hands out a raw *sqlx.DB. dbForTenant (unexported)
	// used to be public as DBForTenant, but returned a pool with no
	// search_path/role applied - see MT4, docs/07-problems/
	// multitenancy-2026-07-23.md. WithTenantTx is the only supported way to
	// get a tenant-scoped connection.
	WithTenantTx(ctx context.Context, tenantID string, opts *sql.TxOptions, fn func(tx *sqlx.Tx) error) error
	// CloseIdleDedicated closes dedicated pools that have been idle for longer than DedicatedIdleTTL.
	// Call this periodically (e.g. every DedicatedIdleTTL/2) to prevent unbounded pool growth.
	CloseIdleDedicated(now time.Time)
	CloseAll() error
}

type managerImpl struct {
	cfg      *Config
	resolver PlacementResolver
	registry ClusterRegistry

	sf singleflight.Group

	mu    sync.Mutex
	pools map[ConnKey]*sqlx.DB

	// adminPools holds the elevated-credential pool for a (cluster,db) key,
	// used only for schema bootstrap / migrations, when ClusterConfig.AppUser
	// is set. Empty/unused when AppUser is not configured - see PZEP-0009.
	adminPools map[ConnKey]*sqlx.DB

	// Track last used for dedicated DB pools (to close idle pools)
	lastUsed map[ConnKey]time.Time
	creating map[ConnKey]struct{}

	// checkouts counts in-flight callers between getOrCreateDB returning a
	// pool and WithTenantTx's BeginTxx call on it actually completing - the
	// narrow window CloseIdleDedicated could otherwise close a pool out
	// from under (MT11, docs/07-problems/multitenancy-2026-07-23.md).
	checkouts map[ConnKey]int

	// bootstrapped tracks schemas that have already had CREATE SCHEMA IF NOT EXISTS run,
	// keyed by "clusterName|dbName|schemaName".
	bootstrapped sync.Map
}

func NewManager(cfg *Config, resolver PlacementResolver, registry ClusterRegistry) Manager {
	if cfg.SharedDB == "" {
		cfg.SharedDB = "backoffice"
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 5 * time.Minute
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 10
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 2
	}
	if cfg.MaxDedicatedPools == 0 {
		cfg.MaxDedicatedPools = 200
	}
	if cfg.DedicatedIdleTTL == 0 {
		cfg.DedicatedIdleTTL = 30 * time.Minute
	}
	if cfg.Logger == nil {
		cfg.Logger = pdlog.NopLogger{}
	}

	return &managerImpl{
		cfg:        cfg,
		resolver:   resolver,
		registry:   registry,
		pools:      make(map[ConnKey]*sqlx.DB),
		adminPools: make(map[ConnKey]*sqlx.DB),
		lastUsed:   make(map[ConnKey]time.Time),
		creating:   make(map[ConnKey]struct{}),
		checkouts:  make(map[ConnKey]int),
	}
}

// dbForTenant resolves a tenant's connection pool. Unexported on purpose -
// see the Manager interface's comment (MT4). Only WithTenantTx may call
// this; every other caller must go through WithTenantTx, which applies
// search_path (and, when enabled, the per-tenant role - PZEP-0015) before
// any query runs.
func (m *managerImpl) dbForTenant(ctx context.Context, tenantID string) (*sqlx.DB, Placement, error) {
	pl, err := m.resolver.Resolve(ctx, tenantID)
	if err != nil {
		m.cfg.Logger.Warn("pdtenantdb: placement resolution failed", "tenant_id", tenantID, "error", err)
		return nil, Placement{}, err
	}
	if pl.ClusterName == "" {
		return nil, Placement{}, fmt.Errorf("pdtenantdb: missing cluster_name for tenant %s", tenantID)
	}
	if pl.DBName == "" {
		return nil, Placement{}, fmt.Errorf("pdtenantdb: missing db_name for tenant %s", tenantID)
	}
	key := ConnKey{ClusterName: pl.ClusterName, DBName: pl.DBName}

	dedicated := pl.Mode == ModeDatabase && key.DBName != m.cfg.SharedDB
	db, err := m.getOrCreateDB(ctx, key, dedicated)
	if err != nil {
		return nil, Placement{}, err
	}
	// Refcount the checkout from the moment the pool reference exists until
	// the caller is done with it (WithTenantTx releases it once the whole
	// transaction completes) - see MT11. release=true is the "something
	// failed before we could hand off" path; the two success paths below
	// both flip it to false.
	m.checkout(key)
	release := true
	defer func() {
		if release {
			m.release(key)
		}
	}()

	if pl.Mode == ModeSchema {
		if pl.SchemaName == "" {
			return nil, Placement{}, fmt.Errorf(
				"pdtenantdb: missing schema_name for tenant %s in schema mode",
				tenantID,
			)
		}
		if err := m.ensureSchema(ctx, db, key, pl.SchemaName); err != nil {
			return nil, Placement{}, err
		}
	}

	if dedicated {
		m.markUsed(key)
	}
	release = false // handing off to the caller, which now owns releasing it
	return db, pl, nil
}

// ensureSchema creates schemaName (if needed) and runs OnSchemaBootstrap,
// then marks the scope done so later calls skip both.
//
// Concurrent callers for the same schema MUST wait for the winner to
// actually finish, not just observe a flag and return - a naive
// LoadOrStore-then-return-if-already-set guard marks "done" before CREATE
// SCHEMA commits, so a loser can proceed believing the schema exists when
// it doesn't yet. Its own current_schema() lookup then falls back to
// "public" (Postgres skips non-existent search_path entries), computing
// the WRONG advisory-lock key for its later migration statements - but
// those statements re-resolve the unqualified target schema dynamically
// per-statement, not once at transaction start. If the real schema becomes
// visible mid-transaction (the actual winner committed concurrently), the
// "public"-keyed transaction's later CREATE TABLE silently retargets into
// the real tenant schema too, colliding with whoever holds the correct
// lock there ("pg_type_typname_nsp_index" duplicate key). singleflight
// fixes this by blocking every concurrent caller until the winner's CREATE
// SCHEMA + OnSchemaBootstrap has fully completed - see
// docs/07-problems/multitenancy-2026-07-23.md MT10 (this was the same
// bootstrapped map root-caused a step further).
func (m *managerImpl) ensureSchema(ctx context.Context, db *sqlx.DB, key ConnKey, schemaName string) error {
	bsKey := key.ClusterName + "|" + key.DBName + "|" + schemaName
	if _, loaded := m.bootstrapped.Load(bsKey); loaded {
		return nil
	}

	_, err, _ := m.sf.Do("schema:"+bsKey, func() (any, error) {
		if _, loaded := m.bootstrapped.Load(bsKey); loaded {
			return nil, nil
		}
		// CREATE SCHEMA and migrations always need the admin credential,
		// even when db (the caller's query pool) is running as the
		// least-privilege app role. adminDB falls back to db itself when
		// no app role is configured - see PZEP-0009.
		adminDB, err := m.adminDBFor(ctx, key, db)
		if err != nil {
			return nil, err
		}
		createSchemaSQL := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, QuoteIdent(schemaName))
		if _, err := adminDB.ExecContext(ctx, createSchemaSQL); err != nil {
			return nil, err
		}
		if m.cfg.OnSchemaBootstrap != nil {
			if err := m.cfg.OnSchemaBootstrap(ctx, adminDB, schemaName); err != nil {
				return nil, err
			}
		}
		m.bootstrapped.Store(bsKey, struct{}{})
		return nil, nil
	})
	return err
}

// adminDBFor returns the elevated-credential pool for key, opening it on
// first use. Returns fallback unchanged when the cluster has no AppUser
// configured, so callers with AppUser unset never open a second connection
// pool - see PZEP-0009.
func (m *managerImpl) adminDBFor(ctx context.Context, key ConnKey, fallback *sqlx.DB) (*sqlx.DB, error) {
	clusterCfg, err := m.registry.GetCluster(ctx, key.ClusterName)
	if err != nil {
		return nil, err
	}
	if clusterCfg.AppUser == "" {
		return fallback, nil
	}

	m.mu.Lock()
	if db := m.adminPools[key]; db != nil {
		m.mu.Unlock()
		return db, nil
	}
	m.mu.Unlock()

	dsn := buildDSN(clusterCfg, key.DBName, false)
	db, err := SQLXOpen("postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(m.cfg.ConnMaxLifetime)
	db.SetMaxOpenConns(m.cfg.MaxOpenConns)
	db.SetMaxIdleConns(m.cfg.MaxIdleConns)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	m.mu.Lock()
	if existing := m.adminPools[key]; existing != nil {
		m.mu.Unlock()
		_ = db.Close()
		return existing, nil
	}
	m.adminPools[key] = db
	m.mu.Unlock()
	return db, nil
}

func (m *managerImpl) WithTenantTx(
	ctx context.Context,
	tenantID string,
	opts *sql.TxOptions,
	fn func(tx *sqlx.Tx) error,
) error {
	db, pl, err := m.dbForTenant(ctx, tenantID)
	if err != nil {
		return err
	}
	defer m.release(ConnKey{ClusterName: pl.ClusterName, DBName: pl.DBName})

	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// IMPORTANT: use SET LOCAL for PgBouncer transaction pooling. No
	// "public" fallback on the search_path - see
	// docs/07-problems/multitenancy-2026-07-23.md MT6. A table missing from
	// the tenant schema must raise "relation does not exist", not silently
	// resolve to a same-named table in public.
	if pl.Mode == ModeSchema && pl.SchemaName != "" {
		_, err = tx.ExecContext(ctx, fmt.Sprintf(
			`SET LOCAL search_path TO %s`,
			QuoteIdent(pl.SchemaName),
		))
		if err != nil {
			return err
		}

		// Phase 2 of MT1 (docs/09-pzep/
		// PZEP-0015-backoffice-per-tenant-db-role-phase-2.md): scope this
		// transaction's Postgres privileges to exactly this tenant's schema,
		// independent of search_path. Off by default (TenantRoleEnabled) -
		// requires migration 0018_create_backoffice_tenant_role.sql to have
		// already provisioned the role for this schema.
		if m.cfg.TenantRoleEnabled {
			_, err = tx.ExecContext(ctx, fmt.Sprintf(
				`SET LOCAL ROLE %s`,
				QuoteIdent(tenantRoleName(pl.SchemaName)),
			))
			if err != nil {
				return err
			}
		}
	}

	if err := fn(tx); err != nil {
		// MT10, docs/07-problems/multitenancy-2026-07-23.md: bootstrapped
		// (and, via OnScopeInvalidated, migrations.appliedScopes) cache a
		// permanent "this schema exists" bit with no expiry. If an operator
		// drops and later restores a schema out of band, every call
		// believes it's still there and fails until a process restart.
		// Confirm the error is genuinely "undefined_table" and the schema
		// itself is actually gone (not just an unrelated app bug hitting a
		// missing table) before evicting - a plain "42P01 seen" is too
		// broad a trigger on its own.
		if pl.Mode == ModeSchema && pl.SchemaName != "" && m.schemaLikelyMissing(ctx, db, pl.SchemaName, err) {
			key := ConnKey{ClusterName: pl.ClusterName, DBName: pl.DBName}
			m.bootstrapped.Delete(key.ClusterName + "|" + key.DBName + "|" + pl.SchemaName)
			if m.cfg.OnScopeInvalidated != nil {
				m.cfg.OnScopeInvalidated(pl.DBName, pl.SchemaName)
			}
			m.cfg.Logger.Warn("pdtenantdb: schema missing, evicted bootstrap cache for re-provisioning",
				"cluster", pl.ClusterName, "db", pl.DBName, "schema", pl.SchemaName)
		}
		return err
	}
	return tx.Commit()
}

// schemaLikelyMissing confirms err is a genuine "the schema itself is gone"
// signal, not an unrelated missing-table bug: SQLSTATE 42P01
// (undefined_table, what an unqualified query resolves to when search_path
// points at a nonexistent schema - confirmed empirically against real
// Postgres 15, since SET LOCAL search_path itself does not validate the
// schema exists), followed by an explicit to_regnamespace check so a
// same-schema, different-missing-table app bug does not trigger a
// schema-wide re-bootstrap. Runs on a fresh query against db (the pool),
// not tx - tx is already aborted by the error being classified.
func (m *managerImpl) schemaLikelyMissing(ctx context.Context, db *sqlx.DB, schemaName string, err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) || pqErr.Code != "42P01" {
		return false
	}
	var exists bool
	if checkErr := db.GetContext(ctx, &exists, `SELECT to_regnamespace($1) IS NOT NULL`, schemaName); checkErr != nil {
		return false
	}
	return !exists
}

func (m *managerImpl) getOrCreateDB(ctx context.Context, key ConnKey, dedicated bool) (*sqlx.DB, error) {
	m.mu.Lock()
	if db := m.pools[key]; db != nil {
		m.mu.Unlock()
		return db, nil
	}
	m.mu.Unlock()

	v, err, _ := m.sf.Do(key.ClusterName+"|"+key.DBName, func() (any, error) {
		m.mu.Lock()
		if db := m.pools[key]; db != nil {
			m.mu.Unlock()
			return db, nil
		}
		if dedicated {
			if err := m.reserveDedicatedLocked(key); err != nil {
				m.mu.Unlock()
				return nil, err
			}
		}
		m.mu.Unlock()
		if dedicated {
			defer m.releaseDedicatedReservation(key)
		}

		clusterCfg, err := m.registry.GetCluster(ctx, key.ClusterName)
		if err != nil {
			return nil, err
		}

		// true: prefer AppUser/AppPassword when configured. This is the
		// query pool WithTenantTx hands to repository code - falls back to
		// the admin credential when no app role is configured, identical to
		// today's DSN. See PZEP-0009.
		dsn := buildDSN(clusterCfg, key.DBName, true)
		db, err := SQLXOpen("postgres", dsn)
		if err != nil {
			return nil, err
		}

		db.SetConnMaxLifetime(m.cfg.ConnMaxLifetime)
		db.SetMaxOpenConns(m.cfg.MaxOpenConns)
		db.SetMaxIdleConns(m.cfg.MaxIdleConns)

		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			return nil, err
		}
		m.cfg.Logger.Info("pdtenantdb: pool opened",
			"cluster", key.ClusterName, "db", key.DBName, "dedicated", dedicated)

		// Dedicated (database-per-tenant) pools have no per-schema bootstrap
		// hook - this is the once-per-pool equivalent for database mode,
		// matching ensureSchema's schema-mode hook. Always runs against the
		// admin credential, even though db above may be the app-role pool.
		if dedicated && m.cfg.OnSchemaBootstrap != nil {
			adminDB, err := m.adminDBFor(ctx, key, db)
			if err != nil {
				_ = db.Close()
				return nil, err
			}
			if err := m.cfg.OnSchemaBootstrap(ctx, adminDB, ""); err != nil {
				_ = db.Close()
				return nil, err
			}
		}

		m.mu.Lock()
		m.pools[key] = db
		m.mu.Unlock()
		return db, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*sqlx.DB), nil
}

func (m *managerImpl) reserveDedicatedLocked(key ConnKey) error {
	if _, exists := m.creating[key]; exists {
		return nil
	}
	count := 0
	for k := range m.pools {
		if k.DBName != m.cfg.SharedDB {
			count++
		}
	}
	count += len(m.creating)
	if count >= m.cfg.MaxDedicatedPools {
		m.cfg.Logger.Warn("pdtenantdb: dedicated pool capacity reached",
			"cluster", key.ClusterName, "db", key.DBName, "count", count, "limit", m.cfg.MaxDedicatedPools)
		return fmt.Errorf("%w (%d)", ErrDedicatedPoolCapacity, m.cfg.MaxDedicatedPools)
	}
	m.creating[key] = struct{}{}
	return nil
}

func (m *managerImpl) releaseDedicatedReservation(key ConnKey) {
	m.mu.Lock()
	delete(m.creating, key)
	m.mu.Unlock()
}

func (m *managerImpl) markUsed(key ConnKey) {
	m.mu.Lock()
	m.lastUsed[key] = time.Now()
	m.mu.Unlock()
}

// checkout/release track in-flight users of a pool so CloseIdleDedicated
// (MT11) cannot close one out from under an active caller. Safe to call for
// any key, dedicated or shared - CloseIdleDedicated only ever consults this
// for dedicated (non-SharedDB) keys, but tracking is unconditional to keep
// the bookkeeping simple.
func (m *managerImpl) checkout(key ConnKey) {
	m.mu.Lock()
	m.checkouts[key]++
	m.mu.Unlock()
}

func (m *managerImpl) release(key ConnKey) {
	m.mu.Lock()
	if m.checkouts[key] > 0 {
		m.checkouts[key]--
		if m.checkouts[key] == 0 {
			delete(m.checkouts, key)
		}
	}
	m.mu.Unlock()
}

// Optional janitor: call this with a ticker to close idle dedicated pools.
func (m *managerImpl) CloseIdleDedicated(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, t := range m.lastUsed {
		if k.DBName == m.cfg.SharedDB {
			continue
		}
		if m.checkouts[k] > 0 {
			continue // MT11: an active caller has this pool checked out
		}
		if now.Sub(t) >= m.cfg.DedicatedIdleTTL {
			if db := m.pools[k]; db != nil {
				_ = db.Close()
				m.cfg.Logger.Info("pdtenantdb: idle dedicated pool closed", "cluster", k.ClusterName, "db", k.DBName)
			}
			delete(m.pools, k)
			delete(m.lastUsed, k)
		}
	}
}

func (m *managerImpl) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, db := range m.pools {
		_ = db.Close()
		delete(m.pools, k)
		delete(m.lastUsed, k)
	}
	for k, db := range m.adminPools {
		_ = db.Close()
		delete(m.adminPools, k)
	}
	return nil
}

// buildDSN builds the connection string for c/dbName. When preferAppRole is
// true and c.AppUser is set, the least-privilege app credential is used
// instead of the admin one - see PZEP-0009. Bootstrap/migration connections
// always pass preferAppRole=false.
func buildDSN(c ClusterConfig, dbName string, preferAppRole bool) string {
	user, password := c.User, c.Password
	if preferAppRole && c.AppUser != "" {
		user, password = c.AppUser, c.AppPassword
	}
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   "/" + dbName,
	}
	q := url.Values{}
	if c.SSLMode != "" {
		q.Set("sslmode", c.SSLMode)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func QuoteIdent(s string) string {
	// Quote identifier to avoid injection in SET search_path.
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// tenantRoleName returns the per-tenant Postgres role name for schemaName.
// It is intentionally the schema name itself, not a further-prefixed
// variant - schema names already carry a tenant prefix (e.g. t_tenant_dev),
// and a role and a schema occupy different Postgres namespaces, so an
// identical name is not a collision. Kept as a named function (not inlined)
// so migrations/sql/0018_create_backoffice_tenant_role.sql and this call
// site cannot drift on the naming convention.
func tenantRoleName(schemaName string) string {
	return schemaName
}
