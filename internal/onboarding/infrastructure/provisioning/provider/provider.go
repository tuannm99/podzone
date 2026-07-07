package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	infrasoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/outputport"
	"github.com/tuannm99/podzone/pkg/pdsql"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

var (
	_ infrasoutputport.CapacityChecker          = (*Provider)(nil)
	_ infrasoutputport.PlacementPolicyEvaluator = (*Provider)(nil)
	_ infrasoutputport.PlacementPlanner         = (*Provider)(nil)
	_ infrasoutputport.ResourceHealthChecker    = (*Provider)(nil)
	_ infrasoutputport.StorageProvisioner       = (*Provider)(nil)
)

type Provider struct {
	cfg       onboardingconfig.StoreProvisioningConfig
	inventory infrasoutputport.ResourceInventoryRepository
}

type ProviderParams struct {
	fx.In

	Config    onboardingconfig.StoreProvisioningConfig
	Inventory infrasoutputport.ResourceInventoryRepository
}

func NewProvider(p ProviderParams) *Provider {
	return &Provider{
		cfg:       p.Config,
		inventory: p.Inventory,
	}
}

func (p *Provider) PlanStorePlacement(
	ctx context.Context,
	request entity.StorePlacementRequest,
) (entity.PlacementPlan, error) {
	if p.inventory == nil {
		return entity.PlacementPlan{}, fmt.Errorf("resource inventory repository is not configured")
	}
	inventory, err := p.inventory.LoadResourceInventory(ctx, request)
	if err != nil {
		return entity.PlacementPlan{}, err
	}
	capacity, err := p.CheckPlacementCapacity(ctx, request, inventory)
	if err != nil {
		return entity.PlacementPlan{}, err
	}
	decision, err := p.EvaluatePlacementPolicy(ctx, request, inventory, capacity)
	if err != nil {
		return entity.PlacementPlan{}, err
	}
	if !capacity.CanPlace {
		return entity.PlacementPlan{}, fmt.Errorf(
			"placement capacity unavailable: %s",
			strings.Join(capacity.Reasons, "; "),
		)
	}
	if decision.ApprovalRequired && !decision.AutoApproved {
		return entity.PlacementPlan{}, fmt.Errorf(
			"placement requires approval: %s",
			strings.Join(decision.Reasons, "; "),
		)
	}

	switch normalizeRuntime(p.cfg.Runtime) {
	case entity.PlacementRuntimeLocalDocker, entity.PlacementRuntimeDocker:
		return p.planDocker(request, inventory, capacity, decision), nil
	case entity.PlacementRuntimeKubernetes, entity.PlacementRuntimeK8s:
		return p.planKubernetes(request, inventory, capacity, decision), nil
	case entity.PlacementRuntimeTerraform:
		return p.planTerraform(request, inventory, capacity, decision)
	default:
		return entity.PlacementPlan{}, fmt.Errorf("unsupported placement runtime: %s", p.cfg.Runtime)
	}
}

func (p *Provider) CheckPlacementCapacity(
	_ context.Context,
	_ entity.StorePlacementRequest,
	inventory entity.ResourceInventory,
) (entity.CapacitySnapshot, error) {
	snapshot := entity.CapacitySnapshot{
		CanPlace: true,
	}

	dbCluster := selectDatabaseCluster(inventory, p.cfg.ClusterName)
	if dbCluster == nil {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "database cluster not found")
	} else {
		snapshot.DBClusterName = dbCluster.Name
		snapshot.DatabaseName = toolkit.FirstNonEmpty(dbCluster.PlacementDB, p.cfg.DBName, "podzone_tenants")
		appendDBCapacityReasons(&snapshot, *dbCluster)
	}

	namespace := selectNamespace(inventory, p.cfg.KubernetesNamespace)
	if namespace == nil {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "kubernetes namespace not found")
	} else {
		snapshot.NamespaceName = namespace.Name
		appendNamespaceCapacityReasons(&snapshot, *namespace)
	}

	runtimePool := selectRuntimePool(
		inventory,
		runtimePoolName(normalizeRuntime(p.cfg.Runtime), snapshot.NamespaceName),
		normalizeRuntime(p.cfg.Runtime),
	)
	if runtimePool == nil {
		snapshot.CanPlace = false
		snapshot.Reasons = append(snapshot.Reasons, "runtime pool not found")
	} else {
		snapshot.RuntimePool = runtimePool.Name
		appendRuntimePoolCapacityReasons(&snapshot, *runtimePool)
	}

	return snapshot, nil
}

func (p *Provider) EvaluatePlacementPolicy(
	_ context.Context,
	_ entity.StorePlacementRequest,
	_ entity.ResourceInventory,
	capacity entity.CapacitySnapshot,
) (entity.PlacementPolicyDecision, error) {
	if !capacity.CanPlace {
		return entity.PlacementPolicyDecision{
			ApprovalRequired: true,
			Reasons:          capacity.Reasons,
		}, nil
	}
	return entity.PlacementPolicyDecision{
		AutoApproved: true,
	}, nil
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
		return entity.PlacementAllocation{}, fmt.Errorf("kubernetes placement provider is declared but not implemented")
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

func (p *Provider) planDocker(
	request entity.StorePlacementRequest,
	inventory entity.ResourceInventory,
	capacity entity.CapacitySnapshot,
	decision entity.PlacementPolicyDecision,
) entity.PlacementPlan {
	now := time.Now().UTC()
	return entity.PlacementPlan{
		RequestID:   request.RequestID,
		TenantID:    request.TenantID,
		StoreID:     request.StoreID,
		Runtime:     entity.PlacementRuntimeLocalDocker,
		ClusterName: capacity.DBClusterName,
		Mode:        toolkit.FirstNonEmpty(p.cfg.Mode, "schema"),
		DBName:      capacity.DatabaseName,
		SchemaName:  toolkit.SchemaName(toolkit.FirstNonEmpty(p.cfg.SchemaPrefix, "t_"), request.TenantID),
		ProviderMeta: map[string]string{
			"provider":          "docker",
			"runtime":           string(entity.PlacementRuntimeLocalDocker),
			"network":           toolkit.FirstNonEmpty(p.cfg.DockerNetwork, "docker_default"),
			"postgres_service":  "postgres",
			"pgbouncer_service": "pgbouncer",
			"strategy":          "shared_postgres_schema",
			"runtime_pool":      capacity.RuntimePool,
		},
		InventorySnapshot: inventory,
		CapacitySnapshot:  capacity,
		PolicyDecision:    decision,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func (p *Provider) planKubernetes(
	request entity.StorePlacementRequest,
	inventory entity.ResourceInventory,
	capacity entity.CapacitySnapshot,
	decision entity.PlacementPolicyDecision,
) entity.PlacementPlan {
	namespace := capacity.NamespaceName
	now := time.Now().UTC()
	return entity.PlacementPlan{
		RequestID:   request.RequestID,
		TenantID:    request.TenantID,
		StoreID:     request.StoreID,
		Runtime:     entity.PlacementRuntimeKubernetes,
		ClusterName: capacity.DBClusterName,
		Mode:        toolkit.FirstNonEmpty(p.cfg.Mode, "schema"),
		DBName:      capacity.DatabaseName,
		SchemaName:  toolkit.SchemaName(toolkit.FirstNonEmpty(p.cfg.SchemaPrefix, "t_"), request.TenantID),
		ProviderMeta: map[string]string{
			"provider":           "kubernetes",
			"runtime":            string(entity.PlacementRuntimeKubernetes),
			"namespace":          namespace,
			"postgres_service":   "postgres",
			"pgbouncer_service":  "pgbouncer",
			"provision_strategy": "service_backed_schema",
			"runtime_pool":       capacity.RuntimePool,
		},
		InventorySnapshot: inventory,
		CapacitySnapshot:  capacity,
		PolicyDecision:    decision,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func (p *Provider) planTerraform(
	request entity.StorePlacementRequest,
	inventory entity.ResourceInventory,
	capacity entity.CapacitySnapshot,
	decision entity.PlacementPolicyDecision,
) (entity.PlacementPlan, error) {
	if p.cfg.TerraformModule == "" {
		return entity.PlacementPlan{}, fmt.Errorf("terraform_module is required for terraform placement runtime")
	}
	now := time.Now().UTC()
	return entity.PlacementPlan{
		RequestID:   request.RequestID,
		TenantID:    request.TenantID,
		StoreID:     request.StoreID,
		Runtime:     entity.PlacementRuntimeTerraform,
		ClusterName: capacity.DBClusterName,
		Mode:        toolkit.FirstNonEmpty(p.cfg.Mode, "schema"),
		DBName:      capacity.DatabaseName,
		SchemaName:  toolkit.SchemaName(toolkit.FirstNonEmpty(p.cfg.SchemaPrefix, "t_"), request.TenantID),
		ProviderMeta: map[string]string{
			"provider":     "terraform",
			"runtime":      string(entity.PlacementRuntimeTerraform),
			"workspace":    toolkit.FirstNonEmpty(p.cfg.TerraformWorkspace, "default"),
			"module":       p.cfg.TerraformModule,
			"strategy":     "future_adapter",
			"runtime_pool": capacity.RuntimePool,
		},
		InventorySnapshot: inventory,
		CapacitySnapshot:  capacity,
		PolicyDecision:    decision,
		CreatedAt:         now,
		UpdatedAt:         now,
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
