package inputport

import (
	"context"
	"time"

	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

type UpsertConnectionRequest struct {
	InfraType entity.InfraType       `json:"infra_type" binding:"required"`
	Name      string                 `json:"name"`
	Endpoint  string                 `json:"endpoint"   binding:"required"`
	SecretRef string                 `json:"secret_ref"`
	Status    string                 `json:"status"`
	Meta      map[string]string      `json:"meta"`
	Config    map[string]interface{} `json:"config"`

	ClusterName string `json:"cluster_name"`
	Mode        string `json:"mode"`
	DBName      string `json:"db_name"`
	SchemaName  string `json:"schema_name"`
}

type UpsertConnectionResponse struct {
	CorrelationID string     `json:"correlation_id"`
	Connection    Connection `json:"connection"`
	Queued        bool       `json:"queued"`
	KVStoreKey    string     `json:"kv_store_key"`
}

type ProvisionStorePlacementRequest struct {
	RequestID   string `json:"request_id"`
	TenantID    string `json:"tenant_id"`
	StoreID     string `json:"store_id"`
	Subdomain   string `json:"subdomain"`
	RequestedBy string `json:"requested_by"`
}

type ProvisionStorePlacementResponse struct {
	CorrelationID string            `json:"correlation_id"`
	AllocationID  string            `json:"allocation_id"`
	Runtime       string            `json:"runtime"`
	ClusterName   string            `json:"cluster_name"`
	Mode          string            `json:"mode"`
	DBName        string            `json:"db_name"`
	SchemaName    string            `json:"schema_name"`
	Endpoint      string            `json:"endpoint"`
	SecretRef     string            `json:"secret_ref"`
	Status        string            `json:"status"`
	ProviderMeta  map[string]string `json:"provider_meta"`
	Queued        bool              `json:"queued"`
}

type PlacementRoute struct {
	ClusterName string `json:"cluster_name"`
	Mode        string `json:"mode"`
	DBName      string `json:"db_name"`
	SchemaName  string `json:"schema_name"`
}

type PlacementStatus struct {
	TenantID        string          `json:"tenant_id"`
	AllocationID    string          `json:"allocation_id,omitempty"`
	AllocationReady bool            `json:"allocation_ready"`
	RouteReady      bool            `json:"route_ready"`
	InSync          bool            `json:"in_sync"`
	NeedsRepair     bool            `json:"needs_repair"`
	Reason          string          `json:"reason,omitempty"`
	Allocation      *PlacementRoute `json:"allocation,omitempty"`
	Route           *PlacementRoute `json:"route,omitempty"`
	UpdatedAt       time.Time       `json:"updated_at,omitempty"`
}

type PlacementReconcileResponse struct {
	Status      PlacementStatus `json:"status"`
	Repaired    bool            `json:"repaired"`
	KVStoreKey  string          `json:"kv_store_key"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
}

type Connection struct {
	TenantID  string                 `json:"tenant_id"`
	InfraType entity.InfraType       `json:"infra_type"`
	Name      string                 `json:"name"`
	Endpoint  string                 `json:"endpoint"`
	SecretRef string                 `json:"secret_ref"`
	Status    string                 `json:"status"`
	Version   int64                  `json:"version"`
	Meta      map[string]string      `json:"meta"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty"`
}

type ListConnectionsResponse struct {
	Items    []Connection        `json:"items"`
	PageInfo collection.PageInfo `json:"pageInfo"`
}

type ConnectionEvent struct {
	ID            string                 `json:"id"`
	CorrelationID string                 `json:"correlation_id"`
	TenantID      string                 `json:"tenant_id"`
	InfraType     entity.InfraType       `json:"infra_type"`
	Name          string                 `json:"name"`
	Action        string                 `json:"action"`
	Status        string                 `json:"status"`
	Request       map[string]interface{} `json:"request,omitempty"`
	Result        map[string]interface{} `json:"result,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Actor         map[string]string      `json:"actor,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

type ListEventsResponse struct {
	Items    []ConnectionEvent   `json:"items"`
	PageInfo collection.PageInfo `json:"pageInfo"`
}

type Usecase interface {
	ProvisionStorePlacement(
		ctx context.Context,
		req ProvisionStorePlacementRequest,
		actor map[string]string,
	) (*ProvisionStorePlacementResponse, error)
	GetTenantPlacementStatus(ctx context.Context, tenantID string) (*PlacementStatus, error)
	ReconcileTenantPlacement(
		ctx context.Context,
		tenantID string,
		actor map[string]string,
	) (*PlacementReconcileResponse, error)
	IsPlacementRouteReady(ctx context.Context, tenantID string) (bool, error)
	EnsurePlacementRoute(ctx context.Context, tenantID string) (bool, error)
	ManualUpsertConnection(
		ctx context.Context,
		tenantID string,
		req UpsertConnectionRequest,
		actor map[string]string,
	) (*UpsertConnectionResponse, error)
	DeleteConnection(
		ctx context.Context,
		tenantID string,
		infraType entity.InfraType,
		name string,
		actor map[string]string,
	) (string, error)
	GetConnection(
		ctx context.Context,
		tenantID string,
		infraType entity.InfraType,
		name string,
	) (*Connection, error)
	ListConnections(
		ctx context.Context,
		tenantID string,
		includeDeleted bool,
		query collection.Query,
	) (collection.Page[Connection], error)
	ListEvents(
		ctx context.Context,
		tenantID string,
		query collection.Query,
	) (collection.Page[ConnectionEvent], error)
	ListDatabaseClusters(
		ctx context.Context,
		query collection.Query,
	) (collection.Page[DatabaseClusterResource], error)
	UpsertDatabaseCluster(ctx context.Context, resource DatabaseClusterResource) error
	CheckDatabaseClusterHealth(ctx context.Context, name string) (*DatabaseClusterHealthCheckResponse, error)
	DeleteDatabaseCluster(ctx context.Context, name string) error
	ListKubernetesClusters(
		ctx context.Context,
		query collection.Query,
	) (collection.Page[KubernetesClusterResource], error)
	UpsertKubernetesCluster(ctx context.Context, resource KubernetesClusterResource) error
	DeleteKubernetesCluster(ctx context.Context, name string) error
	ListRuntimePools(
		ctx context.Context,
		query collection.Query,
	) (collection.Page[RuntimePoolResource], error)
	UpsertRuntimePool(ctx context.Context, resource RuntimePoolResource) error
	DeleteRuntimePool(ctx context.Context, name string) error
}
