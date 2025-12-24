package infrasmanager

import (
	"time"

	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/core"
)

type UpsertConnectionRequest struct {
	InfraType core.InfraType         `json:"infra_type" binding:"required"`
	Name      string                 `json:"name"`
	Endpoint  string                 `json:"endpoint"   binding:"required"`
	SecretRef string                 `json:"secret_ref"`
	Status    string                 `json:"status"`
	Meta      map[string]string      `json:"meta"`
	Config    map[string]interface{} `json:"config"`
}

type UpsertConnectionResponse struct {
	CorrelationID string        `json:"correlation_id"`
	Connection    ConnectionDTO `json:"connection"`
	Queued        bool          `json:"queued"`
	ConsulKey     string        `json:"consul_key"`
}

type ConnectionDTO struct {
	TenantID  string                 `json:"tenant_id"`
	InfraType core.InfraType         `json:"infra_type"`
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
	Items []ConnectionDTO `json:"items"`
}

type ConnectionEventDTO struct {
	ID            string                 `json:"id"`
	CorrelationID string                 `json:"correlation_id"`
	TenantID      string                 `json:"tenant_id"`
	InfraType     core.InfraType         `json:"infra_type"`
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
	Items []ConnectionEventDTO `json:"items"`
}

func toDTO(c core.ConnectionInfo) ConnectionDTO {
	return ConnectionDTO{
		TenantID:  c.TenantID,
		InfraType: c.InfraType,
		Name:      c.Name,
		Endpoint:  c.Endpoint,
		SecretRef: c.SecretRef,
		Status:    c.Status,
		Version:   c.Version,
		Meta:      c.Meta,
		Config:    c.Config,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		DeletedAt: c.DeletedAt,
	}
}

func toEventDTO(e core.ConnectionEvent) ConnectionEventDTO {
	return ConnectionEventDTO{
		ID:            e.ID,
		CorrelationID: e.CorrelationID,
		TenantID:      e.TenantID,
		InfraType:     e.InfraType,
		Name:          e.Name,
		Action:        e.Action,
		Status:        e.Status,
		Request:       e.Request,
		Result:        e.Result,
		Error:         e.Error,
		Actor:         e.Actor,
		CreatedAt:     e.CreatedAt,
	}
}

