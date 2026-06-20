package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	infrasoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/outputport"
	"github.com/tuannm99/podzone/pkg/pdsql"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

var (
	_ infrasoutputport.PlacementPlanner   = (*Provider)(nil)
	_ infrasoutputport.StorageProvisioner = (*Provider)(nil)
)

type Provider struct {
	cfg onboardingconfig.StoreProvisioningConfig
}

func NewProvider(cfg onboardingconfig.StoreProvisioningConfig) *Provider {
	return &Provider{cfg: cfg}
}

func (p *Provider) PlanStorePlacement(
	_ context.Context,
	request entity.StorePlacementRequest,
) (entity.PlacementPlan, error) {
	switch normalizeRuntime(p.cfg.Runtime) {
	case entity.PlacementRuntimeLocalDocker, entity.PlacementRuntimeDocker:
		return p.planDocker(request), nil
	case entity.PlacementRuntimeKubernetes, entity.PlacementRuntimeK8s:
		return p.planKubernetes(request), nil
	case entity.PlacementRuntimeTerraform:
		return p.planTerraform(request)
	default:
		return entity.PlacementPlan{}, fmt.Errorf("unsupported placement runtime: %s", p.cfg.Runtime)
	}
}

func (p *Provider) ProvisionStorePlacement(
	ctx context.Context,
	request entity.StorePlacementRequest,
	plan entity.PlacementPlan,
) (entity.PlacementAllocation, error) {
	switch normalizeRuntime(string(plan.Runtime)) {
	case entity.PlacementRuntimeLocalDocker, entity.PlacementRuntimeDocker:
		if err := p.provisionPostgresSchema(ctx, plan); err != nil {
			return entity.PlacementAllocation{}, err
		}
		return p.allocate(request, plan, p.dockerConnection(plan)), nil
	case entity.PlacementRuntimeKubernetes, entity.PlacementRuntimeK8s:
		if err := p.provisionPostgresSchema(ctx, plan); err != nil {
			return entity.PlacementAllocation{}, err
		}
		return p.allocate(request, plan, p.kubernetesConnection(plan)), nil
	case entity.PlacementRuntimeTerraform:
		return entity.PlacementAllocation{}, fmt.Errorf("terraform placement provider is declared but not implemented")
	default:
		return entity.PlacementAllocation{}, fmt.Errorf("unsupported placement runtime: %s", plan.Runtime)
	}
}

func (p *Provider) provisionPostgresSchema(ctx context.Context, plan entity.PlacementPlan) error {
	if strings.TrimSpace(p.cfg.AdminDSN) == "" {
		return fmt.Errorf("postgres admin_dsn is required for %s placement provisioning", plan.Runtime)
	}
	if plan.Mode != "schema" {
		return fmt.Errorf("unsupported postgres placement mode %q", plan.Mode)
	}
	if err := pdsql.EnsurePostgresDatabase(p.cfg.AdminDSN, plan.DBName); err != nil {
		return fmt.Errorf("ensure postgres database %q: %w", plan.DBName, err)
	}
	targetDSN, err := pdsql.PostgresDSNWithDatabase(p.cfg.AdminDSN, plan.DBName)
	if err != nil {
		return fmt.Errorf("build postgres dsn for database %q: %w", plan.DBName, err)
	}
	if err := pdsql.EnsurePostgresSchema(ctx, targetDSN, plan.SchemaName); err != nil {
		return fmt.Errorf("provision postgres schema %q: %w", plan.SchemaName, err)
	}
	return nil
}

func (p *Provider) planDocker(request entity.StorePlacementRequest) entity.PlacementPlan {
	return entity.PlacementPlan{
		Runtime:     entity.PlacementRuntimeLocalDocker,
		ClusterName: toolkit.FirstNonEmpty(p.cfg.ClusterName, "pg-default"),
		Mode:        toolkit.FirstNonEmpty(p.cfg.Mode, "schema"),
		DBName:      toolkit.FirstNonEmpty(p.cfg.DBName, "podzone_tenants"),
		SchemaName:  toolkit.SchemaName(toolkit.FirstNonEmpty(p.cfg.SchemaPrefix, "t_"), request.TenantID),
		ProviderMeta: map[string]string{
			"provider":          "docker",
			"runtime":           string(entity.PlacementRuntimeLocalDocker),
			"network":           toolkit.FirstNonEmpty(p.cfg.DockerNetwork, "docker_default"),
			"postgres_service":  "postgres",
			"pgbouncer_service": "pgbouncer",
			"strategy":          "shared_postgres_schema",
		},
	}
}

func (p *Provider) planKubernetes(request entity.StorePlacementRequest) entity.PlacementPlan {
	namespace := toolkit.FirstNonEmpty(p.cfg.KubernetesNamespace, "default")
	return entity.PlacementPlan{
		Runtime:     entity.PlacementRuntimeKubernetes,
		ClusterName: toolkit.FirstNonEmpty(p.cfg.ClusterName, "pg-default"),
		Mode:        toolkit.FirstNonEmpty(p.cfg.Mode, "schema"),
		DBName:      toolkit.FirstNonEmpty(p.cfg.DBName, "podzone_tenants"),
		SchemaName:  toolkit.SchemaName(toolkit.FirstNonEmpty(p.cfg.SchemaPrefix, "t_"), request.TenantID),
		ProviderMeta: map[string]string{
			"provider":           "kubernetes",
			"runtime":            string(entity.PlacementRuntimeKubernetes),
			"namespace":          namespace,
			"postgres_service":   "postgres",
			"pgbouncer_service":  "pgbouncer",
			"provision_strategy": "service_backed_schema",
		},
	}
}

func (p *Provider) planTerraform(request entity.StorePlacementRequest) (entity.PlacementPlan, error) {
	if p.cfg.TerraformModule == "" {
		return entity.PlacementPlan{}, fmt.Errorf("terraform_module is required for terraform placement runtime")
	}
	return entity.PlacementPlan{
		Runtime:     entity.PlacementRuntimeTerraform,
		ClusterName: toolkit.FirstNonEmpty(p.cfg.ClusterName, "pg-default"),
		Mode:        toolkit.FirstNonEmpty(p.cfg.Mode, "schema"),
		DBName:      toolkit.FirstNonEmpty(p.cfg.DBName, "podzone_tenants"),
		SchemaName:  toolkit.SchemaName(toolkit.FirstNonEmpty(p.cfg.SchemaPrefix, "t_"), request.TenantID),
		ProviderMeta: map[string]string{
			"provider":  "terraform",
			"runtime":   string(entity.PlacementRuntimeTerraform),
			"workspace": toolkit.FirstNonEmpty(p.cfg.TerraformWorkspace, "default"),
			"module":    p.cfg.TerraformModule,
			"strategy":  "future_adapter",
		},
	}, nil
}

type connectionResult struct {
	endpoint  string
	secretRef string
}

func (p *Provider) dockerConnection(plan entity.PlacementPlan) connectionResult {
	plan.ProviderMeta["connection_source"] = "docker_runtime"
	return connectionResult{
		endpoint:  fmt.Sprintf("postgres://postgres:***@pgbouncer:6432/%s", plan.DBName),
		secretRef: "docker/postgres/default",
	}
}

func (p *Provider) kubernetesConnection(plan entity.PlacementPlan) connectionResult {
	namespace := toolkit.FirstNonEmpty(plan.ProviderMeta["namespace"], p.cfg.KubernetesNamespace, "default")
	plan.ProviderMeta["connection_source"] = "kubernetes_service"
	return connectionResult{
		endpoint: fmt.Sprintf(
			"postgres://postgres:***@pgbouncer.%s.svc.cluster.local:6432/%s",
			namespace,
			plan.DBName,
		),
		secretRef: fmt.Sprintf("k8s/%s/postgres/default", namespace),
	}
}

func (p *Provider) allocate(
	request entity.StorePlacementRequest,
	plan entity.PlacementPlan,
	conn connectionResult,
) entity.PlacementAllocation {
	now := time.Now().UTC()
	return entity.PlacementAllocation{
		ID:           uuid.NewString(),
		RequestID:    request.RequestID,
		TenantID:     request.TenantID,
		StoreID:      request.StoreID,
		Runtime:      plan.Runtime,
		ClusterName:  plan.ClusterName,
		Mode:         plan.Mode,
		DBName:       plan.DBName,
		SchemaName:   plan.SchemaName,
		Endpoint:     conn.endpoint,
		SecretRef:    conn.secretRef,
		Status:       "ready",
		ProviderMeta: plan.ProviderMeta,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func normalizeRuntime(value string) entity.PlacementRuntime {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "local", "local_docker":
		return entity.PlacementRuntimeLocalDocker
	case "docker":
		return entity.PlacementRuntimeDocker
	case "k8s", "kubernetes":
		return entity.PlacementRuntimeKubernetes
	case "terraform":
		return entity.PlacementRuntimeTerraform
	default:
		return entity.PlacementRuntime(value)
	}
}
