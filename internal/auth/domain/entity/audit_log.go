package entity

import "time"

const (
	AuditStatusSuccess = "success"
)

type AuditLog struct {
	ID           string
	ActorUserID  uint
	Action       string
	ResourceType string
	ResourceID   string
	TenantID     string
	Status       string
	PayloadJSON  string
	CreatedAt    time.Time
}
