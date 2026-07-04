package entity

import "time"

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

type User struct {
	ID          uint
	Email       string
	Username    string
	DisplayName string
}

type Permission struct {
	ID       uint64
	Name     string
	Resource string
	Action   string
}
