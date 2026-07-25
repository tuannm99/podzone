package pdtenantdb

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/sync/singleflight"
)

var SQLXOpen = sqlx.Open

type Manager interface {
	DBForTenant(ctx context.Context, tenantID string) (*sqlx.DB, Placement, error)
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

	return &managerImpl{
		cfg:        cfg,
		resolver:   resolver,
		registry:   registry,
		pools:      make(map[ConnKey]*sqlx.DB),
		adminPools: make(map[ConnKey]*sqlx.DB),
		lastUsed:   make(map[ConnKey]time.Time),
		creating:   make(map[ConnKey]struct{}),
	}
}

func (m *managerImpl) DBForTenant(ctx context.Context, tenantID string) (*sqlx.DB, Placement, error) {
	pl, err := m.resolver.Resolve(ctx, tenantID)
	if err != nil {
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
	db, pl, err := m.DBForTenant(ctx, tenantID)
	if err != nil {
		return err
	}

	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// IMPORTANT: use SET LOCAL for PgBouncer transaction pooling.
	if pl.Mode == ModeSchema && pl.SchemaName != "" {
		_, err = tx.ExecContext(ctx, fmt.Sprintf(
			`SET LOCAL search_path TO %s, public`,
			QuoteIdent(pl.SchemaName),
		))
		if err != nil {
			return err
		}
	}

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
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

// Optional janitor: call this with a ticker to close idle dedicated pools.
func (m *managerImpl) CloseIdleDedicated(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k, t := range m.lastUsed {
		if k.DBName == m.cfg.SharedDB {
			continue
		}
		if now.Sub(t) >= m.cfg.DedicatedIdleTTL {
			if db := m.pools[k]; db != nil {
				_ = db.Close()
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
