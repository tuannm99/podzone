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
	Scope       string    `json:"scope"`
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

type Policy struct {
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
	TenantID    string    `json:"tenant_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PolicyStatement struct {
	ID              uint64    `json:"id"`
	PolicyID        uint64    `json:"policy_id"`
	PolicyName      string    `json:"policy_name"`
	Effect          string    `json:"effect"`
	ActionPattern   string    `json:"action_pattern"`
	ResourcePattern string    `json:"resource_pattern"`
	CreatedAt       time.Time `json:"created_at"`
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

type RoleTrustStatement struct {
	ID               uint64    `json:"id"`
	RoleID           uint64    `json:"role_id"`
	Effect           string    `json:"effect"`
	PrincipalType    string    `json:"principal_type"`
	PrincipalPattern string    `json:"principal_pattern"`
	TenantPattern    string    `json:"tenant_pattern"`
	CreatedAt        time.Time `json:"created_at"`
}

type PutRoleTrustPolicyInput struct {
	RoleName   string               `json:"role_name"`
	Statements []RoleTrustStatement `json:"statements"`
}

type AssumeRoleInput struct {
	UserID   uint   `json:"user_id"`
	RoleName string `json:"role_name"`
	TenantID string `json:"tenant_id,omitempty"`
}

type AssumedRole struct {
	RoleID    uint64    `json:"role_id"`
	RoleScope string    `json:"role_scope"`
	RoleName  string    `json:"role_name"`
	TenantID  string    `json:"tenant_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreatePolicyInput struct {
	Scope       string            `json:"scope"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Statements  []PolicyStatement `json:"statements"`
}

type CreateGroupInput struct {
	Scope       string `json:"scope"`
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

type AccessRequest struct {
	TenantID string `json:"tenant_id,omitempty"`
	UserID   uint   `json:"user_id,omitempty"`
	Action   string `json:"action"`
	Resource string `json:"resource"`
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

const (
	RolePlatformOwner = "platform_owner"
	RolePlatformAdmin = "platform_admin"

	RoleTenantOwner  = "tenant_owner"
	RoleTenantAdmin  = "tenant_admin"
	RoleTenantEditor = "tenant_editor"
	RoleTenantViewer = "tenant_viewer"

	PolicyScopePlatform = "platform"
	PolicyScopeTenant   = "tenant"

	PolicyEffectAllow = "allow"
	PolicyEffectDeny  = "deny"

	TrustPrincipalUser         = "user"
	TrustPrincipalPlatformRole = "platform_role"
	TrustPrincipalTenantRole   = "tenant_role"

	MembershipStatusActive = "active"

	InviteStatusPending  = "pending"
	InviteStatusAccepted = "accepted"
	InviteStatusRevoked  = "revoked"
)

var (
	ErrTenantNotFound      = errors.New("iam: tenant not found")
	ErrRoleNotFound        = errors.New("iam: role not found")
	ErrPolicyNotFound      = errors.New("iam: policy not found")
	ErrGroupNotFound       = errors.New("iam: group not found")
	ErrMembershipNotFound  = errors.New("iam: membership not found")
	ErrPermissionDenied    = errors.New("iam: permission denied")
	ErrTenantSlugTaken     = errors.New("iam: tenant slug already exists")
	ErrInvalidTenantName   = errors.New("iam: tenant name is required")
	ErrInvalidTenantSlug   = errors.New("iam: tenant slug is required")
	ErrInvalidUserID       = errors.New("iam: user id is required")
	ErrInvalidRoleName     = errors.New("iam: role name is required")
	ErrInvalidInviteEmail  = errors.New("iam: invite email is required")
	ErrInvalidInviteToken  = errors.New("iam: invite token is required")
	ErrInactiveMembership  = errors.New("iam: membership is inactive")
	ErrInviteNotFound      = errors.New("iam: invite not found")
	ErrInviteExpired       = errors.New("iam: invite expired")
	ErrInviteRevoked       = errors.New("iam: invite revoked")
	ErrInviteAccepted      = errors.New("iam: invite already accepted")
	ErrInviteEmailMismatch = errors.New("iam: invite email does not match authenticated user")
	ErrImmutablePolicy     = errors.New("iam: managed/system policy cannot be deleted")
	ErrImmutableGroup      = errors.New("iam: system group cannot be deleted")
	ErrPolicyInUse         = errors.New("iam: policy is still attached")
	ErrInvalidPolicyName   = errors.New("iam: policy name is required")
	ErrInvalidAssumeRole   = errors.New("iam: invalid assume role target")
	ErrAssumeRoleDenied    = errors.New("iam: assume role denied")
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
