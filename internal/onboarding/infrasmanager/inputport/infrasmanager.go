package inputport

import (
	"time"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/entity"
)

type UpsertConnectionRequest struct {
	InfraType entity.InfraType         `json:"infra_type" binding:"required"`
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
	ConsulKey     string     `json:"consul_key"`
}

type Connection struct {
	TenantID  string                 `json:"tenant_id"`
	InfraType entity.InfraType         `json:"infra_type"`
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
	Items []Connection `json:"items"`
}

type ConnectionEvent struct {
	ID            string                 `json:"id"`
	CorrelationID string                 `json:"correlation_id"`
	TenantID      string                 `json:"tenant_id"`
	InfraType     entity.InfraType         `json:"infra_type"`
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
	Items []ConnectionEvent `json:"items"`
}

type Usecase interface {
	ManualUpsertConnection(
		tenantID string,
		req UpsertConnectionRequest,
		actor map[string]string,
	) (*UpsertConnectionResponse, error)
	DeleteConnection(
		tenantID string,
		infraType entity.InfraType,
		name string,
		actor map[string]string,
	) (string, error)
	GetConnection(tenantID string, infraType entity.InfraType, name string) (*Connection, error)
	ListConnections(
		tenantID string,
		infraType entity.InfraType,
		includeDeleted bool,
		limit, offset int,
	) ([]Connection, error)
	ListEvents(
		tenantID string,
		infraType entity.InfraType,
		name string,
		correlationID string,
		limit, offset int,
	) ([]ConnectionEvent, error)
}
