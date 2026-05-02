package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
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

type TenantInvite struct {
	ID              string     `json:"id"`
	TenantID        string     `json:"tenant_id"`
	Email           string     `json:"email"`
	RoleID          uint64     `json:"role_id"`
	RoleName        string     `json:"role_name"`
	Status          string     `json:"status"`
	InvitedByUserID uint       `json:"invited_by_user_id"`
	AcceptedByUserID *uint     `json:"accepted_by_user_id,omitempty"`
	TokenHash       string     `json:"-"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ExpiresAt       time.Time  `json:"expires_at"`
	AcceptedAt      *time.Time `json:"accepted_at,omitempty"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
}

const (
	RolePlatformOwner = "platform_owner"
	RolePlatformAdmin = "platform_admin"

	RoleTenantOwner  = "tenant_owner"
	RoleTenantAdmin  = "tenant_admin"
	RoleTenantEditor = "tenant_editor"
	RoleTenantViewer = "tenant_viewer"

	MembershipStatusActive = "active"

	InviteStatusPending  = "pending"
	InviteStatusAccepted = "accepted"
	InviteStatusRevoked  = "revoked"
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
	ErrInvalidInviteEmail = errors.New("iam: invite email is required")
	ErrInvalidInviteToken = errors.New("iam: invite token is required")
	ErrInactiveMembership = errors.New("iam: membership is inactive")
	ErrInviteNotFound     = errors.New("iam: invite not found")
	ErrInviteExpired      = errors.New("iam: invite expired")
	ErrInviteRevoked      = errors.New("iam: invite revoked")
	ErrInviteAccepted     = errors.New("iam: invite already accepted")
	ErrInviteEmailMismatch = errors.New("iam: invite email does not match authenticated user")
)

func NormalizeInviteEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func HashInviteToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func NewInviteToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
