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
