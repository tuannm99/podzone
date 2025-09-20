package pdpostgres

import (
	"context"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var PostgresFactory = func(ctx context.Context, cfg InstanceConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.URI), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("postgres connect failed: %w", err)
	}

	return db, nil
}

var NoopPostgresFactory = func(ctx context.Context, _ InstanceConfig) (*gorm.DB, error) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		return nil, err
	}
	gdb, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return gdb, nil
}
