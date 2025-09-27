// TODO: FINISHING ME
package tenantdb

// import (
// 	"context"
// 	"fmt"
// 	"sync"
// 	"time"
//
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// 	"gorm.io/gorm/logger"
//
// 	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
// 	"github.com/tuannm99/podzone/pkg/toolkit/cache"
// )

// type TenantType string
//
// const (
// 	TenantTypeGroup     TenantType = "group"     // separate schema per tenant
// 	TenantTypeDedicated TenantType = "dedicated" // per-tenant database isolation
// )
//
// const (
// 	defaultGroupSize = 999
// 	// MaxPoolSize is the maximum number of connections in the pool for group-based isolation
// 	MaxPoolSize = 200
// 	// MaxDedicatedConnections is the maximum number of connections per application for dedicated databases
// 	MaxDedicatedConnections = 50
// )
//
// // TenantDBManager manages tenant-specific database connections
// type TenantDBManager struct {
// 	config   *Config
// 	pool     sync.Map
// 	logger   pdlog.Logger
// 	lruCache *cache.Lru
// }
//
// type Config struct {
// 	Host     string
// 	Port     int
// 	User     string
// 	Password string
// 	DBName   string
// 	SSLMode  string
// 	// TenantType determines the isolation strategy
// 	TenantType TenantType
// 	// GroupSize is the number of tenants per group (only used for group-based isolation)
// 	GroupSize int
// 	// MaxPoolSize overrides the default MaxPoolSize for group-based isolation
// 	MaxPoolSize int
// 	// MaxDedicatedConnections overrides the default MaxDedicatedConnections for dedicated databases
// 	MaxDedicatedConnections int
// }
//
// // NewTenantDBManager creates a new tenant database manager
// func NewTenantDBManager(config *Config, logger pdlog.Logger) *TenantDBManager {
// 	if config.GroupSize <= 0 {
// 		config.GroupSize = defaultGroupSize
// 	}
// 	if config.MaxPoolSize <= 0 {
// 		config.MaxPoolSize = MaxPoolSize
// 	}
// 	if config.MaxDedicatedConnections <= 0 {
// 		config.MaxDedicatedConnections = MaxDedicatedConnections
// 	}
//
// 	return &TenantDBManager{
// 		config:   config,
// 		logger:   logger,
// 		lruCache: cache.NewCache(config.MaxPoolSize),
// 	}
// }
//
// // getGroupID returns the group ID for a tenant
// func (m *TenantDBManager) getGroupID(tenantID string) string {
// 	// Simple hash function to determine group
// 	hash := 0
// 	for _, c := range tenantID {
// 		hash = 31*hash + int(c)
// 	}
// 	groupID := (hash % m.config.GroupSize) + 1
// 	return fmt.Sprintf("group_%03d", groupID)
// }
//
// // getDatabaseName returns the database name based on tenant type
// func (m *TenantDBManager) getDatabaseName(tenantID string) string {
// 	switch m.config.TenantType {
// 	case TenantTypeDedicated:
// 		return fmt.Sprintf("tenant_%s", tenantID)
// 	case TenantTypeGroup:
// 		return m.getGroupID(tenantID)
// 	default:
// 		return m.config.DBName
// 	}
// }
//
// // getSchemaName returns the schema name based on tenant type
// func (m *TenantDBManager) getSchemaName(tenantID string) string {
// 	switch m.config.TenantType {
// 	case TenantTypeDedicated:
// 		return "default"
// 	case TenantTypeGroup:
// 		return fmt.Sprintf("tenant_%s", tenantID)
// 	default:
// 		return "public"
// 	}
// }
//
// // CreateTenantSchema creates a new schema for a tenant
// func (m *TenantDBManager) CreateTenantSchema(ctx context.Context, tenantID string) error {
// 	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
// 		m.config.Host,
// 		m.config.Port,
// 		m.config.User,
// 		m.config.Password,
// 		m.getDatabaseName(tenantID),
// 		m.config.SSLMode,
// 	)
//
// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to database: %w", err)
// 	}
//
// 	// Create schema for tenant
// 	schemaName := m.getSchemaName(tenantID)
// 	err = db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error
// 	if err != nil {
// 		return fmt.Errorf("failed to create schema: %w", err)
// 	}
//
// 	m.logger.Info("Created new schema for tenant").
// 		With("tenant_id", tenantID).
// 		With("schema", schemaName).
// 		With("database", m.getDatabaseName(tenantID)).
// 		Send()
//
// 	return nil
// }
//
// // GetDB gets a database instance for the current tenant
// func (m *TenantDBManager) GetDB(ctx context.Context) (*gorm.DB, error) {
// 	// tenantID, ok := pdcontext.GetTenantID(ctx)
// 	// if !ok {
// 	// 	return nil, pdcontext.ErrUnauthorized
// 	// }
//
// 	tenantID := "tenant_1"
//
// 	// For dedicated databases, check connection limit
// 	if m.config.TenantType == TenantTypeDedicated {
// 		count := 0
// 		m.pool.Range(func(_, _ interface{}) bool {
// 			count++
// 			return true
// 		})
// 		if count >= m.config.MaxDedicatedConnections {
// 			return nil, fmt.Errorf("maximum dedicated connections (%d) reached", m.config.MaxDedicatedConnections)
// 		}
// 	}
//
// 	// Try to get from pool first
// 	if db, ok := m.pool.Load(tenantID); ok {
// 		if m.config.TenantType == TenantTypeGroup {
// 			m.lruCache.Update(tenantID)
// 		}
// 		return db.(*gorm.DB), nil
// 	}
//
// 	// For group-based isolation, check LRU cache
// 	if m.config.TenantType == TenantTypeGroup {
// 		// If cache is full, remove least recently used connection
// 		if m.lruCache.IsFull() {
// 			oldTenantID := m.lruCache.RemoveOldest()
// 			if oldTenantID != "" {
// 				if oldDB, ok := m.pool.Load(oldTenantID); ok {
// 					sqlDB, err := oldDB.(*gorm.DB).DB()
// 					if err == nil {
// 						sqlDB.Close()
// 					}
// 					m.pool.Delete(oldTenantID)
// 					m.logger.Info("Removed old connection from pool").With("tenant_id", oldTenantID).Send()
// 				}
// 			}
// 		}
// 	}
//
// 	// Create new database connection
// 	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
// 		m.config.Host,
// 		m.config.Port,
// 		m.config.User,
// 		m.config.Password,
// 		m.getDatabaseName(tenantID),
// 		m.config.SSLMode,
// 	)
//
// 	// Configure GORM logger
// 	gormLogger := logger.New(
// 		&GormLogger{Logger: m.logger},
// 		logger.Config{
// 			SlowThreshold:             time.Second,
// 			LogLevel:                  logger.Info,
// 			IgnoreRecordNotFoundError: true,
// 			Colorful:                  false,
// 		},
// 	)
//
// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
// 		Logger: gormLogger,
// 	})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to database: %w", err)
// 	}
//
// 	// Set schema for this connection
// 	schemaName := m.getSchemaName(tenantID)
// 	err = db.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)).Error
// 	if err != nil {
// 		// If schema doesn't exist, try to create it
// 		if err := m.CreateTenantSchema(ctx, tenantID); err != nil {
// 			return nil, fmt.Errorf("failed to create tenant schema: %w", err)
// 		}
// 		// Try setting schema again
// 		err = db.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)).Error
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to set schema: %w", err)
// 		}
// 	}
//
// 	// Store in pool
// 	m.pool.Store(tenantID, db)
// 	if m.config.TenantType == TenantTypeGroup {
// 		m.lruCache.Add(tenantID)
// 	}
//
// 	m.logger.Info("Created new database connection for tenant").
// 		With("tenant_id", tenantID).
// 		With("schema", schemaName).
// 		With("database", m.getDatabaseName(tenantID)).
// 		Send()
//
// 	return db, nil
// }
//
// // Close closes all database connections
// func (m *TenantDBManager) Close() {
// 	m.pool.Range(func(key, value interface{}) bool {
// 		db := value.(*gorm.DB)
// 		sqlDB, err := db.DB()
// 		if err != nil {
// 			m.logger.Error("Failed to get underlying SQL DB").With("tenant_id", key.(string)).Err(err).Send()
// 			return true
// 		}
// 		if err := sqlDB.Close(); err != nil {
// 			m.logger.Error("Failed to close database connection").With("tenant_id", key.(string)).Err(err).Send()
// 		}
// 		m.logger.Info("Closed database connection for tenant").With("tenant_id", key.(string)).Send()
// 		return true
// 	})
// 	m.pool = sync.Map{}
// 	m.lruCache.Clear()
// }
//
// // GormLogger implements gorm.Logger interface
// type GormLogger struct {
// 	pdlog.Logger
// }
//
// func (l *GormLogger) Printf(format string, args ...interface{}) {
// 	l.Info(fmt.Sprintf(format, args...)).Send()
// }
