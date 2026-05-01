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
	CloseAll() error
}

type managerImpl struct {
	cfg      *Config
	resolver PlacementResolver
	registry ClusterRegistry

	sf singleflight.Group

	mu    sync.Mutex
	pools map[ConnKey]*sqlx.DB

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
		cfg:      cfg,
		resolver: resolver,
		registry: registry,
		pools:    make(map[ConnKey]*sqlx.DB),
		lastUsed: make(map[ConnKey]time.Time),
		creating: make(map[ConnKey]struct{}),
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

func (m *managerImpl) ensureSchema(ctx context.Context, db *sqlx.DB, key ConnKey, schemaName string) error {
	bsKey := key.ClusterName + "|" + key.DBName + "|" + schemaName
	if _, loaded := m.bootstrapped.LoadOrStore(bsKey, struct{}{}); loaded {
		return nil
	}
	_, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, pgQuoteIdent(schemaName)))
	if err != nil {
		// Remove so the next call can retry.
		m.bootstrapped.Delete(bsKey)
	}
	return err
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
			pgQuoteIdent(pl.SchemaName),
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

		dsn := buildDSN(clusterCfg, key.DBName)
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
	return nil
}

func buildDSN(c ClusterConfig, dbName string) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
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

func pgQuoteIdent(s string) string {
	// Quote identifier to avoid injection in SET search_path.
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
