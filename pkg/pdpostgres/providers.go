package pdpostgres

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Config struct {
	Provider string `mapstructure:"provider"` // "real" | "mock"
	URI      string `mapstructure:"uri"`
}

type ProviderFn func(ctx context.Context, cfg Config) (*gorm.DB, error)

var RealProvider ProviderFn = func(ctx context.Context, cfg Config) (*gorm.DB, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("postgres URI is empty (provider=real)")
	}
	db, err := gorm.Open(postgres.Open(cfg.URI), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("postgres connect failed: %w", err)
	}
	return db, nil
}

var MockProvider ProviderFn = func(ctx context.Context, _ Config) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
}
