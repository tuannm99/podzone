package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
)

func TestProvider_DockerProvisionProducesConnectionFromRuntime(t *testing.T) {
	p := NewProvider(onboardingconfig.StoreProvisioningConfig{
		Runtime:       "local_docker",
		ClusterName:   "pg-default",
		Mode:          "schema",
		DBName:        "postgres",
		SchemaPrefix:  "t_",
		DockerNetwork: "podzone_default",
	})
	request := entity.StorePlacementRequest{
		RequestID: "request-1",
		TenantID:  "2e0df8f6-4964-447d-a287-67eabd0e65c9",
		StoreID:   "store-1",
	}

	plan, err := p.PlanStorePlacement(context.Background(), request)
	require.NoError(t, err)
	require.Equal(t, entity.PlacementRuntimeLocalDocker, plan.Runtime)
	require.Equal(t, "t_tenant_2e0df8f6_4964_447d_a287_67eabd0e65c9", plan.SchemaName)

	allocation := p.allocate(request, plan, p.dockerConnection(plan), "ready")
	require.Equal(t, "postgres://postgres:***@pgbouncer:6432/postgres", allocation.Endpoint)
	require.Equal(t, "docker/postgres/default", allocation.SecretRef)
	require.Equal(t, "docker_runtime", allocation.ProviderMeta["connection_source"])
}

func TestProvider_KubernetesProvisionProducesConnectionFromRuntime(t *testing.T) {
	p := NewProvider(onboardingconfig.StoreProvisioningConfig{
		Runtime:             "kubernetes",
		ClusterName:         "pg-default",
		Mode:                "schema",
		DBName:              "postgres",
		SchemaPrefix:        "t_",
		KubernetesNamespace: "podzone",
	})
	request := entity.StorePlacementRequest{
		RequestID: "request-1",
		TenantID:  "tenant-1",
		StoreID:   "store-1",
	}

	plan, err := p.PlanStorePlacement(context.Background(), request)
	require.NoError(t, err)
	allocation := p.allocate(request, plan, p.kubernetesConnection(plan), "ready")

	require.Equal(t, "postgres://postgres:***@pgbouncer.podzone.svc.cluster.local:6432/postgres", allocation.Endpoint)
	require.Equal(t, "k8s/podzone/postgres/default", allocation.SecretRef)
	require.Equal(t, "kubernetes_service", allocation.ProviderMeta["connection_source"])
}

func TestProvider_ProvisionRequiresAdminDSN(t *testing.T) {
	p := NewProvider(onboardingconfig.StoreProvisioningConfig{
		Runtime:      "local_docker",
		ClusterName:  "pg-default",
		Mode:         "schema",
		DBName:       "postgres",
		SchemaPrefix: "t_",
	})
	request := entity.StorePlacementRequest{
		RequestID: "request-1",
		TenantID:  "tenant-1",
		StoreID:   "store-1",
	}

	plan, err := p.PlanStorePlacement(context.Background(), request)
	require.NoError(t, err)
	_, err = p.ProvisionStorePlacement(context.Background(), request, plan)
	require.ErrorContains(t, err, "admin_dsn is required")
}

func TestProvider_TerraformRuntimeRequiresFutureAdapter(t *testing.T) {
	p := NewProvider(onboardingconfig.StoreProvisioningConfig{
		Runtime:         "terraform",
		ClusterName:     "pg-cloud",
		Mode:            "schema",
		DBName:          "postgres",
		SchemaPrefix:    "t_",
		TerraformModule: "modules/postgres-tenant",
	})
	request := entity.StorePlacementRequest{
		RequestID: "request-1",
		TenantID:  "tenant-1",
		StoreID:   "store-1",
	}

	plan, err := p.PlanStorePlacement(context.Background(), request)
	require.NoError(t, err)
	_, err = p.ProvisionStorePlacement(context.Background(), request, plan)
	require.ErrorContains(t, err, "terraform placement provider is declared but not implemented")
}
