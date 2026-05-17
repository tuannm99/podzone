package repository

import (
	"time"

	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
)

type tenantModel struct {
	ID        string    `db:"id"`
	Slug      string    `db:"slug"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type roleModel struct {
	ID          uint64    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	IsSystem    bool      `db:"is_system"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type policyModel struct {
	ID          uint64    `db:"id"`
	Scope       string    `db:"scope"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	IsSystem    bool      `db:"is_system"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type groupModel struct {
	ID          uint64    `db:"id"`
	Scope       string    `db:"scope"`
	TenantID    string    `db:"tenant_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	IsSystem    bool      `db:"is_system"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type policyStatementModel struct {
	ID              uint64    `db:"id"`
	PolicyID        uint64    `db:"policy_id"`
	PolicyName      string    `db:"policy_name"`
	Effect          string    `db:"effect"`
	ActionPattern   string    `db:"action_pattern"`
	ResourcePattern string    `db:"resource_pattern"`
	CreatedAt       time.Time `db:"created_at"`
}

type membershipModel struct {
	TenantID  string    `db:"tenant_id"`
	UserID    uint      `db:"user_id"`
	RoleID    uint64    `db:"role_id"`
	RoleName  string    `db:"role_name"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type platformMembershipModel struct {
	UserID    uint      `db:"user_id"`
	RoleID    uint64    `db:"role_id"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type inviteModel struct {
	ID               string     `db:"id"`
	TenantID         string     `db:"tenant_id"`
	Email            string     `db:"email"`
	RoleID           uint64     `db:"role_id"`
	RoleName         string     `db:"role_name"`
	Status           string     `db:"status"`
	InvitedByUserID  uint       `db:"invited_by_user_id"`
	AcceptedByUserID *uint      `db:"accepted_by_user_id"`
	TokenHash        string     `db:"token_hash"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
	ExpiresAt        time.Time  `db:"expires_at"`
	AcceptedAt       *time.Time `db:"accepted_at"`
	RevokedAt        *time.Time `db:"revoked_at"`
}

func (m inviteModel) toEntity() *iamdomain.TenantInvite {
	return &iamdomain.TenantInvite{
		ID:               m.ID,
		TenantID:         m.TenantID,
		Email:            m.Email,
		RoleID:           m.RoleID,
		RoleName:         m.RoleName,
		Status:           m.Status,
		InvitedByUserID:  m.InvitedByUserID,
		AcceptedByUserID: m.AcceptedByUserID,
		TokenHash:        m.TokenHash,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
		ExpiresAt:        m.ExpiresAt,
		AcceptedAt:       m.AcceptedAt,
		RevokedAt:        m.RevokedAt,
	}
}

func (m policyStatementModel) toEntity() iamdomain.PolicyStatement {
	return iamdomain.PolicyStatement{
		ID:              m.ID,
		PolicyID:        m.PolicyID,
		PolicyName:      m.PolicyName,
		Effect:          m.Effect,
		ActionPattern:   m.ActionPattern,
		ResourcePattern: m.ResourcePattern,
		CreatedAt:       m.CreatedAt,
	}
}

func (m policyModel) toEntity() iamdomain.Policy {
	return iamdomain.Policy{
		ID:          m.ID,
		Scope:       m.Scope,
		Name:        m.Name,
		Description: m.Description,
		IsSystem:    m.IsSystem,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (m groupModel) toEntity() iamdomain.Group {
	return iamdomain.Group{
		ID:          m.ID,
		Scope:       m.Scope,
		TenantID:    m.TenantID,
		Name:        m.Name,
		Description: m.Description,
		IsSystem:    m.IsSystem,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
