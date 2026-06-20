package config

import (
	"strings"
	"time"

	"github.com/knadh/koanf/v2"

	storeentity "github.com/tuannm99/podzone/internal/onboarding/domain/store/entity"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type StoreProvisioningConfig struct {
	Enabled      bool          `koanf:"enabled"       mapstructure:"enabled"`
	AutoApprove  bool          `koanf:"auto_approve"  mapstructure:"auto_approve"`
	Interval     time.Duration `koanf:"interval"      mapstructure:"interval"`
	Runtime      string        `koanf:"runtime"       mapstructure:"runtime"`
	ClusterName  string        `koanf:"cluster_name"  mapstructure:"cluster_name"`
	Mode         string        `koanf:"mode"          mapstructure:"mode"`
	DBName       string        `koanf:"db_name"       mapstructure:"db_name"`
	SchemaPrefix string        `koanf:"schema_prefix" mapstructure:"schema_prefix"`
	AdminDSN     string        `koanf:"admin_dsn"     mapstructure:"admin_dsn"`

	DockerNetwork       string `koanf:"docker_network"       mapstructure:"docker_network"`
	KubernetesNamespace string `koanf:"kubernetes_namespace" mapstructure:"kubernetes_namespace"`
	TerraformWorkspace  string `koanf:"terraform_workspace"  mapstructure:"terraform_workspace"`
	TerraformModule     string `koanf:"terraform_module"     mapstructure:"terraform_module"`
}

func DefaultStoreProvisioningConfig() StoreProvisioningConfig {
	return StoreProvisioningConfig{
		Enabled:      true,
		AutoApprove:  true,
		Interval:     5 * time.Second,
		Runtime:      "local_docker",
		ClusterName:  "pg-default",
		Mode:         "schema",
		DBName:       "podzone_tenants",
		SchemaPrefix: "t_",

		DockerNetwork:       "docker_default",
		KubernetesNamespace: "default",
		TerraformWorkspace:  "default",
	}
}

func NewStoreProvisioningConfig(k *koanf.Koanf) StoreProvisioningConfig {
	cfg := DefaultStoreProvisioningConfig()
	if k != nil {
		_ = k.Unmarshal("onboarding.store_provisioning", &cfg)
	}
	if adminDSN := toolkit.GetEnv("ONBOARDING_POSTGRES_ADMIN_DSN", ""); adminDSN != "" {
		cfg.AdminDSN = adminDSN
	} else if strings.HasPrefix(cfg.AdminDSN, "${") {
		cfg.AdminDSN = ""
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 5 * time.Second
	}
	if cfg.Runtime == "" {
		cfg.Runtime = "local_docker"
	}
	if cfg.ClusterName == "" {
		cfg.ClusterName = "pg-default"
	}
	if cfg.Mode == "" {
		cfg.Mode = "schema"
	}
	if cfg.DBName == "" {
		cfg.DBName = "podzone_tenants"
	}
	if cfg.SchemaPrefix == "" {
		cfg.SchemaPrefix = "t_"
	}
	if cfg.DockerNetwork == "" {
		cfg.DockerNetwork = "docker_default"
	}
	if cfg.KubernetesNamespace == "" {
		cfg.KubernetesNamespace = "default"
	}
	if cfg.TerraformWorkspace == "" {
		cfg.TerraformWorkspace = "default"
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
	}
}
