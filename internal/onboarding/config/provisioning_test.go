package config

import (
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

func TestNewStoreProvisioningConfig_UsesAdminDSNFromEnvironment(t *testing.T) {
	t.Setenv("ONBOARDING_POSTGRES_ADMIN_DSN", "postgres://admin:secret@postgres:5432/postgres?sslmode=disable")
	k := koanf.New(".")
	require.NoError(t, k.Set("onboarding.store_provisioning.admin_dsn", "${ONBOARDING_POSTGRES_ADMIN_DSN}"))

	cfg := NewStoreProvisioningConfig(k)

	require.Equal(
		t,
		"postgres://admin:secret@postgres:5432/postgres?sslmode=disable",
		cfg.AdminDSN,
	)
}

func TestNewStoreProvisioningConfig_DefaultsTenantPlacementDatabase(t *testing.T) {
	cfg := NewStoreProvisioningConfig(koanf.New("."))

	require.Equal(t, "podzone_tenants", cfg.DBName)
}

func TestNewStoreProvisioningConfig_DefaultsPlacementCapacityPolicy(t *testing.T) {
	cfg := NewStoreProvisioningConfig(koanf.New("."))

	require.Equal(t, 5, cfg.BatchSize)
	require.Equal(t, 2*time.Minute, cfg.LeaseTTL)
	require.Equal(t, 100, cfg.MaxTenantsPerDBCluster)
	require.Equal(t, 500, cfg.MaxSchemasPerDatabase)
	require.Equal(t, 1000, cfg.MaxConnectionsPerDBCluster)
	require.Equal(t, 25, cfg.MaxTenantsPerNamespace)
	require.Equal(t, 1000, cfg.NamespaceCPUMilli)
	require.Equal(t, 1024, cfg.NamespaceMemoryMi)
	require.Equal(t, 100, cfg.RuntimePoolCapacity)
}

func TestNewStoreProvisioningConfig_DropsUnexpandedAdminDSN(t *testing.T) {
	k := koanf.New(".")
	require.NoError(t, k.Set("onboarding.store_provisioning.admin_dsn", "${ONBOARDING_POSTGRES_ADMIN_DSN}"))

	cfg := NewStoreProvisioningConfig(k)

	require.Empty(t, cfg.AdminDSN)
}
