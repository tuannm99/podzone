package repository

import (
	"time"

	entity "github.com/tuannm99/podzone/internal/iam/domain/entity"
)

type tenantModel struct {
	ID        string    `db:"id"`
	Slug      string    `db:"slug"`
	Name      string    `db:"name"`
	OrgID     string    `db:"org_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type organizationModel struct {
	ID         string    `db:"id"`
	Slug       string    `db:"slug"`
	Name       string    `db:"name"`
	RootUserID uint      `db:"root_user_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type organizationMembershipModel struct {
	OrgID     string    `db:"org_id"`
	UserID    uint      `db:"user_id"`
	RoleID    uint64    `db:"role_id"`
	RoleName  string    `db:"role_name"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type roleModel struct {
	ID          uint64    `db:"id"`
	Scope       string    `db:"scope"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	IsSystem    bool      `db:"is_system"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type roleTrustStatementModel struct {
	ID                uint64    `db:"id"`
	RoleID            uint64    `db:"role_id"`
	Effect            string    `db:"effect"`
	PrincipalType     string    `db:"principal_type"`
	PrincipalPattern  string    `db:"principal_pattern"`
	TenantPattern     string    `db:"tenant_pattern"`
	ExternalIDPattern string    `db:"external_id_pattern"`
	CreatedAt         time.Time `db:"created_at"`
}

type policyModel struct {
	ID             uint64    `db:"id"`
	Scope          string    `db:"scope"`
	OrgID          string    `db:"org_id"`
	Name           string    `db:"name"`
	Description    string    `db:"description"`
	IsSystem       bool      `db:"is_system"`
	DefaultVersion string    `db:"default_version"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type policyVersionModel struct {
	ID         uint64    `db:"id"`
	PolicyID   uint64    `db:"policy_id"`
	PolicyName string    `db:"policy_name"`
	Version    string    `db:"version"`
	IsDefault  bool      `db:"is_default"`
	CreatedAt  time.Time `db:"created_at"`
}

type groupModel struct {
	ID          uint64    `db:"id"`
	Scope       string    `db:"scope"`
	OrgID       string    `db:"org_id"`
	TenantID    string    `db:"tenant_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	IsSystem    bool      `db:"is_system"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type groupInlinePolicyModel struct {
	GroupID     uint64    `db:"group_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type userInlinePolicyModel struct {
	Scope       string    `db:"scope"`
	TenantID    string    `db:"tenant_id"`
	UserID      uint      `db:"user_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
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
	ConditionsJSON  string    `db:"conditions_json"`
	CreatedAt       time.Time `db:"created_at"`
}

type policyAttachmentModel struct {
	AttachmentType string    `db:"attachment_type"`
	Scope          string    `db:"scope"`
	TenantID       string    `db:"tenant_id"`
	RoleID         uint64    `db:"role_id"`
	RoleName       string    `db:"role_name"`
	UserID         uint      `db:"user_id"`
	GroupID        uint64    `db:"group_id"`
	GroupName      string    `db:"group_name"`
	CreatedAt      time.Time `db:"created_at"`
}

type permissionBoundaryModel struct {
	Scope      string    `db:"scope"`
	TenantID   string    `db:"tenant_id"`
	UserID     uint      `db:"user_id"`
	PolicyID   uint64    `db:"policy_id"`
	PolicyName string    `db:"policy_name"`
	CreatedAt  time.Time `db:"created_at"`
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

type platformMembershipRoleModel struct {
	UserID    uint      `db:"user_id"`
	RoleID    uint64    `db:"role_id"`
	RoleName  string    `db:"role_name"`
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

func (m inviteModel) toEntity() *entity.TenantInvite {
	return &entity.TenantInvite{
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

func (m policyStatementModel) toEntity() entity.PolicyStatement {
	return entity.PolicyStatement{
		ID:              m.ID,
		PolicyID:        m.PolicyID,
		PolicyName:      m.PolicyName,
		Effect:          m.Effect,
		ActionPattern:   m.ActionPattern,
		ResourcePattern: m.ResourcePattern,
		Conditions:      parsePolicyConditionsJSON(m.ConditionsJSON),
		CreatedAt:       m.CreatedAt,
	}
}

func (m policyAttachmentModel) toEntity() entity.PolicyAttachment {
	return entity.PolicyAttachment{
		AttachmentType: m.AttachmentType,
		Scope:          m.Scope,
		TenantID:       m.TenantID,
		RoleID:         m.RoleID,
		RoleName:       m.RoleName,
		UserID:         m.UserID,
		GroupID:        m.GroupID,
		GroupName:      m.GroupName,
		CreatedAt:      m.CreatedAt,
	}
}

func (m permissionBoundaryModel) toEntity() *entity.PermissionBoundary {
	return &entity.PermissionBoundary{
		Scope:      m.Scope,
		TenantID:   m.TenantID,
		UserID:     m.UserID,
		PolicyID:   m.PolicyID,
		PolicyName: m.PolicyName,
		CreatedAt:  m.CreatedAt,
	}
}

func (m policyModel) toEntity() entity.Policy {
	return entity.Policy{
		ID:             m.ID,
		Scope:          m.Scope,
		OrgID:          m.OrgID,
		Name:           m.Name,
		Description:    m.Description,
		IsSystem:       m.IsSystem,
		DefaultVersion: m.DefaultVersion,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func (m policyVersionModel) toEntity() entity.PolicyVersion {
	return entity.PolicyVersion{
		ID:         m.ID,
		PolicyID:   m.PolicyID,
		PolicyName: m.PolicyName,
		Version:    m.Version,
		IsDefault:  m.IsDefault,
		CreatedAt:  m.CreatedAt,
	}
}

func (m groupModel) toEntity() entity.Group {
	return entity.Group{
		ID:          m.ID,
		Scope:       m.Scope,
		OrgID:       m.OrgID,
		TenantID:    m.TenantID,
		Name:        m.Name,
		Description: m.Description,
		IsSystem:    m.IsSystem,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (m groupInlinePolicyModel) toEntity(statements []entity.PolicyStatement) entity.GroupInlinePolicy {
	return entity.GroupInlinePolicy{
		GroupID:     m.GroupID,
		Name:        m.Name,
		Description: m.Description,
		Statements:  statements,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (m userInlinePolicyModel) toEntity(statements []entity.PolicyStatement) entity.UserInlinePolicy {
	return entity.UserInlinePolicy{
		Scope:       m.Scope,
		TenantID:    m.TenantID,
		UserID:      m.UserID,
		Name:        m.Name,
		Description: m.Description,
		Statements:  statements,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (m roleTrustStatementModel) toEntity() entity.RoleTrustStatement {
	return entity.RoleTrustStatement{
		ID:                m.ID,
		RoleID:            m.RoleID,
		Effect:            m.Effect,
		PrincipalType:     m.PrincipalType,
		PrincipalPattern:  m.PrincipalPattern,
		TenantPattern:     m.TenantPattern,
		ExternalIDPattern: m.ExternalIDPattern,
		CreatedAt:         m.CreatedAt,
	}
}
