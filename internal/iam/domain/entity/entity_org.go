package entity

import (
	"errors"
	"time"
)

type Organization struct {
	ID         string    `json:"id"`
	Slug       string    `json:"slug"`
	Name       string    `json:"name"`
	RootUserID uint      `json:"root_user_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type OrganizationMembership struct {
	OrgID     string    `json:"org_id"`
	UserID    uint      `json:"user_id"`
	RoleID    uint64    `json:"role_id"`
	RoleName  string    `json:"role_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ServiceControlPolicyAttachment struct {
	OrgID      string    `json:"org_id"`
	OrgName    string    `json:"org_name"`
	PolicyID   uint64    `json:"policy_id"`
	PolicyName string    `json:"policy_name"`
	CreatedAt  time.Time `json:"created_at"`
}

var (
	ErrOrganizationNotFound           = errors.New("iam: organization not found")
	ErrInvalidOrganizationName        = errors.New("iam: organization name is required")
	ErrInvalidOrganizationSlug        = errors.New("iam: organization slug is required")
	ErrOrganizationRootExists         = errors.New("iam: user already owns an organization")
	ErrOrganizationMembershipNotFound = errors.New("iam: organization membership not found")
	ErrImmutableOrganizationRoot      = errors.New("iam: organization root membership is immutable")
)
