package config

import (
	"time"

	"github.com/knadh/koanf/v2"

	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
)

type StoreProvisioningConfig struct {
	Enabled      bool          `koanf:"enabled"       mapstructure:"enabled"`
	AutoApprove  bool          `koanf:"auto_approve"  mapstructure:"auto_approve"`
	Interval     time.Duration `koanf:"interval"      mapstructure:"interval"`
	ClusterName  string        `koanf:"cluster_name"  mapstructure:"cluster_name"`
	Mode         string        `koanf:"mode"          mapstructure:"mode"`
	DBName       string        `koanf:"db_name"       mapstructure:"db_name"`
	SchemaPrefix string        `koanf:"schema_prefix" mapstructure:"schema_prefix"`
	Endpoint     string        `koanf:"endpoint"      mapstructure:"endpoint"`
	SecretRef    string        `koanf:"secret_ref"    mapstructure:"secret_ref"`
}

func DefaultStoreProvisioningConfig() StoreProvisioningConfig {
	return StoreProvisioningConfig{
		Enabled:      true,
		AutoApprove:  true,
		Interval:     5 * time.Second,
		ClusterName:  "pg-default",
		Mode:         "schema",
		DBName:       "postgres",
		SchemaPrefix: "t_",
		Endpoint:     "postgres://postgres:***@pgbouncer:6432/postgres",
		SecretRef:    "postgres/default",
	}
}

func NewStoreProvisioningConfig(k *koanf.Koanf) StoreProvisioningConfig {
	cfg := DefaultStoreProvisioningConfig()
	if k != nil {
		_ = k.Unmarshal("onboarding.store_provisioning", &cfg)
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Second
	}
	if cfg.ClusterName == "" {
		cfg.ClusterName = "pg-default"
	}
	if cfg.Mode == "" {
		cfg.Mode = "schema"
	}
	if cfg.DBName == "" {
		cfg.DBName = "postgres"
	}
	if cfg.SchemaPrefix == "" {
		cfg.SchemaPrefix = "t_"
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "postgres://postgres:***@pgbouncer:6432/postgres"
	}
	if cfg.SecretRef == "" {
		cfg.SecretRef = "postgres/default"
	}
	return cfg
}

func NewStoreProvisioningDomainConfig(cfg StoreProvisioningConfig) storeentity.ProvisioningConfig {
	return storeentity.ProvisioningConfig{
		Enabled:      cfg.Enabled,
		AutoApprove:  cfg.AutoApprove,
		ClusterName:  cfg.ClusterName,
		Mode:         cfg.Mode,
		DBName:       cfg.DBName,
		SchemaPrefix: cfg.SchemaPrefix,
		Endpoint:     cfg.Endpoint,
		SecretRef:    cfg.SecretRef,
	}
}
