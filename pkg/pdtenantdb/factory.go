package pdtenantdb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	SSLMode  string
	BaseDB   string
}

type managerImpl struct {
	mu     sync.RWMutex
	dbs    map[string]*sqlx.DB
	config *Config
}

func NewManager(cfg *Config) *managerImpl {
	return &managerImpl{
		dbs:    make(map[string]*sqlx.DB),
		config: cfg,
	}
}

func (m *managerImpl) GetDB(ctx context.Context, tenant string) (*sqlx.DB, error) {
	m.mu.RLock()
	if db, ok := m.dbs[tenant]; ok {
		m.mu.RUnlock()
		return db, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if db, ok := m.dbs[tenant]; ok {
		return db, nil
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s_%s?sslmode=%s",
		m.config.User, m.config.Password,
		m.config.Host, m.config.Port,
		m.config.BaseDB, tenant, m.config.SSLMode,
	)

	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)

	m.dbs[tenant] = db
	return db, nil
}

func (m *managerImpl) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, db := range m.dbs {
		if err := db.Close(); err != nil {
			fmt.Printf("error closing db for tenant %s: %v\n", name, err)
		}
		delete(m.dbs, name)
	}
	return nil
}
