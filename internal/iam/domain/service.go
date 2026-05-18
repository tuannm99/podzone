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
	orgs                OrganizationRepository
	platformMemberships PlatformMembershipRepository
	memberships         MembershipRepository
	invites             InviteRepository
}

func NewIAMUsecase(
	tenants TenantRepository,
	roles RoleRepository,
	policies PolicyRepository,
	groups GroupRepository,
	orgs OrganizationRepository,
	platformMemberships PlatformMembershipRepository,
	memberships MembershipRepository,
	invites InviteRepository,
) IAMUsecase {
	return &iamService{
		tenants:             tenants,
		roles:               roles,
		policies:            policies,
		groups:              groups,
		orgs:                orgs,
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

func (s *iamService) CreateOrganization(ctx context.Context, name string, slug string) (*Organization, error) {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return nil, ErrInvalidOrganizationName
	}
	if slug == "" {
		return nil, ErrInvalidOrganizationSlug
	}
	now := time.Now().UTC()
	return s.orgs.Create(ctx, Organization{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *iamService) ListOrganizations(ctx context.Context) ([]Organization, error) {
	return s.orgs.List(ctx)
}

func (s *iamService) AttachTenantToOrganization(ctx context.Context, tenantID string, orgID string) error {
	tenantID = strings.TrimSpace(tenantID)
	orgID = strings.TrimSpace(orgID)
	if tenantID == "" {
		return ErrTenantNotFound
	}
	if orgID == "" {
		return ErrOrganizationNotFound
	}
	if _, err := s.tenants.GetByID(ctx, tenantID); err != nil {
		return err
	}
	if _, err := s.orgs.GetByID(ctx, orgID); err != nil {
		return err
	}
	return s.tenants.AttachOrganization(ctx, tenantID, orgID)
}

func (s *iamService) DetachTenantFromOrganization(ctx context.Context, tenantID string) error {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return ErrTenantNotFound
	}
	return s.tenants.DetachOrganization(ctx, tenantID)
}

func (s *iamService) AttachServiceControlPolicy(ctx context.Context, orgID string, policyName string) error {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return ErrOrganizationNotFound
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.orgs.AttachServiceControlPolicy(ctx, orgID, policy.ID)
}

func (s *iamService) DetachServiceControlPolicy(ctx context.Context, orgID string, policyName string) error {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return ErrOrganizationNotFound
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.orgs.DetachServiceControlPolicy(ctx, orgID, policy.ID)
}

func (s *iamService) ListServiceControlPolicies(ctx context.Context, orgID string) ([]Policy, error) {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return nil, ErrOrganizationNotFound
	}
	return s.orgs.ListServiceControlPolicies(ctx, orgID)
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
		normalized, err := normalizePolicyStatement(statement, now)
		if err != nil {
			return nil, nil, err
		}
		statements = append(statements, normalized)
	}

	return s.policies.CreatePolicy(ctx, policy, statements)
}

func (s *iamService) CreatePolicyVersion(
	ctx context.Context,
	input CreatePolicyVersionInput,
) (*PolicyVersion, []PolicyStatement, error) {
	name := strings.TrimSpace(input.PolicyName)
	if name == "" {
		return nil, nil, ErrInvalidPolicyName
	}
	if len(input.Statements) == 0 {
		return nil, nil, fmt.Errorf("iam: at least one policy statement is required")
	}
	policy, err := s.policies.GetPolicyByName(ctx, name)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	statements := make([]PolicyStatement, 0, len(input.Statements))
	for _, statement := range input.Statements {
		normalized, normErr := normalizePolicyStatement(statement, now)
		if normErr != nil {
			return nil, nil, normErr
		}
		statements = append(statements, normalized)
	}
	return s.policies.CreatePolicyVersion(ctx, policy.ID, policy.Name, statements, input.SetAsDefault)
}

func (s *iamService) DeletePolicyVersion(ctx context.Context, name string, version string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return ErrInvalidPolicyName
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return ErrPolicyVersionNotFound
	}
	policy, err := s.policies.GetPolicyByName(ctx, name)
	if err != nil {
		return err
	}
	return s.policies.DeletePolicyVersion(ctx, policy.ID, version)
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

func (s *iamService) ListPolicyVersions(ctx context.Context, name string) ([]PolicyVersion, error) {
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return nil, err
	}
	return s.policies.ListPolicyVersions(ctx, policy.ID, policy.Name)
}

func (s *iamService) SetDefaultPolicyVersion(ctx context.Context, name string, version string) error {
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return err
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return ErrInvalidPolicyName
	}
	return s.policies.SetDefaultPolicyVersion(ctx, policy.ID, version)
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
			RoleID:            role.ID,
			Effect:            effect,
			PrincipalType:     principalType,
			PrincipalPattern:  principalPattern,
			TenantPattern:     tenantPattern,
			ExternalIDPattern: strings.TrimSpace(statement.ExternalIDPattern),
			CreatedAt:         now,
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

func (s *iamService) PutRolePermissionBoundary(ctx context.Context, roleName string, policyName string) error {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return err
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	if role.Scope != policy.Scope {
		return ErrPermissionDenied
	}
	return s.roles.PutPermissionBoundary(ctx, role.ID, policy.ID)
}

func (s *iamService) GetRolePermissionBoundary(ctx context.Context, roleName string) (*RolePermissionBoundary, error) {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return nil, err
	}
	return s.roles.GetPermissionBoundary(ctx, role.ID)
}

func (s *iamService) DeleteRolePermissionBoundary(ctx context.Context, roleName string) error {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return err
	}
	return s.roles.DeletePermissionBoundary(ctx, role.ID)
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
	if input.ServicePrincipal != "" && !validServicePrincipal(input.ServicePrincipal) {
		return nil, ErrInvalidServicePrincipal
	}
	trustStatements, err := s.roles.GetTrustPolicy(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	if !s.canAssumeRole(ctx, input.UserID, tenantID, strings.TrimSpace(input.ExternalID), strings.TrimSpace(input.ServicePrincipal), trustStatements) {
		return nil, ErrAssumeRoleDenied
	}
	now := time.Now().UTC()
	duration := time.Duration(input.DurationSeconds) * time.Second
	if duration <= 0 {
		duration = time.Hour
	}
	if duration > 12*time.Hour {
		duration = 12 * time.Hour
	}
	return &AssumedRole{
		RoleID:           role.ID,
		RoleScope:        role.Scope,
		RoleName:         role.Name,
		TenantID:         tenantID,
		ServicePrincipal: strings.TrimSpace(input.ServicePrincipal),
		SessionName:      strings.TrimSpace(input.SessionName),
		SourceIdentity:   strings.TrimSpace(input.SourceIdentity),
		SessionTags:      cloneStringMap(input.SessionTags),
		ExpiresAt:        now.Add(duration),
		CreatedAt:        now,
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
		normalized, err := normalizePolicyStatement(statement, now)
		if err != nil {
			return err
		}
		statements = append(statements, normalized)
	}
	return s.groups.PutInlinePolicy(ctx, PutGroupInlinePolicyInput{
		GroupID:     input.GroupID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Statements:  statements,
	})
}

func (s *iamService) GetGroupInlinePolicy(
	ctx context.Context,
	groupID uint64,
	name string,
) (*GroupInlinePolicy, error) {
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
			UserID:     userID,
			Action:     permission,
			Resource:   "*",
			Attributes: requestAttributesFromContext(ctx),
		}, assumedRole.RoleID, permission)
	}
	request := AccessRequest{
		UserID:     userID,
		Action:     permission,
		Resource:   "*",
		Attributes: requestAttributesFromContext(ctx),
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
	if len(statements) > 0 {
		result := explainPolicyStatements("identity", request, statements)
		if !result.Allowed {
			if result.Reason == "explicit deny matched" {
				return false, nil
			}
			goto roleEvaluation
		}
		if !s.evaluatePlatformUserBoundary(ctx, request, userID) {
			return false, nil
		}
		sessionStatements := GetSessionPolicyStatements(ctx)
		if len(sessionStatements) > 0 {
			return evaluatePolicyStatements(request, sessionStatements), nil
		}
		return true, nil
	}
roleEvaluation:
	for _, roleID := range roleIDs {
		roleStatements, err := s.policies.ListRoleStatements(ctx, roleID)
		if err != nil {
			return false, err
		}
		if len(roleStatements) > 0 {
			allowed := evaluatePolicyStatements(request, roleStatements)
			if !allowed {
				continue
			}
			if !s.evaluateRoleBoundary(ctx, request, roleID) {
				continue
			}
			if !s.evaluatePlatformUserBoundary(ctx, request, userID) {
				return false, nil
			}
			sessionStatements := GetSessionPolicyStatements(ctx)
			if len(sessionStatements) > 0 {
				return evaluatePolicyStatements(request, sessionStatements), nil
			}
			return true, nil
		}
		allowed, err := s.roles.RoleHasPermission(ctx, roleID, permission)
		if err != nil {
			return false, err
		}
		if allowed {
			if !s.evaluateRoleBoundary(ctx, request, roleID) {
				continue
			}
			if !s.evaluatePlatformUserBoundary(ctx, request, userID) {
				return false, nil
			}
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
		normalized, err := normalizePolicyStatement(statement, now)
		if err != nil {
			return err
		}
		statements = append(statements, normalized)
	}
	return s.policies.PutPlatformUserInlinePolicy(ctx, PutPlatformUserInlinePolicyInput{
		UserID:      input.UserID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Statements:  statements,
	})
}

func (s *iamService) GetPlatformUserInlinePolicy(
	ctx context.Context,
	userID uint,
	name string,
) (*UserInlinePolicy, error) {
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

func (s *iamService) PutPlatformUserPermissionBoundary(ctx context.Context, userID uint, policyName string) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	if policy.Scope != PolicyScopePlatform {
		return ErrPermissionDenied
	}
	return s.policies.PutPlatformUserPermissionBoundary(ctx, userID, policy.ID)
}

func (s *iamService) GetPlatformUserPermissionBoundary(ctx context.Context, userID uint) (*PermissionBoundary, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.policies.GetPlatformUserPermissionBoundary(ctx, userID)
}

func (s *iamService) DeletePlatformUserPermissionBoundary(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	return s.policies.DeletePlatformUserPermissionBoundary(ctx, userID)
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
			TenantID:   tenantID,
			UserID:     userID,
			Action:     permission,
			Resource:   "*",
			Attributes: requestAttributesFromContext(ctx),
		}, assumedRole.RoleID, permission)
	}
	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return false, err
	}
	membership, err := s.GetMembership(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	if membership.Status != MembershipStatusActive {
		return false, ErrInactiveMembership
	}
	request := AccessRequest{
		TenantID:   tenantID,
		OrgID:      tenant.OrgID,
		UserID:     userID,
		Action:     permission,
		Resource:   "*",
		Attributes: requestAttributesFromContext(ctx),
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
	if len(statements) > 0 {
		result := explainPolicyStatements("identity", request, statements)
		if !result.Allowed {
			if result.Reason == "explicit deny matched" {
				return false, nil
			}
			goto tenantRoleEvaluation
		}
		if !s.evaluateTenantUserBoundary(ctx, request, tenantID, userID) {
			return false, nil
		}
		if !s.evaluateOrganizationSCP(ctx, request, tenant.OrgID) {
			return false, nil
		}
		sessionStatements := GetSessionPolicyStatements(ctx)
		if len(sessionStatements) > 0 {
			return evaluatePolicyStatements(request, sessionStatements), nil
		}
		return true, nil
	}
tenantRoleEvaluation:
	roleStatements, err := s.policies.ListRoleStatements(ctx, membership.RoleID)
	if err != nil {
		return false, err
	}
	if len(roleStatements) > 0 {
		allowed := evaluatePolicyStatements(request, roleStatements)
		if allowed {
			if !s.evaluateRoleBoundary(ctx, request, membership.RoleID) {
				return false, nil
			}
			if !s.evaluateTenantUserBoundary(ctx, request, tenantID, userID) {
				return false, nil
			}
			if !s.evaluateOrganizationSCP(ctx, request, tenant.OrgID) {
				return false, nil
			}
			sessionStatements := GetSessionPolicyStatements(ctx)
			if len(sessionStatements) > 0 {
				return evaluatePolicyStatements(request, sessionStatements), nil
			}
			return true, nil
		}
	}
	allowed, err := s.roles.RoleHasPermission(ctx, membership.RoleID, permission)
	if err != nil || !allowed {
		return allowed, err
	}
	if !s.evaluateRoleBoundary(ctx, request, membership.RoleID) {
		return false, nil
	}
	if !s.evaluateTenantUserBoundary(ctx, request, tenantID, userID) {
		return false, nil
	}
	if !s.evaluateOrganizationSCP(ctx, request, tenant.OrgID) {
		return false, nil
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

func (s *iamService) SimulateAccess(ctx context.Context, input SimulateAccessInput) (*SimulateAccessResult, error) {
	scope := strings.TrimSpace(input.Scope)
	if scope == "" {
		scope = PolicyScopeTenant
	}
	request := AccessRequest{
		TenantID:   strings.TrimSpace(input.TenantID),
		UserID:     input.UserID,
		Action:     strings.TrimSpace(input.Action),
		Resource:   strings.TrimSpace(input.Resource),
		Attributes: input.Attributes,
	}
	result := SimulateAccessResult{}

	if scope == PolicyScopeTenant && request.TenantID != "" {
		tenant, err := s.tenants.GetByID(ctx, request.TenantID)
		if err != nil {
			return nil, err
		}
		if tenant != nil {
			request.OrgID = tenant.OrgID
		}
	}

	if input.UseAssumedRole && input.AssumedRole != nil {
		statements, err := s.policies.ListRoleStatements(ctx, input.AssumedRole.RoleID)
		if err != nil {
			return nil, err
		}
		identity := explainPolicyStatements("assumed_role", request, statements)
		result.Layers = append(result.Layers, decisionLayerFromResult("assumed_role", identity))
		if !identity.Allowed {
			identity.Layers = result.Layers
			return &identity, nil
		}
		if boundaryStatements, err := s.roles.GetPermissionBoundaryStatements(ctx, input.AssumedRole.RoleID); err != nil {
			return nil, err
		} else if len(boundaryStatements) > 0 {
			boundary := explainPolicyStatements("role_permission_boundary", request, boundaryStatements)
			result.Layers = append(result.Layers, decisionLayerFromResult("role_permission_boundary", boundary))
			if !boundary.Allowed {
				boundary.Layers = result.Layers
				return &boundary, nil
			}
			identity.MatchedStatements = append(identity.MatchedStatements, boundary.MatchedStatements...)
		}
		if len(input.SessionPolicy) > 0 {
			sessionResult := explainPolicyStatements("session_policy", request, input.SessionPolicy)
			result.Layers = append(result.Layers, decisionLayerFromResult("session_policy", sessionResult))
			if !sessionResult.Allowed {
				sessionResult.Layers = result.Layers
				return &sessionResult, nil
			}
			identity.MatchedStatements = append(identity.MatchedStatements, sessionResult.MatchedStatements...)
			identity.DecisionSource = sessionResult.DecisionSource
			identity.Reason = sessionResult.Reason
		}
		if scope == PolicyScopeTenant && request.OrgID != "" {
			scpResult, err := s.explainOrganizationSCP(ctx, request, request.OrgID)
			if err != nil {
				return nil, err
			}
			if scpResult != nil {
				result.Layers = append(result.Layers, decisionLayerFromResult("service_control_policy", *scpResult))
				if !scpResult.Allowed {
					scpResult.Layers = result.Layers
					return scpResult, nil
				}
				identity.MatchedStatements = append(identity.MatchedStatements, scpResult.MatchedStatements...)
			}
		}
		identity.Layers = result.Layers
		return &identity, nil
	}

	var statements []PolicyStatement
	var err error
	switch scope {
	case PolicyScopePlatform:
		statements, err = s.policies.ListPlatformUserStatements(ctx, input.UserID)
		if err != nil {
			return nil, err
		}
		groupStatements, err := s.policies.ListPlatformGroupStatements(ctx, input.UserID)
		if err != nil {
			return nil, err
		}
		statements = append(statements, groupStatements...)
		roleIDs, err := s.platformMemberships.ListRoleIDsByUser(ctx, input.UserID)
		if err != nil {
			return nil, err
		}
		for _, roleID := range roleIDs {
			roleStatements, roleErr := s.policies.ListRoleStatements(ctx, roleID)
			if roleErr != nil {
				return nil, roleErr
			}
			statements = append(statements, roleStatements...)
		}
	default:
		statements, err = s.policies.ListTenantUserStatements(ctx, request.TenantID, input.UserID)
		if err != nil {
			return nil, err
		}
		groupStatements, err := s.policies.ListTenantGroupStatements(ctx, request.TenantID, input.UserID)
		if err != nil {
			return nil, err
		}
		statements = append(statements, groupStatements...)
		membership, membershipErr := s.memberships.GetByTenantAndUser(ctx, request.TenantID, input.UserID)
		if membershipErr == nil && membership != nil {
			roleStatements, roleErr := s.policies.ListRoleStatements(ctx, membership.RoleID)
			if roleErr != nil {
				return nil, roleErr
			}
			statements = append(statements, roleStatements...)
		}
	}

	identity := explainPolicyStatements("identity", request, statements)
	result.Layers = append(result.Layers, decisionLayerFromResult("identity", identity))
	if !identity.Allowed {
		identity.Layers = result.Layers
		return &identity, nil
	}

	if scope == PolicyScopePlatform {
		boundaryStatements, boundaryErr := s.policies.GetPlatformUserPermissionBoundaryStatements(ctx, input.UserID)
		if boundaryErr != nil {
			return nil, boundaryErr
		}
		if len(boundaryStatements) > 0 {
			boundaryResult := explainPolicyStatements("permission_boundary", request, boundaryStatements)
			result.Layers = append(result.Layers, decisionLayerFromResult("permission_boundary", boundaryResult))
			if !boundaryResult.Allowed {
				boundaryResult.Layers = result.Layers
				return &boundaryResult, nil
			}
			identity.MatchedStatements = append(identity.MatchedStatements, boundaryResult.MatchedStatements...)
		}
	} else {
		boundaryStatements, boundaryErr := s.policies.GetTenantUserPermissionBoundaryStatements(ctx, request.TenantID, input.UserID)
		if boundaryErr != nil {
			return nil, boundaryErr
		}
		if len(boundaryStatements) > 0 {
			boundaryResult := explainPolicyStatements("permission_boundary", request, boundaryStatements)
			result.Layers = append(result.Layers, decisionLayerFromResult("permission_boundary", boundaryResult))
			if !boundaryResult.Allowed {
				boundaryResult.Layers = result.Layers
				return &boundaryResult, nil
			}
			identity.MatchedStatements = append(identity.MatchedStatements, boundaryResult.MatchedStatements...)
		}
		if request.OrgID != "" {
			scpResult, err := s.explainOrganizationSCP(ctx, request, request.OrgID)
			if err != nil {
				return nil, err
			}
			if scpResult != nil {
				result.Layers = append(result.Layers, decisionLayerFromResult("service_control_policy", *scpResult))
				if !scpResult.Allowed {
					scpResult.Layers = result.Layers
					return scpResult, nil
				}
				identity.MatchedStatements = append(identity.MatchedStatements, scpResult.MatchedStatements...)
			}
		}
	}

	if len(input.SessionPolicy) > 0 {
		sessionResult := explainPolicyStatements("session_policy", request, input.SessionPolicy)
		result.Layers = append(result.Layers, decisionLayerFromResult("session_policy", sessionResult))
		if !sessionResult.Allowed {
			sessionResult.Layers = result.Layers
			return &sessionResult, nil
		}
		identity.MatchedStatements = append(identity.MatchedStatements, sessionResult.MatchedStatements...)
	}

	identity.Layers = result.Layers
	return &identity, nil
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
		if !s.evaluateRoleBoundary(ctx, request, roleID) {
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
	if !s.evaluateRoleBoundary(ctx, request, roleID) {
		return false, nil
	}
	sessionStatements := GetSessionPolicyStatements(ctx)
	if len(sessionStatements) > 0 {
		return evaluatePolicyStatements(request, sessionStatements), nil
	}
	return true, nil
}

func (s *iamService) evaluatePlatformUserBoundary(ctx context.Context, request AccessRequest, userID uint) bool {
	statements, err := s.policies.GetPlatformUserPermissionBoundaryStatements(ctx, userID)
	if err != nil {
		return false
	}
	if len(statements) == 0 {
		return true
	}
	return evaluatePolicyStatements(request, statements)
}

func (s *iamService) evaluateTenantUserBoundary(ctx context.Context, request AccessRequest, tenantID string, userID uint) bool {
	statements, err := s.policies.GetTenantUserPermissionBoundaryStatements(ctx, tenantID, userID)
	if err != nil {
		return false
	}
	if len(statements) == 0 {
		return true
	}
	return evaluatePolicyStatements(request, statements)
}

func (s *iamService) evaluateRoleBoundary(ctx context.Context, request AccessRequest, roleID uint64) bool {
	statements, err := s.roles.GetPermissionBoundaryStatements(ctx, roleID)
	if err != nil {
		return false
	}
	if len(statements) == 0 {
		return true
	}
	return evaluatePolicyStatements(request, statements)
}

func (s *iamService) evaluateOrganizationSCP(ctx context.Context, request AccessRequest, orgID string) bool {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return true
	}
	statements, err := s.orgs.ListServiceControlPolicyStatements(ctx, orgID)
	if err != nil {
		return false
	}
	if len(statements) == 0 {
		return true
	}
	return evaluatePolicyStatements(request, statements)
}

func (s *iamService) explainOrganizationSCP(ctx context.Context, request AccessRequest, orgID string) (*SimulateAccessResult, error) {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return nil, nil
	}
	statements, err := s.orgs.ListServiceControlPolicyStatements(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if len(statements) == 0 {
		return nil, nil
	}
	result := explainPolicyStatements("service_control_policy", request, statements)
	return &result, nil
}

func decisionLayerFromResult(layer string, result SimulateAccessResult) SimulateDecisionLayer {
	return SimulateDecisionLayer{
		Layer:             layer,
		Allowed:           result.Allowed,
		Reason:            result.Reason,
		MatchedStatements: result.MatchedStatements,
	}
}

func (s *iamService) canAssumeRole(
	ctx context.Context,
	userID uint,
	tenantID string,
	externalID string,
	servicePrincipal string,
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
		if !matchesTrustStatement(statement, userID, tenantID, externalID, servicePrincipal, platformMemberships, tenantMembership) {
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
	externalID string,
	servicePrincipal string,
	platformMemberships []PlatformMembership,
	tenantMembership *Membership,
) bool {
	if !matchesPattern(statement.TenantPattern, tenantID) {
		return false
	}
	if statement.ExternalIDPattern != "" && !matchesPattern(statement.ExternalIDPattern, externalID) {
		return false
	}
	switch statement.PrincipalType {
	case TrustPrincipalUser:
		return matchesPattern(statement.PrincipalPattern, fmt.Sprintf("%d", userID))
	case TrustPrincipalService:
		return servicePrincipal != "" && matchesPattern(statement.PrincipalPattern, servicePrincipal)
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

func normalizePolicyConditions(conditions []PolicyCondition) []PolicyCondition {
	if len(conditions) == 0 {
		return nil
	}
	normalized := make([]PolicyCondition, 0, len(conditions))
	for _, condition := range conditions {
		operator := strings.TrimSpace(condition.Operator)
		key := strings.TrimSpace(condition.Key)
		value := strings.TrimSpace(condition.Value)
		if operator == "" || key == "" {
			continue
		}
		normalized = append(normalized, PolicyCondition{
			Operator: operator,
			Key:      key,
			Value:    value,
		})
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizePolicyStatement(statement PolicyStatement, createdAt time.Time) (PolicyStatement, error) {
	effect := strings.ToLower(strings.TrimSpace(statement.Effect))
	if effect == "" {
		effect = PolicyEffectAllow
	}
	if effect != PolicyEffectAllow && effect != PolicyEffectDeny {
		return PolicyStatement{}, ErrInvalidPolicyStatement
	}
	action := strings.TrimSpace(statement.ActionPattern)
	if action == "" || len(action) > 128 || !strings.Contains(action, ":") {
		return PolicyStatement{}, ErrInvalidPolicyStatement
	}
	resource := strings.TrimSpace(statement.ResourcePattern)
	if resource == "" {
		resource = "*"
	}
	if resource != "*" && len(resource) > 256 {
		return PolicyStatement{}, ErrInvalidPolicyStatement
	}
	conditions := normalizePolicyConditions(statement.Conditions)
	if len(conditions) > 16 {
		return PolicyStatement{}, ErrInvalidPolicyStatement
	}
	return PolicyStatement{
		Effect:          effect,
		ActionPattern:   action,
		ResourcePattern: resource,
		Conditions:      conditions,
		CreatedAt:       createdAt,
	}, nil
}

func cloneStringMap(items map[string]string) map[string]string {
	if len(items) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(items))
	for k, v := range items {
		cloned[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return cloned
}

func validServicePrincipal(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && strings.Contains(value, ".")
}

func requestAttributesFromContext(ctx context.Context) map[string]string {
	tags := GetSessionTags(ctx)
	if len(tags) == 0 {
		return nil
	}
	out := make(map[string]string, len(tags))
	for k, v := range tags {
		out["principal_tag:"+k] = v
		out["request_tag:"+k] = v
	}
	return out
}
