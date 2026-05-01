package domain

import (
	"errors"
	"time"
)

type Tenant struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Role struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Membership struct {
	TenantID  string    `json:"tenant_id"`
	UserID    uint      `json:"user_id"`
	RoleID    uint64    `json:"role_id"`
	RoleName  string    `json:"role_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PlatformMembership struct {
	UserID    uint      `json:"user_id"`
	RoleID    uint64    `json:"role_id"`
	RoleName  string    `json:"role_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateTenantCmd struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

const (
	RolePlatformOwner = "platform_owner"
	RolePlatformAdmin = "platform_admin"

	RoleTenantOwner  = "tenant_owner"
	RoleTenantAdmin  = "tenant_admin"
	RoleTenantEditor = "tenant_editor"
	RoleTenantViewer = "tenant_viewer"

	MembershipStatusActive = "active"
)

var (
	ErrTenantNotFound     = errors.New("iam: tenant not found")
	ErrRoleNotFound       = errors.New("iam: role not found")
	ErrMembershipNotFound = errors.New("iam: membership not found")
	ErrPermissionDenied   = errors.New("iam: permission denied")
	ErrTenantSlugTaken    = errors.New("iam: tenant slug already exists")
	ErrInvalidTenantName  = errors.New("iam: tenant name is required")
	ErrInvalidTenantSlug  = errors.New("iam: tenant slug is required")
	ErrInvalidUserID      = errors.New("iam: user id is required")
	ErrInvalidRoleName    = errors.New("iam: role name is required")
	ErrInactiveMembership = errors.New("iam: membership is inactive")
)
