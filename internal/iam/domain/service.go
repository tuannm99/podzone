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

func (s *iamService) ListPolicyAttachments(ctx context.Context, name string) ([]PolicyAttachment, error) {
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return nil, err
	}
	return s.policies.ListPolicyAttachments(ctx, policy.ID)
}

func (s *iamService) DeletePolicy(ctx context.Context, name string) error {
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return err
	}
	return s.policies.DeletePolicy(ctx, policy.ID)
}

func (s *iamService) PutRoleTrustPolicy(ctx context.Context, input PutRoleTrustPolicyInput) error {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(input.RoleName))
	if err != nil {
		return err
	}
	if len(input.Statements) == 0 {
		return fmt.Errorf("iam: at least one trust statement is required")
	}
	now := time.Now().UTC()
	statements := make([]RoleTrustStatement, 0, len(input.Statements))
	for _, statement := range input.Statements {
		effect := strings.ToLower(strings.TrimSpace(statement.Effect))
		if effect == "" {
			effect = PolicyEffectAllow
		}
		principalType := strings.TrimSpace(statement.PrincipalType)
		principalPattern := strings.TrimSpace(statement.PrincipalPattern)
		if principalType == "" || principalPattern == "" {
			return ErrInvalidAssumeRole
		}
		tenantPattern := strings.TrimSpace(statement.TenantPattern)
		if tenantPattern == "" {
			tenantPattern = "*"
		}
		statements = append(statements, RoleTrustStatement{
			RoleID:           role.ID,
			Effect:           effect,
			PrincipalType:    principalType,
			PrincipalPattern: principalPattern,
			TenantPattern:    tenantPattern,
			CreatedAt:        now,
		})
	}
	return s.roles.PutTrustPolicy(ctx, role.ID, statements)
}

func (s *iamService) GetRoleTrustPolicy(ctx context.Context, roleName string) ([]RoleTrustStatement, error) {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return nil, err
	}
	return s.roles.GetTrustPolicy(ctx, role.ID)
}

func (s *iamService) DeleteRoleTrustPolicy(ctx context.Context, roleName string) error {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return err
	}
	return s.roles.DeleteTrustPolicy(ctx, role.ID)
}

func (s *iamService) AssumeRole(ctx context.Context, input AssumeRoleInput) (*AssumedRole, error) {
	if input.UserID == 0 {
		return nil, ErrInvalidUserID
	}
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(input.RoleName))
	if err != nil {
		return nil, err
	}
	tenantID := strings.TrimSpace(input.TenantID)
	if role.Scope == PolicyScopeTenant && tenantID == "" {
		return nil, ErrInvalidAssumeRole
	}
	if role.Scope == PolicyScopePlatform && tenantID != "" {
		return nil, ErrInvalidAssumeRole
	}
	trustStatements, err := s.roles.GetTrustPolicy(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	if !s.canAssumeRole(ctx, input.UserID, tenantID, trustStatements) {
		return nil, ErrAssumeRoleDenied
	}
	return &AssumedRole{
		RoleID:    role.ID,
		RoleScope: role.Scope,
		RoleName:  role.Name,
		TenantID:  tenantID,
		CreatedAt: time.Now().UTC(),
	}, nil
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

func (s *iamService) PutGroupInlinePolicy(ctx context.Context, input PutGroupInlinePolicyInput) error {
	if input.GroupID == 0 {
		return ErrGroupNotFound
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
	return s.groups.PutInlinePolicy(ctx, PutGroupInlinePolicyInput{
		GroupID:     input.GroupID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Statements:  statements,
	})
}

func (s *iamService) GetGroupInlinePolicy(ctx context.Context, groupID uint64, name string) (*GroupInlinePolicy, error) {
	if groupID == 0 {
		return nil, ErrGroupNotFound
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidPolicyName
	}
	return s.groups.GetInlinePolicy(ctx, groupID, strings.TrimSpace(name))
}

func (s *iamService) ListGroupInlinePolicies(ctx context.Context, groupID uint64) ([]GroupInlinePolicy, error) {
	if groupID == 0 {
		return nil, ErrGroupNotFound
	}
	return s.groups.ListInlinePolicies(ctx, groupID)
}

func (s *iamService) DeleteGroupInlinePolicy(ctx context.Context, groupID uint64, name string) error {
	if groupID == 0 {
		return ErrGroupNotFound
	}
	if strings.TrimSpace(name) == "" {
		return ErrInvalidPolicyName
	}
	return s.groups.DeleteInlinePolicy(ctx, groupID, strings.TrimSpace(name))
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
	if assumedRole, ok := GetAssumedRole(ctx); ok {
		if assumedRole.RoleScope != PolicyScopePlatform {
			return false, nil
		}
		return s.evaluateAssumedRolePermission(ctx, AccessRequest{
			UserID:   userID,
			Action:   permission,
			Resource: "*",
		}, assumedRole.RoleID, permission)
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
		allowed := evaluatePolicyStatements(request, statements)
		if !allowed {
			return false, nil
		}
		sessionStatements := GetSessionPolicyStatements(ctx)
		if len(sessionStatements) > 0 {
			return evaluatePolicyStatements(request, sessionStatements), nil
		}
		return true, nil
	}
	for _, roleID := range roleIDs {
		allowed, err := s.roles.RoleHasPermission(ctx, roleID, permission)
		if err != nil {
			return false, err
		}
		if allowed {
			sessionStatements := GetSessionPolicyStatements(ctx)
			if len(sessionStatements) > 0 {
				return evaluatePolicyStatements(request, sessionStatements), nil
			}
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

func (s *iamService) PutPlatformUserInlinePolicy(ctx context.Context, input PutPlatformUserInlinePolicyInput) error {
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
	return s.policies.PutPlatformUserInlinePolicy(ctx, PutPlatformUserInlinePolicyInput{
		UserID:      input.UserID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Statements:  statements,
	})
}

func (s *iamService) GetPlatformUserInlinePolicy(ctx context.Context, userID uint, name string) (*UserInlinePolicy, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrInvalidPolicyName
	}
	return s.policies.GetPlatformUserInlinePolicy(ctx, userID, strings.TrimSpace(name))
}

func (s *iamService) ListPlatformUserInlinePolicies(ctx context.Context, userID uint) ([]UserInlinePolicy, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.policies.ListPlatformUserInlinePolicies(ctx, userID)
}

func (s *iamService) DeletePlatformUserInlinePolicy(ctx context.Context, userID uint, name string) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	if strings.TrimSpace(name) == "" {
		return ErrInvalidPolicyName
	}
	return s.policies.DeletePlatformUserInlinePolicy(ctx, userID, strings.TrimSpace(name))
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
	return s.policies.PutTenantUserInlinePolicy(ctx, PutTenantUserInlinePolicyInput{
		TenantID:    strings.TrimSpace(input.TenantID),
		UserID:      input.UserID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Statements:  statements,
	})
}

func (s *iamService) GetTenantUserInlinePolicy(ctx context.Context, tenantID string, userID uint, name string) (*UserInlinePolicy, error) {
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

func (s *iamService) ListTenantUserInlinePolicies(ctx context.Context, tenantID string, userID uint) ([]UserInlinePolicy, error) {
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrTenantNotFound
	}
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.policies.ListTenantUserInlinePolicies(ctx, strings.TrimSpace(tenantID), userID)
}

func (s *iamService) DeleteTenantUserInlinePolicy(ctx context.Context, tenantID string, userID uint, name string) error {
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
	if assumedRole, ok := GetAssumedRole(ctx); ok {
		if assumedRole.RoleScope == PolicyScopeTenant && assumedRole.TenantID != strings.TrimSpace(tenantID) {
			return false, nil
		}
		return s.evaluateAssumedRolePermission(ctx, AccessRequest{
			TenantID: tenantID,
			UserID:   userID,
			Action:   permission,
			Resource: "*",
		}, assumedRole.RoleID, permission)
	}
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
		allowed := evaluatePolicyStatements(request, statements)
		if !allowed {
			return false, nil
		}
		sessionStatements := GetSessionPolicyStatements(ctx)
		if len(sessionStatements) > 0 {
			return evaluatePolicyStatements(request, sessionStatements), nil
		}
		return true, nil
	}
	allowed, err := s.roles.RoleHasPermission(ctx, membership.RoleID, permission)
	if err != nil || !allowed {
		return allowed, err
	}
	sessionStatements := GetSessionPolicyStatements(ctx)
	if len(sessionStatements) > 0 {
		return evaluatePolicyStatements(request, sessionStatements), nil
	}
	return true, nil
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

func (s *iamService) evaluateAssumedRolePermission(
	ctx context.Context,
	request AccessRequest,
	roleID uint64,
	permission string,
) (bool, error) {
	statements, err := s.policies.ListRoleStatements(ctx, roleID)
	if err != nil {
		return false, err
	}
	if len(statements) > 0 {
		allowed := evaluatePolicyStatements(request, statements)
		if !allowed {
			return false, nil
		}
		sessionStatements := GetSessionPolicyStatements(ctx)
		if len(sessionStatements) > 0 {
			return evaluatePolicyStatements(request, sessionStatements), nil
		}
		return true, nil
	}
	allowed, err := s.roles.RoleHasPermission(ctx, roleID, permission)
	if err != nil || !allowed {
		return allowed, err
	}
	sessionStatements := GetSessionPolicyStatements(ctx)
	if len(sessionStatements) > 0 {
		return evaluatePolicyStatements(request, sessionStatements), nil
	}
	return true, nil
}

func (s *iamService) canAssumeRole(
	ctx context.Context,
	userID uint,
	tenantID string,
	statements []RoleTrustStatement,
) bool {
	if len(statements) == 0 {
		return false
	}

	platformMemberships, _ := s.platformMemberships.ListByUser(ctx, userID)
	var tenantMembership *Membership
	if tenantID != "" {
		tenantMembership, _ = s.memberships.GetByTenantAndUser(ctx, tenantID, userID)
	}

	allowed := false
	for _, statement := range statements {
		if !matchesTrustStatement(statement, userID, tenantID, platformMemberships, tenantMembership) {
			continue
		}
		if statement.Effect == PolicyEffectDeny {
			return false
		}
		if statement.Effect == PolicyEffectAllow {
			allowed = true
		}
	}
	return allowed
}

func matchesTrustStatement(
	statement RoleTrustStatement,
	userID uint,
	tenantID string,
	platformMemberships []PlatformMembership,
	tenantMembership *Membership,
) bool {
	if !matchesPattern(statement.TenantPattern, tenantID) {
		return false
	}
	switch statement.PrincipalType {
	case TrustPrincipalUser:
		return matchesPattern(statement.PrincipalPattern, fmt.Sprintf("%d", userID))
	case TrustPrincipalPlatformRole:
		for _, membership := range platformMemberships {
			if matchesPattern(statement.PrincipalPattern, membership.RoleName) {
				return true
			}
		}
	case TrustPrincipalTenantRole:
		if tenantMembership == nil || tenantMembership.Status != MembershipStatusActive {
			return false
		}
		return matchesPattern(statement.PrincipalPattern, tenantMembership.RoleName)
	}
	return false
}
