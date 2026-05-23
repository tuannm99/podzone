package entity

import "time"

type Tenant struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	OrgID     string    `json:"org_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

type TenantInvite struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	Email            string     `json:"email"`
	RoleID           uint64     `json:"role_id"`
	RoleName         string     `json:"role_name"`
	Status           string     `json:"status"`
	InvitedByUserID  uint       `json:"invited_by_user_id"`
	AcceptedByUserID *uint      `json:"accepted_by_user_id,omitempty"`
	TokenHash        string     `json:"-"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	AcceptedAt       *time.Time `json:"accepted_at,omitempty"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
}
