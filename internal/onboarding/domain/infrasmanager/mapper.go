package infrasmanager

import (
	"github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/entity"
	inputport "github.com/tuannm99/podzone/internal/onboarding/domain/infrasmanager/inputport"
)

func toDTO(c entity.ConnectionInfo) inputport.Connection {
	return inputport.Connection{
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

func toEventDTO(e entity.ConnectionEvent) inputport.ConnectionEvent {
	return inputport.ConnectionEvent{
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
