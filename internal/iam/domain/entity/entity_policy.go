package entity

import "time"

type Policy struct {
	ID             uint64    `json:"id"`
	Scope          string    `json:"scope"`
	OrgID          string    `json:"org_id,omitempty"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	IsSystem       bool      `json:"is_system"`
	DefaultVersion string    `json:"default_version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PolicyRef struct {
	Scope string `json:"scope"`
	OrgID string `json:"org_id,omitempty"`
	Name  string `json:"name"`
}

type PolicyVersion struct {
	ID         uint64    `json:"id"`
	PolicyID   uint64    `json:"policy_id"`
	PolicyName string    `json:"policy_name"`
	Version    string    `json:"version"`
	IsDefault  bool      `json:"is_default"`
	CreatedAt  time.Time `json:"created_at"`
}

type PolicyCondition struct {
	Operator string `json:"operator"`
	Key      string `json:"key"`
	Value    string `json:"value"`
}

type PolicyStatement struct {
	ID              uint64            `json:"id"`
	PolicyID        uint64            `json:"policy_id"`
	PolicyName      string            `json:"policy_name"`
	Effect          string            `json:"effect"`
	ActionPattern   string            `json:"action_pattern"`
	ResourcePattern string            `json:"resource_pattern"`
	Conditions      []PolicyCondition `json:"conditions,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

type PolicyAttachment struct {
	AttachmentType string    `json:"attachment_type"`
	Scope          string    `json:"scope,omitempty"`
	TenantID       string    `json:"tenant_id,omitempty"`
	RoleID         uint64    `json:"role_id,omitempty"`
	RoleName       string    `json:"role_name,omitempty"`
	UserID         uint      `json:"user_id,omitempty"`
	GroupID        uint64    `json:"group_id,omitempty"`
	GroupName      string    `json:"group_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreatePolicyInput struct {
	Scope       string            `json:"scope"`
	OrgID       string            `json:"org_id,omitempty"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Statements  []PolicyStatement `json:"statements"`
}

type CreatePolicyVersionInput struct {
	Scope        string            `json:"scope"`
	OrgID        string            `json:"org_id,omitempty"`
	PolicyName   string            `json:"policy_name"`
	Statements   []PolicyStatement `json:"statements"`
	SetAsDefault bool              `json:"set_as_default"`
}

type PermissionBoundary struct {
	Scope      string    `json:"scope"`
	TenantID   string    `json:"tenant_id,omitempty"`
	UserID     uint      `json:"user_id"`
	PolicyID   uint64    `json:"policy_id"`
	PolicyName string    `json:"policy_name"`
	CreatedAt  time.Time `json:"created_at"`
}

type RolePermissionBoundary struct {
	RoleID     uint64    `json:"role_id"`
	RoleName   string    `json:"role_name"`
	PolicyID   uint64    `json:"policy_id"`
	PolicyName string    `json:"policy_name"`
	CreatedAt  time.Time `json:"created_at"`
}
