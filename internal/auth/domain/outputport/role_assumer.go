package outputport

import (
	"context"
	"time"
)

const PolicyScopeTenant = "tenant"

type AssumeRoleInput struct {
	AccessToken      string
	UserID           uint
	RoleName         string
	TenantID         string
	ExternalID       string
	SessionName      string
	SourceIdentity   string
	DurationSeconds  uint32
	ServicePrincipal string
	SessionTags      map[string]string
}

type AssumedRole struct {
	RoleID           uint64
	RoleScope        string
	RoleName         string
	TenantID         string
	ServicePrincipal string
	SessionName      string
	SourceIdentity   string
	SessionTags      map[string]string
	ExpiresAt        time.Time
}

type RoleAssumer interface {
	AssumeRole(ctx context.Context, input AssumeRoleInput) (*AssumedRole, error)
}
