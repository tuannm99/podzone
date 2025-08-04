package dbpool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/pkg/toolkit/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBType string

const (
	TenantDB DBType = "tenantdb"
	GroupDB  DBType = "groupdb"
)

type DBConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	DBName    string `json:"dbname"`
	SSLMode   string `json:"sslmode"`
	Type      DBType `json:"type"`
	Namespace string `json:"namespace"`
}

type Pool struct {
	mu        sync.RWMutex
	conns     map[string]*gorm.DB
	configs   map[string]DBConfig
	appConfig *config.AppConfig
	maxConns  int
	ns        string // Kubernetes namespace
}

func NewPool(appConfig config.AppConfig, apiURL string, maxConns int) *Pool {
	ns := toolkit.GetEnv("KUBERNETES_NAMESPACE", "default")

	return &Pool{
		conns:     make(map[string]*gorm.DB),
		configs:   make(map[string]DBConfig),
		appConfig: &appConfig,
		maxConns:  maxConns,
		ns:        ns,
	}
}

func (p *Pool) FetchConfigs(ctx context.Context) error {
	configs, err := p.appConfig.KVStores.Get("/connection/db")
	if err != nil {
		return fmt.Errorf("failed to fetch database configurations: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.configs = configs

	for name, db := range p.conns {
		if _, exists := configs[name]; !exists {
			sqlDB, err := db.DB()
			if err == nil {
				sqlDB.Close()
			}
			delete(p.conns, name)
		}
	}

	return nil
}

func (p *Pool) findAvailableDB(dbType DBType) (*DBConfig, error) {
	var availableDB *DBConfig

	for _, config := range p.configs {
		if config.Type == dbType && config.Namespace == p.ns {
			availableDB = &config
		}
	}

	if availableDB == nil {
		return nil, fmt.Errorf("no available database found for type %s in namespace %s", dbType, p.ns)
	}

	return availableDB, nil
}

func (p *Pool) GetDB(dbType DBType) (*gorm.DB, error) {
	config, err := p.findAvailableDB(dbType)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s-%s-%s", p.ns, config.DBName, dbType)

	p.mu.RLock()
	db, exists := p.conns[key]
	p.mu.RUnlock()

	if exists {
		return db, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if db, exists = p.conns[key]; exists {
		return db, nil
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying *sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(p.maxConns)
	sqlDB.SetMaxIdleConns(p.maxConns / 2)
	sqlDB.SetConnMaxLifetime(time.Hour)

	p.conns[key] = db
	return db, nil
}

func (p *Pool) GetTenantDB() (*gorm.DB, error) {
	return p.GetDB(TenantDB)
}

func (p *Pool) GetGroupDB() (*gorm.DB, error) {
	return p.GetDB(GroupDB)
}

func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var lastErr error
	for name, db := range p.conns {
		sqlDB, err := db.DB()
		if err != nil {
			lastErr = fmt.Errorf("failed to get underlying *sql.DB for %s: %w", name, err)
			continue
		}
		if err := sqlDB.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close connection %s: %w", name, err)
		}
		delete(p.conns, name)
	}

	return lastErr
}
