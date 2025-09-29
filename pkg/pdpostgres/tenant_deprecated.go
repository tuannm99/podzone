package pdpostgres

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	// "github.com/tuannm99/podzone/pkg/pdcontext"
)

// TenantDBManager manages tenant-specific database connections
type TenantDBManager struct {
	config *ConfigT
	pool   sync.Map
	logger pdlog.Logger
}

// Config holds PostgreSQL connection configuration
type ConfigT struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewTenantDBManager creates a new tenant database manager
func NewTenantDBManager(config *ConfigT, logger pdlog.Logger) *TenantDBManager {
	return &TenantDBManager{
		config: config,
		logger: logger,
	}
}

// CreateTenantSchema creates a new schema for a tenant
func (m *TenantDBManager) CreateTenantSchema(ctx context.Context, tenantID string) error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		m.config.Host,
		m.config.Port,
		m.config.User,
		m.config.Password,
		m.config.DBName,
		m.config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create schema for tenant
	schemaName := "tenant_" + tenantID
	err = db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// m.logger.Info("Created new schema for tenant").With("tenant_id", tenantID).With("schema", schemaName).Send()

	return nil
}

// GetDB gets a database instance for the current tenant
func (m *TenantDBManager) GetDB(ctx context.Context) (*gorm.DB, error) {
	// tenantID, ok := pdcontext.GetTenantID(ctx)
	// if !ok {
	// 	return nil, pdcontext.ErrUnauthorized
	// }
	tenantID := "tenant_1"

	// Try to get from pool first
	if db, ok := m.pool.Load(tenantID); ok {
		return db.(*gorm.DB), nil
	}

	// Create new database connection
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		m.config.Host,
		m.config.Port,
		m.config.User,
		m.config.Password,
		m.config.DBName,
		m.config.SSLMode,
	)

	// Configure GORM logger
	gormLogger := logger.New(
		&GormLogger{Logger: m.logger},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set schema for this connection
	schemaName := "tenant_" + tenantID
	err = db.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)).Error
	if err != nil {
		// If schema doesn't exist, try to create it
		if err := m.CreateTenantSchema(ctx, tenantID); err != nil {
			return nil, fmt.Errorf("failed to create tenant schema: %w", err)
		}
		// Try setting schema again
		err = db.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)).Error
		if err != nil {
			return nil, fmt.Errorf("failed to set schema: %w", err)
		}
	}

	// Store in pool
	m.pool.Store(tenantID, db)
	m.logger.Info("Created new database connection for tenant", "tenant_id", tenantID, "schema", "schemaName")

	return db, nil
}

// Close closes all database connections
func (m *TenantDBManager) Close() {
	m.pool.Range(func(key, value interface{}) bool {
		db := value.(*gorm.DB)
		sqlDB, err := db.DB()
		if err != nil {
			m.logger.Error("Failed to get underlying SQL DB", "tenant_id", key.(string), "err", err)
			return true
		}
		if err := sqlDB.Close(); err != nil {
			// m.logger.Error("Failed to close database connection").With("tenant_id", key.(string)).Err(err).Send()
		}
		// m.logger.Info("Closed database connection for tenant").With("tenant_id", key.(string)).Send()
		return true
	})
	m.pool = sync.Map{}
}

// GormLogger implements gorm.Logger interface
type GormLogger struct {
	pdlog.Logger
}

func (l *GormLogger) Printf(format string, args ...interface{}) {
	// l.Info(msg, args)
}
