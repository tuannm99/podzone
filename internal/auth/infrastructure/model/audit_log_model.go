package model

import (
	"time"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
)

type AuditLog struct {
	ID           string    `db:"id"`
	ActorUserID  uint      `db:"actor_user_id"`
	Action       string    `db:"action"`
	ResourceType string    `db:"resource_type"`
	ResourceID   string    `db:"resource_id"`
	TenantID     string    `db:"tenant_id"`
	Status       string    `db:"status"`
	PayloadJSON  string    `db:"payload_json"`
	CreatedAt    time.Time `db:"created_at"`
}

func (a AuditLog) ToEntity() *entity.AuditLog {
	return &entity.AuditLog{
		ID:           a.ID,
		ActorUserID:  a.ActorUserID,
		Action:       a.Action,
		ResourceType: a.ResourceType,
		ResourceID:   a.ResourceID,
		TenantID:     a.TenantID,
		Status:       a.Status,
		PayloadJSON:  a.PayloadJSON,
		CreatedAt:    a.CreatedAt,
	}
}
