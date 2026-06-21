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

	MaxTenantsPerDBCluster     int `koanf:"max_tenants_per_db_cluster"     mapstructure:"max_tenants_per_db_cluster"`
	MaxSchemasPerDatabase      int `koanf:"max_schemas_per_database"       mapstructure:"max_schemas_per_database"`
	MaxConnectionsPerDBCluster int `koanf:"max_connections_per_db_cluster" mapstructure:"max_connections_per_db_cluster"`
	MaxTenantsPerNamespace     int `koanf:"max_tenants_per_namespace"      mapstructure:"max_tenants_per_namespace"`
	NamespaceCPUMilli          int `koanf:"namespace_cpu_milli"            mapstructure:"namespace_cpu_milli"`
	NamespaceMemoryMi          int `koanf:"namespace_memory_mi"            mapstructure:"namespace_memory_mi"`
	RuntimePoolCapacity        int `koanf:"runtime_pool_capacity"          mapstructure:"runtime_pool_capacity"`
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

		MaxTenantsPerDBCluster:     100,
		MaxSchemasPerDatabase:      500,
		MaxConnectionsPerDBCluster: 1000,
		MaxTenantsPerNamespace:     25,
		NamespaceCPUMilli:          1000,
		NamespaceMemoryMi:          1024,
		RuntimePoolCapacity:        100,
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
	if cfg.MaxTenantsPerDBCluster <= 0 {
		cfg.MaxTenantsPerDBCluster = 100
	}
	if cfg.MaxSchemasPerDatabase <= 0 {
		cfg.MaxSchemasPerDatabase = 500
	}
	if cfg.MaxConnectionsPerDBCluster <= 0 {
		cfg.MaxConnectionsPerDBCluster = 1000
	}
	if cfg.MaxTenantsPerNamespace <= 0 {
		cfg.MaxTenantsPerNamespace = 25
	}
	if cfg.NamespaceCPUMilli <= 0 {
		cfg.NamespaceCPUMilli = 1000
	}
	if cfg.NamespaceMemoryMi <= 0 {
		cfg.NamespaceMemoryMi = 1024
	}
	if cfg.RuntimePoolCapacity <= 0 {
		cfg.RuntimePoolCapacity = 100
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
