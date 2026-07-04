package entity

import "time"

type Role struct {
	ID          uint64    `json:"id"`
	Scope       string    `json:"scope"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Group struct {
	ID          uint64    `json:"id"`
	Scope       string    `json:"scope"`
	OrgID       string    `json:"org_id,omitempty"`
	TenantID    string    `json:"tenant_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoleTrustStatement struct {
	ID                uint64    `json:"id"`
	RoleID            uint64    `json:"role_id"`
	Effect            string    `json:"effect"`
	PrincipalType     string    `json:"principal_type"`
	PrincipalPattern  string    `json:"principal_pattern"`
	TenantPattern     string    `json:"tenant_pattern"`
	ExternalIDPattern string    `json:"external_id_pattern,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type PutRoleTrustPolicyInput struct {
	RoleName   string               `json:"role_name"`
	Statements []RoleTrustStatement `json:"statements"`
}

type CreateGroupInput struct {
	Scope       string `json:"scope"`
	OrgID       string `json:"org_id,omitempty"`
	TenantID    string `json:"tenant_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PutGroupInlinePolicyInput struct {
	GroupID     uint64            `json:"group_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Statements  []PolicyStatement `json:"statements"`
}

type GroupInlinePolicy struct {
	GroupID     uint64            `json:"group_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Statements  []PolicyStatement `json:"statements"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type PutPlatformUserInlinePolicyInput struct {
	UserID      uint              `json:"user_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Statements  []PolicyStatement `json:"statements"`
}

type PutTenantUserInlinePolicyInput struct {
	TenantID    string            `json:"tenant_id"`
	UserID      uint              `json:"user_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Statements  []PolicyStatement `json:"statements"`
}

type UserInlinePolicy struct {
	Scope       string            `json:"scope"`
	TenantID    string            `json:"tenant_id,omitempty"`
	UserID      uint              `json:"user_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Statements  []PolicyStatement `json:"statements"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type AssumeRoleInput struct {
	UserID           uint              `json:"user_id"`
	RoleName         string            `json:"role_name"`
	TenantID         string            `json:"tenant_id,omitempty"`
	ExternalID       string            `json:"external_id,omitempty"`
	ServicePrincipal string            `json:"service_principal,omitempty"`
	SessionName      string            `json:"session_name,omitempty"`
	SourceIdentity   string            `json:"source_identity,omitempty"`
	DurationSeconds  uint32            `json:"duration_seconds,omitempty"`
	SessionTags      map[string]string `json:"session_tags,omitempty"`
}

type AssumedRole struct {
	RoleID           uint64            `json:"role_id"`
	RoleScope        string            `json:"role_scope"`
	RoleName         string            `json:"role_name"`
	TenantID         string            `json:"tenant_id,omitempty"`
	ServicePrincipal string            `json:"service_principal,omitempty"`
	SessionName      string            `json:"session_name,omitempty"`
	SourceIdentity   string            `json:"source_identity,omitempty"`
	SessionTags      map[string]string `json:"session_tags,omitempty"`
	ExpiresAt        time.Time         `json:"expires_at"`
	CreatedAt        time.Time         `json:"created_at"`
}
