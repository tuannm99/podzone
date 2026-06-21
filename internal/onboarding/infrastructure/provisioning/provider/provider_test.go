package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	infrasmocks "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/outputport/mocks"
)

func TestProvider_DockerProvisionProducesConnectionFromRuntime(t *testing.T) {
	request := entity.StorePlacementRequest{
		RequestID: "request-1",
		TenantID:  "2e0df8f6-4964-447d-a287-67eabd0e65c9",
		StoreID:   "store-1",
	}
	cfg := testProvisioningConfig(onboardingconfig.StoreProvisioningConfig{
		Runtime:       "local_docker",
		ClusterName:   "pg-default",
		Mode:          "schema",
		DBName:        "podzone_tenants",
		SchemaPrefix:  "t_",
		DockerNetwork: "podzone_default",
	})
	inventory := infrasmocks.NewMockResourceInventoryRepository(t)
	inventory.EXPECT().
		LoadResourceInventory(mock.Anything, request).
		Return(testResourceInventory(cfg), nil)
	p := NewProvider(ProviderParams{
		Config:    cfg,
		Inventory: inventory,
	})

	plan, err := p.PlanStorePlacement(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, entity.PlacementRuntimeLocalDocker, plan.Runtime)
	require.Equal(t, "t_tenant_2e0df8f6_4964_447d_a287_67eabd0e65c9", plan.SchemaName)
	require.Equal(t, "docker/default", plan.ProviderMeta["runtime_pool"])

	allocation := p.allocate(request, plan, p.dockerConnection(plan))
	require.Equal(t, "postgres://postgres:***@pgbouncer:6432/podzone_tenants", allocation.Endpoint)
	require.Equal(t, "docker/postgres/default", allocation.SecretRef)
	require.Equal(t, "docker_runtime", allocation.ProviderMeta["connection_source"])
}

func TestProvider_KubernetesPlanIsAllowedButProvisioningIsNotImplemented(t *testing.T) {
	request := entity.StorePlacementRequest{
		RequestID: "request-1",
		TenantID:  "tenant-1",
		StoreID:   "store-1",
	}
	cfg := testProvisioningConfig(onboardingconfig.StoreProvisioningConfig{
		Runtime:             "kubernetes",
		ClusterName:         "pg-default",
		Mode:                "schema",
		DBName:              "podzone_tenants",
		SchemaPrefix:        "t_",
		KubernetesNamespace: "podzone",
	})
	inventory := infrasmocks.NewMockResourceInventoryRepository(t)
	inventory.EXPECT().
		LoadResourceInventory(mock.Anything, request).
		Return(testResourceInventory(cfg), nil)
	p := NewProvider(ProviderParams{
		Config:    cfg,
		Inventory: inventory,
	})

	plan, err := p.PlanStorePlacement(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, entity.PlacementRuntimeKubernetes, plan.Runtime)
	require.Equal(t, "podzone", plan.ProviderMeta["namespace"])
	require.Equal(t, "k8s/podzone", plan.ProviderMeta["runtime_pool"])

	_, err = p.ProvisionStorePlacement(context.Background(), request, plan)
	require.ErrorContains(t, err, "kubernetes placement provider is declared but not implemented")
}

func TestProvider_PlanUsesDeclaredInventoryWhenConfigPreferenceIsMissing(t *testing.T) {
	request := entity.StorePlacementRequest{
		RequestID: "request-1",
		TenantID:  "tenant-1",
		StoreID:   "store-1",
	}
	cfg := testProvisioningConfig(onboardingconfig.StoreProvisioningConfig{
		Runtime:             "kubernetes",
		ClusterName:         "pg-default",
		Mode:                "schema",
		DBName:              "podzone_tenants",
		SchemaPrefix:        "t_",
		KubernetesNamespace: "default",
	})
	declared := testResourceInventory(cfg)
	declared.DBClusters[0].Name = "pg-east-1"
	declared.DBClusters[0].PlacementDB = "podzone_tenants_east"
	declared.K8sClusters[0].Namespaces[0].Name = "podzone-east"
	declared.RuntimePools[0].Name = "k8s/podzone-east"

	inventory := infrasmocks.NewMockResourceInventoryRepository(t)
	inventory.EXPECT().
		LoadResourceInventory(mock.Anything, request).
		Return(declared, nil)
	p := NewProvider(ProviderParams{
		Config:    cfg,
		Inventory: inventory,
	})

	plan, err := p.PlanStorePlacement(context.Background(), request)
	require.NoError(t, err)

	require.Equal(t, "pg-east-1", plan.ClusterName)
	require.Equal(t, "podzone_tenants_east", plan.DBName)
	require.Equal(t, "podzone-east", plan.ProviderMeta["namespace"])
	require.Equal(t, "k8s/podzone-east", plan.ProviderMeta["runtime_pool"])
}

func TestProvider_ProvisionRequiresAdminDSN(t *testing.T) {
	request := entity.StorePlacementRequest{
		RequestID: "request-1",
		TenantID:  "tenant-1",
		StoreID:   "store-1",
	}
	cfg := testProvisioningConfig(onboardingconfig.StoreProvisioningConfig{
		Runtime:      "local_docker",
		ClusterName:  "pg-default",
		Mode:         "schema",
		DBName:       "podzone_tenants",
		SchemaPrefix: "t_",
	})
	inventory := infrasmocks.NewMockResourceInventoryRepository(t)
	inventory.EXPECT().
		LoadResourceInventory(mock.Anything, request).
		Return(testResourceInventory(cfg), nil)
	p := NewProvider(ProviderParams{
		Config:    cfg,
		Inventory: inventory,
	})

	plan, err := p.PlanStorePlacement(context.Background(), request)
	require.NoError(t, err)
	_, err = p.ProvisionStorePlacement(context.Background(), request, plan)
	require.ErrorContains(t, err, "admin_dsn is required")
}

func TestProvider_TerraformRuntimeRequiresFutureAdapter(t *testing.T) {
	request := entity.StorePlacementRequest{
		RequestID: "request-1",
		TenantID:  "tenant-1",
		StoreID:   "store-1",
	}
	cfg := testProvisioningConfig(onboardingconfig.StoreProvisioningConfig{
		Runtime:         "terraform",
		ClusterName:     "pg-cloud",
		Mode:            "schema",
		DBName:          "podzone_tenants",
		SchemaPrefix:    "t_",
		TerraformModule: "modules/postgres-tenant",
	})
	inventory := infrasmocks.NewMockResourceInventoryRepository(t)
	inventory.EXPECT().
		LoadResourceInventory(mock.Anything, request).
		Return(testResourceInventory(cfg), nil)
	p := NewProvider(ProviderParams{
		Config:    cfg,
		Inventory: inventory,
	})

	plan, err := p.PlanStorePlacement(context.Background(), request)
	require.NoError(t, err)
	_, err = p.ProvisionStorePlacement(context.Background(), request, plan)
	require.ErrorContains(t, err, "terraform placement provider is declared but not implemented")
}

func TestProvider_PlanFailsWhenResourceInventoryIsNotConfigured(t *testing.T) {
	request := entity.StorePlacementRequest{
		RequestID: "request-1",
		TenantID:  "tenant-1",
		StoreID:   "store-1",
	}
	cfg := testProvisioningConfig(onboardingconfig.StoreProvisioningConfig{
		Runtime:      "local_docker",
		ClusterName:  "pg-default",
		Mode:         "schema",
		DBName:       "podzone_tenants",
		SchemaPrefix: "t_",
	})
	inventory := infrasmocks.NewMockResourceInventoryRepository(t)
	inventory.EXPECT().
		LoadResourceInventory(mock.Anything, request).
		Return(entity.ResourceInventory{}, errors.New("resource inventory is not configured"))
	p := NewProvider(ProviderParams{
		Config:    cfg,
		Inventory: inventory,
	})

	_, err := p.PlanStorePlacement(context.Background(), request)
	require.ErrorContains(t, err, "resource inventory is not configured")
}

func TestProvider_CapacityCheckFailsWhenNamespaceIsFull(t *testing.T) {
	cfg := testProvisioningConfig(onboardingconfig.StoreProvisioningConfig{
		Runtime:                "kubernetes",
		ClusterName:            "pg-default",
		Mode:                   "schema",
		DBName:                 "podzone_tenants",
		SchemaPrefix:           "t_",
		KubernetesNamespace:    "podzone",
		MaxTenantsPerNamespace: 1,
	})
	p := NewProvider(ProviderParams{
		Config: cfg,
	})
	inventory := testResourceInventory(cfg)
	inventory.K8sClusters[0].Namespaces[0].CurrentTenants = 1

	capacity, err := p.CheckPlacementCapacity(context.Background(), entity.StorePlacementRequest{}, inventory)
	require.NoError(t, err)

	require.False(t, capacity.CanPlace)
	require.Contains(t, capacity.Reasons, "kubernetes namespace tenant capacity exceeded")
}

func testResourceInventory(cfg onboardingconfig.StoreProvisioningConfig) entity.ResourceInventory {
	namespace := cfg.KubernetesNamespace
	if namespace == "" {
		namespace = "default"
	}
	runtimePool := "docker/default"
	if cfg.Runtime == "kubernetes" || cfg.Runtime == "k8s" {
		runtimePool = "k8s/" + namespace
	}
	if cfg.Runtime == "terraform" {
		runtimePool = "terraform/default"
	}
	return entity.ResourceInventory{
		Environment: "test",
		DBClusters: []entity.DatabaseCluster{
			{
				Name:           cfg.ClusterName,
				Engine:         "postgres",
				PlacementDB:    cfg.DBName,
				MaxTenants:     cfg.MaxTenantsPerDBCluster,
				MaxSchemas:     cfg.MaxSchemasPerDatabase,
				MaxConnections: cfg.MaxConnectionsPerDBCluster,
				Healthy:        true,
			},
		},
		K8sClusters: []entity.KubernetesCluster{
			{
				Name:    cfg.ClusterName,
				Healthy: true,
				Namespaces: []entity.KubernetesNamespace{
					{
						Name:       namespace,
						MaxTenants: cfg.MaxTenantsPerNamespace,
						CPUMilli:   cfg.NamespaceCPUMilli,
						MemoryMi:   cfg.NamespaceMemoryMi,
						Healthy:    true,
					},
				},
			},
		},
		RuntimePools: []entity.RuntimePool{
			{
				Name:       runtimePool,
				Kind:       cfg.Runtime,
				MaxTenants: cfg.RuntimePoolCapacity,
				Healthy:    true,
			},
		},
	}
}

func testProvisioningConfig(cfg onboardingconfig.StoreProvisioningConfig) onboardingconfig.StoreProvisioningConfig {
	if cfg.MaxTenantsPerDBCluster == 0 {
		cfg.MaxTenantsPerDBCluster = 100
	}
	if cfg.MaxSchemasPerDatabase == 0 {
		cfg.MaxSchemasPerDatabase = 500
	}
	if cfg.MaxConnectionsPerDBCluster == 0 {
		cfg.MaxConnectionsPerDBCluster = 1000
	}
	if cfg.MaxTenantsPerNamespace == 0 {
		cfg.MaxTenantsPerNamespace = 25
	}
	if cfg.NamespaceCPUMilli == 0 {
		cfg.NamespaceCPUMilli = 1000
	}
	if cfg.NamespaceMemoryMi == 0 {
		cfg.NamespaceMemoryMi = 1024
	}
	if cfg.RuntimePoolCapacity == 0 {
		cfg.RuntimePoolCapacity = 100
	}
	return cfg
}
