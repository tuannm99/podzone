package infrasmanager

import (
	"github.com/tuannm99/podzone/internal/onboarding/infrasmanager/entity"
	inputport "github.com/tuannm99/podzone/internal/onboarding/infrasmanager/inputport"
)

type UpsertConnectionRequest = inputport.UpsertConnectionRequest
type UpsertConnectionResponse = inputport.UpsertConnectionResponse
type ConnectionDTO = inputport.Connection
type ConnectionEventDTO = inputport.ConnectionEvent
type ListConnectionsResponse = inputport.ListConnectionsResponse
type ListEventsResponse = inputport.ListEventsResponse

func toDTO(c entity.ConnectionInfo) ConnectionDTO {
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

func toEventDTO(e entity.ConnectionEvent) ConnectionEventDTO {
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
