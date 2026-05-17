package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type iamService struct {
	tenants             TenantRepository
	roles               RoleRepository
	policies            PolicyRepository
	groups              GroupRepository
	platformMemberships PlatformMembershipRepository
	memberships         MembershipRepository
	invites             InviteRepository
}

func NewIAMUsecase(
	tenants TenantRepository,
	roles RoleRepository,
	policies PolicyRepository,
	groups GroupRepository,
	platformMemberships PlatformMembershipRepository,
	memberships MembershipRepository,
	invites InviteRepository,
) IAMUsecase {
	return &iamService{
		tenants:             tenants,
		roles:               roles,
		policies:            policies,
		groups:              groups,
		platformMemberships: platformMemberships,
		memberships:         memberships,
		invites:             invites,
	}
}

func (s *iamService) CreateTenant(ctx context.Context, ownerUserID uint, cmd CreateTenantCmd) (*Tenant, error) {
	if ownerUserID == 0 {
		return nil, ErrInvalidUserID
	}
	name := strings.TrimSpace(cmd.Name)
	if name == "" {
		return nil, ErrInvalidTenantName
	}
	slug := strings.TrimSpace(cmd.Slug)
	if slug == "" {
		return nil, ErrInvalidTenantSlug
	}

	now := time.Now().UTC()
	tenant, err := s.tenants.Create(ctx, Tenant{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, err
	}

	role, err := s.roles.GetByName(ctx, RoleTenantOwner)
	if err != nil {
		return nil, err
	}

	if err := s.memberships.Upsert(ctx, Membership{
		TenantID:  tenant.ID,
		UserID:    ownerUserID,
		RoleID:    role.ID,
		RoleName:  role.Name,
		Status:    MembershipStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *iamService) CreatePolicy(
	ctx context.Context,
	input CreatePolicyInput,
) (*Policy, []PolicyStatement, error) {
	scope := strings.TrimSpace(input.Scope)
	if scope == "" {
		scope = PolicyScopeTenant
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, nil, ErrInvalidRoleName
	}
	if len(input.Statements) == 0 {
		return nil, nil, fmt.Errorf("iam: at least one policy statement is required")
	}

	now := time.Now().UTC()
	policy := Policy{
		Scope:       scope,
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	statements := make([]PolicyStatement, 0, len(input.Statements))
	for _, statement := range input.Statements {
		effect := strings.ToLower(strings.TrimSpace(statement.Effect))
		if effect == "" {
			effect = PolicyEffectAllow
		}
		statements = append(statements, PolicyStatement{
			Effect:          effect,
			ActionPattern:   strings.TrimSpace(statement.ActionPattern),
			ResourcePattern: strings.TrimSpace(statement.ResourcePattern),
			CreatedAt:       now,
		})
	}

	return s.policies.CreatePolicy(ctx, policy, statements)
}

func (s *iamService) ListPolicies(ctx context.Context, scope string) ([]Policy, error) {
	return s.policies.ListPolicies(ctx, strings.TrimSpace(scope))
}

func (s *iamService) GetPolicy(ctx context.Context, name string) (*Policy, []PolicyStatement, error) {
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return nil, nil, err
	}
	statements, err := s.policies.GetPolicyStatements(ctx, policy.ID)
	if err != nil {
		return nil, nil, err
	}
	return policy, statements, nil
}

func (s *iamService) DeletePolicy(ctx context.Context, name string) error {
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return err
	}
	return s.policies.DeletePolicy(ctx, policy.ID)
}

func (s *iamService) CreateGroup(ctx context.Context, input CreateGroupInput) (*Group, error) {
	scope := strings.TrimSpace(input.Scope)
	if scope == "" {
		scope = PolicyScopeTenant
	}
	now := time.Now().UTC()
	group := Group{
		Scope:       scope,
		TenantID:    strings.TrimSpace(input.TenantID),
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.groups.CreateGroup(ctx, group)
}

func (s *iamService) ListGroups(ctx context.Context, scope string, tenantID string) ([]Group, error) {
	return s.groups.ListGroups(ctx, strings.TrimSpace(scope), strings.TrimSpace(tenantID))
}

func (s *iamService) DeleteGroup(ctx context.Context, groupID uint64) error {
	if groupID == 0 {
		return ErrGroupNotFound
	}
	return s.groups.DeleteGroup(ctx, groupID)
}

func (s *iamService) AddGroupMember(ctx context.Context, groupID uint64, userID uint) error {
	if groupID == 0 {
		return ErrRoleNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	return s.groups.AddMember(ctx, groupID, userID)
}

func (s *iamService) RemoveGroupMember(ctx context.Context, groupID uint64, userID uint) error {
	if groupID == 0 {
		return ErrRoleNotFound
	}
	if userID == 0 {
		return ErrInvalidUserID
	}
	return s.groups.RemoveMember(ctx, groupID, userID)
}

func (s *iamService) ListGroupMembers(ctx context.Context, groupID uint64) ([]uint, error) {
	if groupID == 0 {
		return nil, ErrRoleNotFound
	}
	return s.groups.ListMembers(ctx, groupID)
}

func (s *iamService) AttachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error {
	if groupID == 0 {
		return ErrRoleNotFound
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.groups.AttachPolicy(ctx, groupID, policy.ID)
}

func (s *iamService) DetachGroupPolicy(ctx context.Context, groupID uint64, policyName string) error {
	if groupID == 0 {
		return ErrRoleNotFound
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.groups.DetachPolicy(ctx, groupID, policy.ID)
}

func (s *iamService) ListGroupPolicies(ctx context.Context, groupID uint64) ([]Policy, error) {
	if groupID == 0 {
		return nil, ErrRoleNotFound
	}
	return s.groups.ListPolicies(ctx, groupID)
}

func (s *iamService) CheckPlatformPermission(ctx context.Context, userID uint, permission string) (bool, error) {
	if userID == 0 {
		return false, ErrInvalidUserID
	}
	request := AccessRequest{
		UserID:   userID,
		Action:   permission,
		Resource: "*",
	}
	statements, err := s.policies.ListPlatformUserStatements(ctx, userID)
	if err != nil {
		return false, err
	}
	groupStatements, err := s.policies.ListPlatformGroupStatements(ctx, userID)
	if err != nil {
		return false, err
	}
	statements = append(statements, groupStatements...)
	roleIDs, err := s.platformMemberships.ListRoleIDsByUser(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, roleID := range roleIDs {
		roleStatements, err := s.policies.ListRoleStatements(ctx, roleID)
		if err != nil {
			return false, err
		}
		statements = append(statements, roleStatements...)
	}
	if len(statements) > 0 {
		return evaluatePolicyStatements(request, statements), nil
	}
	for _, roleID := range roleIDs {
		allowed, err := s.roles.RoleHasPermission(ctx, roleID, permission)
		if err != nil {
			return false, err
		}
		if allowed {
			return true, nil
		}
	}
	return false, nil
}

func (s *iamService) RequirePlatformPermission(ctx context.Context, userID uint, permission string) error {
	allowed, err := s.CheckPlatformPermission(ctx, userID, permission)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

func (s *iamService) AddPlatformRole(ctx context.Context, userID uint, roleName string) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return ErrInvalidRoleName
	}
	role, err := s.roles.GetByName(ctx, roleName)
	if err != nil {
		return err
	}
	return s.platformMemberships.Upsert(ctx, userID, role.ID, MembershipStatusActive)
}

func (s *iamService) ListPlatformRoles(ctx context.Context, userID uint) ([]PlatformMembership, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.platformMemberships.ListByUser(ctx, userID)
}

func (s *iamService) RemovePlatformRole(ctx context.Context, userID uint, roleName string) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	roleName = strings.TrimSpace(roleName)
	if roleName == "" {
		return ErrInvalidRoleName
	}
	role, err := s.roles.GetByName(ctx, roleName)
	if err != nil {
		return err
	}
	return s.platformMemberships.Delete(ctx, userID, role.ID)
}

func (s *iamService) AttachPlatformUserPolicy(ctx context.Context, userID uint, policyName string) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.policies.AttachPlatformUserPolicy(ctx, userID, policy.ID)
}

func (s *iamService) DetachPlatformUserPolicy(ctx context.Context, userID uint, policyName string) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.policies.DetachPlatformUserPolicy(ctx, userID, policy.ID)
}

func (s *iamService) ListPlatformUserPolicies(ctx context.Context, userID uint) ([]Policy, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.policies.ListPlatformUserPolicies(ctx, userID)
}

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

func (s *iamService) CheckPermission(
	ctx context.Context,
	tenantID string,
	userID uint,
	permission string,
) (bool, error) {
	membership, err := s.GetMembership(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	if membership.Status != MembershipStatusActive {
		return false, ErrInactiveMembership
	}
	request := AccessRequest{
		TenantID: tenantID,
		UserID:   userID,
		Action:   permission,
		Resource: "*",
	}
	statements, err := s.policies.ListTenantUserStatements(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	groupStatements, err := s.policies.ListTenantGroupStatements(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	statements = append(statements, groupStatements...)
	roleStatements, err := s.policies.ListRoleStatements(ctx, membership.RoleID)
	if err != nil {
		return false, err
	}
	statements = append(statements, roleStatements...)
	if len(statements) > 0 {
		return evaluatePolicyStatements(request, statements), nil
	}
	return s.roles.RoleHasPermission(ctx, membership.RoleID, permission)
}

func (s *iamService) RequirePermission(ctx context.Context, tenantID string, userID uint, permission string) error {
	allowed, err := s.CheckPermission(ctx, tenantID, userID, permission)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}
