package pdpostgres

import (
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GetConfigFromViper(name string, v *viper.Viper) (*Config, error) {
	base := "gorm." + name
	var cfg Config
	if sub := v.Sub(base); sub != nil {
		_ = sub.Unmarshal(&cfg)
	}
	return &cfg, nil
}

func NewDbFromConfig(cfg *Config) (*gorm.DB, error) {
	switch cfg.Provider {
	case "", PostgresProvider:
		return NewPostgresDB(cfg)
	case SqliteProvider:
		return NewSqliteDB(cfg)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

func NewPostgresDB(cfg *Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.URI), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("postgres connect failed: %w", err)
	}
	return db, nil
}

func NewSqliteDB(cfg *Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.URI), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("sqlite connect failed: %w", err)
	}
	return db, nil
}
