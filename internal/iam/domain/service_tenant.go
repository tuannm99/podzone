package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *iamService) AddMember(ctx context.Context, tenantID string, userID uint, roleName string) error {
	if strings.TrimSpace(tenantID) == "" {
		return ErrTenantNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return ErrInvalidRoleName
	}

	if _, err := s.tenants.GetByID(ctx, tenantID); err != nil {
		return err
	}
	role, err := s.roles.GetByName(ctx, roleName)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return s.memberships.Upsert(ctx, Membership{
		TenantID:  tenantID,
		UserID:    userID,
		RoleID:    role.ID,
		RoleName:  role.Name,
		Status:    MembershipStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *iamService) PutTenantUserInlinePolicy(ctx context.Context, input PutTenantUserInlinePolicyInput) error {
	if strings.TrimSpace(input.TenantID) == "" {
		return ErrTenantNotFound
	}
	if input.UserID == 0 {
		return ErrInvalidUserID
	}
	if strings.TrimSpace(input.Name) == "" {
		return ErrInvalidPolicyName
	}
	if len(input.Statements) == 0 {
		return fmt.Errorf("iam: at least one policy statement is required")
	}
	now := time.Now().UTC()
	statements := make([]PolicyStatement, 0, len(input.Statements))
	for _, statement := range input.Statements {
		normalized, err := normalizePolicyStatement(statement, now)
		if err != nil {
			return err
		}
		statements = append(statements, normalized)
	}
	return s.policies.PutTenantUserInlinePolicy(ctx, PutTenantUserInlinePolicyInput{
		TenantID:    strings.TrimSpace(input.TenantID),
		UserID:      input.UserID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Statements:  statements,
	})
}

func (s *iamService) GetTenantUserInlinePolicy(
	ctx context.Context,
	tenantID string,
	userID uint,
	name string,
) (*UserInlinePolicy, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrTenantNotFound
	}
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidPolicyName
	}
	return s.policies.GetTenantUserInlinePolicy(ctx, strings.TrimSpace(tenantID), userID, strings.TrimSpace(name))
}

func (s *iamService) ListTenantUserInlinePolicies(
	ctx context.Context,
	tenantID string,
	userID uint,
) ([]UserInlinePolicy, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrTenantNotFound
	}
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.policies.ListTenantUserInlinePolicies(ctx, strings.TrimSpace(tenantID), userID)
}

func (s *iamService) DeleteTenantUserInlinePolicy(
	ctx context.Context,
	tenantID string,
	userID uint,
	name string,
) error {
	if strings.TrimSpace(tenantID) == "" {
		return ErrTenantNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	if strings.TrimSpace(name) == "" {
		return ErrInvalidPolicyName
	}
	return s.policies.DeleteTenantUserInlinePolicy(ctx, strings.TrimSpace(tenantID), userID, strings.TrimSpace(name))
}

func (s *iamService) AttachTenantUserPolicy(
	ctx context.Context,
	tenantID string,
	userID uint,
	policyName string,
) error {
	if strings.TrimSpace(tenantID) == "" {
		return ErrTenantNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.policies.AttachTenantUserPolicy(ctx, tenantID, userID, policy.ID)
}

func (s *iamService) DetachTenantUserPolicy(
	ctx context.Context,
	tenantID string,
	userID uint,
	policyName string,
) error {
	if strings.TrimSpace(tenantID) == "" {
		return ErrTenantNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.policies.DetachTenantUserPolicy(ctx, tenantID, userID, policy.ID)
}

func (s *iamService) ListTenantUserPolicies(ctx context.Context, tenantID string, userID uint) ([]Policy, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrTenantNotFound
	}
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.policies.ListTenantUserPolicies(ctx, tenantID, userID)
}

func (s *iamService) PutTenantUserPermissionBoundary(
	ctx context.Context,
	tenantID string,
	userID uint,
	policyName string,
) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return ErrTenantNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	if policy.Scope != PolicyScopeTenant {
		return ErrPermissionDenied
	}
	return s.policies.PutTenantUserPermissionBoundary(ctx, tenantID, userID, policy.ID)
}

func (s *iamService) GetTenantUserPermissionBoundary(
	ctx context.Context,
	tenantID string,
	userID uint,
) (*PermissionBoundary, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrTenantNotFound
	}
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.policies.GetTenantUserPermissionBoundary(ctx, tenantID, userID)
}

func (s *iamService) DeleteTenantUserPermissionBoundary(
	ctx context.Context,
	tenantID string,
	userID uint,
) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return ErrTenantNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	return s.policies.DeleteTenantUserPermissionBoundary(ctx, tenantID, userID)
}

func (s *iamService) CreateInvite(
	ctx context.Context,
	tenantID, email, roleName string,
	invitedByUserID uint,
) (*TenantInvite, string, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, "", ErrTenantNotFound
	}
	if invitedByUserID == 0 {
		return nil, "", ErrInvalidUserID
	}
	email = NormalizeInviteEmail(email)
	if email == "" {
		return nil, "", ErrInvalidInviteEmail
	}
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return nil, "", ErrInvalidRoleName
	}
	if _, err := s.tenants.GetByID(ctx, tenantID); err != nil {
		return nil, "", err
	}
	role, err := s.roles.GetByName(ctx, roleName)
	if err != nil {
		return nil, "", err
	}
	rawToken, err := NewInviteToken()
	if err != nil {
		return nil, "", err
	}
	now := time.Now().UTC()
	invite := TenantInvite{
		ID:              uuid.NewString(),
		TenantID:        tenantID,
		Email:           email,
		RoleID:          role.ID,
		RoleName:        role.Name,
		Status:          InviteStatusPending,
		InvitedByUserID: invitedByUserID,
		TokenHash:       HashInviteToken(rawToken),
		CreatedAt:       now,
		UpdatedAt:       now,
		ExpiresAt:       now.Add(7 * 24 * time.Hour),
	}
	if err := s.invites.Create(ctx, invite); err != nil {
		return nil, "", err
	}
	return &invite, rawToken, nil
}

func (s *iamService) ListTenantInvites(ctx context.Context, tenantID string) ([]TenantInvite, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrTenantNotFound
	}
	return s.invites.ListByTenant(ctx, tenantID)
}

func (s *iamService) GetInvite(ctx context.Context, inviteID string) (*TenantInvite, error) {
	if strings.TrimSpace(inviteID) == "" {
		return nil, ErrInviteNotFound
	}
	return s.invites.GetByID(ctx, inviteID)
}

func (s *iamService) RevokeInvite(ctx context.Context, inviteID string) error {
	invite, err := s.invites.GetByID(ctx, inviteID)
	if err != nil {
		return err
	}
	if invite.Status == InviteStatusAccepted {
		return ErrInviteAccepted
	}
	if invite.Status == InviteStatusRevoked {
		return ErrInviteRevoked
	}
	return s.invites.MarkRevoked(ctx, inviteID, time.Now().UTC())
}

func (s *iamService) AcceptInvite(
	ctx context.Context,
	inviteToken string,
	userID uint,
	email string,
) (*Membership, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	email = NormalizeInviteEmail(email)
	if email == "" {
		return nil, ErrInvalidInviteEmail
	}
	tokenHash := HashInviteToken(inviteToken)
	if tokenHash == "" {
		return nil, ErrInvalidInviteToken
	}
	invite, err := s.invites.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if invite.Status == InviteStatusRevoked {
		return nil, ErrInviteRevoked
	}
	if invite.Status == InviteStatusAccepted {
		return nil, ErrInviteAccepted
	}
	now := time.Now().UTC()
	if now.After(invite.ExpiresAt) {
		return nil, ErrInviteExpired
	}
	if NormalizeInviteEmail(invite.Email) != email {
		return nil, ErrInviteEmailMismatch
	}

	membership := Membership{
		TenantID:  invite.TenantID,
		UserID:    userID,
		RoleID:    invite.RoleID,
		RoleName:  invite.RoleName,
		Status:    MembershipStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.memberships.Upsert(ctx, membership); err != nil {
		return nil, err
	}
	if err := s.invites.MarkAccepted(ctx, invite.ID, userID, now); err != nil {
		return nil, err
	}
	return &membership, nil
}

func (s *iamService) GetMembership(ctx context.Context, tenantID string, userID uint) (*Membership, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.memberships.GetByTenantAndUser(ctx, tenantID, userID)
}

func (s *iamService) ListUserTenants(ctx context.Context, userID uint) ([]Membership, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.memberships.ListByUser(ctx, userID)
}

func (s *iamService) ListTenantMembers(ctx context.Context, tenantID string) ([]Membership, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrTenantNotFound
	}
	return s.memberships.ListByTenant(ctx, tenantID)
}

func (s *iamService) RemoveMember(ctx context.Context, tenantID string, userID uint) error {
	if strings.TrimSpace(tenantID) == "" {
		return ErrTenantNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	return s.memberships.Delete(ctx, tenantID, userID)
}
