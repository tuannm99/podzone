package pdsql

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func GetConfigFromViper(name string, v *viper.Viper) (*Config, error) {
	base := "sql." + name
	var cfg Config
	if sub := v.Sub(base); sub != nil {
		_ = sub.Unmarshal(&cfg)
	}
	return &cfg, nil
}

func NewDbFromConfig(cfg *Config) (*sqlx.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil config")
	}
	driver := string(cfg.Provider)

	switch driver {
	case "postgres", "pgx", "sqlmock":
		db, err := sqlx.Open(driver, cfg.URI)
		if err != nil {
			return nil, err
		}
		return db, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}
