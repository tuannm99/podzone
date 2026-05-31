package interactor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/iam/domain/entity"
)

func (s *interactor) AddMember(ctx context.Context, tenantID string, userID uint, roleName string) error {
	if strings.TrimSpace(tenantID) == "" {
		return entity.ErrTenantNotFound
	}
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return entity.ErrInvalidRoleName
	}

	if _, err := s.tenants.GetByID(ctx, tenantID); err != nil {
		return err
	}
	role, err := s.roles.GetByName(ctx, roleName)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	membership := entity.Membership{
		TenantID:  tenantID,
		UserID:    userID,
		RoleID:    role.ID,
		RoleName:  role.Name,
		Status:    entity.MembershipStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.memberships.Upsert(ctx, membership); err != nil {
		return err
	}
	record, err := newIAMEventOutboxRecord(now, "tenant.member.added", tenantID, tenantID, tenantID, map[string]any{
		"tenant_id": tenantID,
		"user_id":   userID,
		"role_name": role.Name,
		"status":    membership.Status,
	})
	if err != nil {
		return err
	}
	return s.appendOutboxRecord(ctx, now, record)
}

func (s *interactor) PutTenantUserInlinePolicy(ctx context.Context, input entity.PutTenantUserInlinePolicyInput) error {
	if strings.TrimSpace(input.TenantID) == "" {
		return entity.ErrTenantNotFound
	}
	if input.UserID == 0 {
		return entity.ErrInvalidUserID
	}
	if strings.TrimSpace(input.Name) == "" {
		return entity.ErrInvalidPolicyName
	}
	if len(input.Statements) == 0 {
		return fmt.Errorf("iam: at least one policy statement is required")
	}
	now := time.Now().UTC()
	statements := make([]entity.PolicyStatement, 0, len(input.Statements))
	for _, statement := range input.Statements {
		normalized, err := normalizePolicyStatement(statement, now)
		if err != nil {
			return err
		}
		statements = append(statements, normalized)
	}
	return s.policies.PutTenantUserInlinePolicy(ctx, entity.PutTenantUserInlinePolicyInput{
		TenantID:    strings.TrimSpace(input.TenantID),
		UserID:      input.UserID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Statements:  statements,
	})
}

func (s *interactor) GetTenantUserInlinePolicy(
	ctx context.Context,
	tenantID string,
	userID uint,
	name string,
) (*entity.UserInlinePolicy, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, entity.ErrTenantNotFound
	}
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	if strings.TrimSpace(name) == "" {
		return nil, entity.ErrInvalidPolicyName
	}
	return s.policies.GetTenantUserInlinePolicy(ctx, strings.TrimSpace(tenantID), userID, strings.TrimSpace(name))
}

func (s *interactor) ListTenantUserInlinePolicies(
	ctx context.Context,
	tenantID string,
	userID uint,
) ([]entity.UserInlinePolicy, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, entity.ErrTenantNotFound
	}
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	return s.policies.ListTenantUserInlinePolicies(ctx, strings.TrimSpace(tenantID), userID)
}

func (s *interactor) DeleteTenantUserInlinePolicy(
	ctx context.Context,
	tenantID string,
	userID uint,
	name string,
) error {
	if strings.TrimSpace(tenantID) == "" {
		return entity.ErrTenantNotFound
	}
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	if strings.TrimSpace(name) == "" {
		return entity.ErrInvalidPolicyName
	}
	return s.policies.DeleteTenantUserInlinePolicy(ctx, strings.TrimSpace(tenantID), userID, strings.TrimSpace(name))
}

func (s *interactor) AttachTenantUserPolicy(
	ctx context.Context,
	tenantID string,
	userID uint,
	policyName string,
) error {
	if strings.TrimSpace(tenantID) == "" {
		return entity.ErrTenantNotFound
	}
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	if err := s.policies.AttachTenantUserPolicy(ctx, tenantID, userID, policy.ID); err != nil {
		return err
	}
	now := time.Now().UTC()
	record, err := newIAMEventOutboxRecord(now, "policy.attached", tenantID, tenantID, tenantID, map[string]any{
		"tenant_id":        tenantID,
		"user_id":          userID,
		"policy_id":        policy.ID,
		"policy_name":      policy.Name,
		"policy_scope":     policy.Scope,
		"attachment_type":  "tenant_user",
		"attachment_scope": entity.PolicyScopeTenant,
	})
	if err != nil {
		return err
	}
	return s.appendOutboxRecord(ctx, now, record)
}

func (s *interactor) DetachTenantUserPolicy(
	ctx context.Context,
	tenantID string,
	userID uint,
	policyName string,
) error {
	if strings.TrimSpace(tenantID) == "" {
		return entity.ErrTenantNotFound
	}
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.policies.DetachTenantUserPolicy(ctx, tenantID, userID, policy.ID)
}

func (s *interactor) ListTenantUserPolicies(
	ctx context.Context,
	tenantID string,
	userID uint,
) ([]entity.Policy, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, entity.ErrTenantNotFound
	}
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	return s.policies.ListTenantUserPolicies(ctx, tenantID, userID)
}

func (s *interactor) PutTenantUserPermissionBoundary(
	ctx context.Context,
	tenantID string,
	userID uint,
	policyName string,
) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return entity.ErrTenantNotFound
	}
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	if policy.Scope != entity.PolicyScopeTenant {
		return entity.ErrPermissionDenied
	}
	return s.policies.PutTenantUserPermissionBoundary(ctx, tenantID, userID, policy.ID)
}

func (s *interactor) GetTenantUserPermissionBoundary(
	ctx context.Context,
	tenantID string,
	userID uint,
) (*entity.PermissionBoundary, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, entity.ErrTenantNotFound
	}
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	return s.policies.GetTenantUserPermissionBoundary(ctx, tenantID, userID)
}

func (s *interactor) DeleteTenantUserPermissionBoundary(
	ctx context.Context,
	tenantID string,
	userID uint,
) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return entity.ErrTenantNotFound
	}
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	return s.policies.DeleteTenantUserPermissionBoundary(ctx, tenantID, userID)
}

func (s *interactor) CreateInvite(
	ctx context.Context,
	tenantID, email, roleName string,
	invitedByUserID uint,
) (*entity.TenantInvite, string, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, "", entity.ErrTenantNotFound
	}
	if invitedByUserID == 0 {
		return nil, "", entity.ErrInvalidUserID
	}
	email = entity.NormalizeInviteEmail(email)
	if email == "" {
		return nil, "", entity.ErrInvalidInviteEmail
	}
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return nil, "", entity.ErrInvalidRoleName
	}
	if _, err := s.tenants.GetByID(ctx, tenantID); err != nil {
		return nil, "", err
	}
	role, err := s.roles.GetByName(ctx, roleName)
	if err != nil {
		return nil, "", err
	}
	rawToken, err := entity.NewInviteToken()
	if err != nil {
		return nil, "", err
	}
	now := time.Now().UTC()
	invite := entity.TenantInvite{
		ID:              uuid.NewString(),
		TenantID:        tenantID,
		Email:           email,
		RoleID:          role.ID,
		RoleName:        role.Name,
		Status:          entity.InviteStatusPending,
		InvitedByUserID: invitedByUserID,
		TokenHash:       entity.HashInviteToken(rawToken),
		CreatedAt:       now,
		UpdatedAt:       now,
		ExpiresAt:       now.Add(7 * 24 * time.Hour),
	}
	if err := s.invites.Create(ctx, invite); err != nil {
		return nil, "", err
	}
	return &invite, rawToken, nil
}

func (s *interactor) ListTenantInvites(ctx context.Context, tenantID string) ([]entity.TenantInvite, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, entity.ErrTenantNotFound
	}
	return s.invites.ListByTenant(ctx, tenantID)
}

func (s *interactor) GetInvite(ctx context.Context, inviteID string) (*entity.TenantInvite, error) {
	if strings.TrimSpace(inviteID) == "" {
		return nil, entity.ErrInviteNotFound
	}
	return s.invites.GetByID(ctx, inviteID)
}

func (s *interactor) RevokeInvite(ctx context.Context, inviteID string) error {
	invite, err := s.invites.GetByID(ctx, inviteID)
	if err != nil {
		return err
	}
	if invite.Status == entity.InviteStatusAccepted {
		return entity.ErrInviteAccepted
	}
	if invite.Status == entity.InviteStatusRevoked {
		return entity.ErrInviteRevoked
	}
	return s.invites.MarkRevoked(ctx, inviteID, time.Now().UTC())
}

func (s *interactor) AcceptInvite(
	ctx context.Context,
	inviteToken string,
	userID uint,
	email string,
) (*entity.Membership, error) {
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	email = entity.NormalizeInviteEmail(email)
	if email == "" {
		return nil, entity.ErrInvalidInviteEmail
	}
	tokenHash := entity.HashInviteToken(inviteToken)
	if tokenHash == "" {
		return nil, entity.ErrInvalidInviteToken
	}
	invite, err := s.invites.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if invite.Status == entity.InviteStatusRevoked {
		return nil, entity.ErrInviteRevoked
	}
	if invite.Status == entity.InviteStatusAccepted {
		return nil, entity.ErrInviteAccepted
	}
	now := time.Now().UTC()
	if now.After(invite.ExpiresAt) {
		return nil, entity.ErrInviteExpired
	}
	if entity.NormalizeInviteEmail(invite.Email) != email {
		return nil, entity.ErrInviteEmailMismatch
	}

	membership := entity.Membership{
		TenantID:  invite.TenantID,
		UserID:    userID,
		RoleID:    invite.RoleID,
		RoleName:  invite.RoleName,
		Status:    entity.MembershipStatusActive,
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

func (s *interactor) GetMembership(ctx context.Context, tenantID string, userID uint) (*entity.Membership, error) {
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	return s.memberships.GetByTenantAndUser(ctx, tenantID, userID)
}

func (s *interactor) ListUserTenants(ctx context.Context, userID uint) ([]entity.Membership, error) {
	if userID == 0 {
		return nil, entity.ErrInvalidUserID
	}
	return s.memberships.ListByUser(ctx, userID)
}

func (s *interactor) ListTenantMembers(ctx context.Context, tenantID string) ([]entity.Membership, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, entity.ErrTenantNotFound
	}
	return s.memberships.ListByTenant(ctx, tenantID)
}

func (s *interactor) RemoveMember(ctx context.Context, tenantID string, userID uint) error {
	if strings.TrimSpace(tenantID) == "" {
		return entity.ErrTenantNotFound
	}
	if userID == 0 {
		return entity.ErrInvalidUserID
	}
	return s.memberships.Delete(ctx, tenantID, userID)
}
