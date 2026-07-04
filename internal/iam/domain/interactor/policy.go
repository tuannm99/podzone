package interactor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
)

func (s *interactor) CreatePolicy(
	ctx context.Context,
	input entity.CreatePolicyInput,
) (*entity.Policy, []entity.PolicyStatement, error) {
	scope := strings.TrimSpace(input.Scope)
	if scope == "" {
		scope = entity.PolicyScopeTenant
	}
	orgID := strings.TrimSpace(input.OrgID)
	if err := validatePolicyOwner(scope, orgID); err != nil {
		return nil, nil, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, nil, entity.ErrInvalidRoleName
	}
	if len(input.Statements) == 0 {
		return nil, nil, fmt.Errorf("iam: at least one policy statement is required")
	}

	now := time.Now().UTC()
	policy := entity.Policy{
		Scope:       scope,
		OrgID:       orgID,
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	statements := make([]entity.PolicyStatement, 0, len(input.Statements))
	for _, statement := range input.Statements {
		normalized, err := normalizePolicyStatement(statement, now)
		if err != nil {
			return nil, nil, err
		}
		statements = append(statements, normalized)
	}

	return s.policyCommands.CreatePolicy(ctx, policy, statements)
}

func (s *interactor) CreatePolicyVersion(
	ctx context.Context,
	input entity.CreatePolicyVersionInput,
) (*entity.PolicyVersion, []entity.PolicyStatement, error) {
	name := strings.TrimSpace(input.PolicyName)
	if name == "" {
		return nil, nil, entity.ErrInvalidPolicyName
	}
	if len(input.Statements) == 0 {
		return nil, nil, fmt.Errorf("iam: at least one policy statement is required")
	}
	ref, err := normalizePolicyRef(entity.PolicyRef{
		Scope: input.Scope,
		OrgID: input.OrgID,
		Name:  name,
	})
	if err != nil {
		return nil, nil, err
	}
	policy, err := s.policyQueries.GetPolicy(ctx, ref)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	statements := make([]entity.PolicyStatement, 0, len(input.Statements))
	for _, statement := range input.Statements {
		normalized, normErr := normalizePolicyStatement(statement, now)
		if normErr != nil {
			return nil, nil, normErr
		}
		statements = append(statements, normalized)
	}
	return s.policyCommands.CreatePolicyVersion(ctx, policy.ID, policy.Name, statements, input.SetAsDefault)
}

func (s *interactor) DeletePolicyVersion(ctx context.Context, ref entity.PolicyRef, version string) error {
	ref, err := normalizePolicyRef(ref)
	if err != nil {
		return err
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return entity.ErrPolicyVersionNotFound
	}
	policy, err := s.policyQueries.GetPolicy(ctx, ref)
	if err != nil {
		return err
	}
	return s.policyCommands.DeletePolicyVersion(ctx, policy.ID, version)
}

func (s *interactor) ListPolicies(
	ctx context.Context,
	scope string,
	orgID string,
	query collection.Query,
) (collection.Page[entity.Policy], error) {
	scope = strings.TrimSpace(scope)
	orgID = strings.TrimSpace(orgID)
	if err := validatePolicyOwner(scope, orgID); err != nil {
		return collection.Page[entity.Policy]{}, err
	}
	return s.policyQueries.ListPolicies(ctx, scope, orgID, query.Normalize())
}

func (s *interactor) GetPolicy(
	ctx context.Context,
	ref entity.PolicyRef,
) (*entity.Policy, []entity.PolicyStatement, error) {
	ref, err := normalizePolicyRef(ref)
	if err != nil {
		return nil, nil, err
	}
	policy, err := s.policyQueries.GetPolicy(ctx, ref)
	if err != nil {
		return nil, nil, err
	}
	statements, err := s.policyQueries.GetPolicyStatements(ctx, policy.ID)
	if err != nil {
		return nil, nil, err
	}
	return policy, statements, nil
}

func (s *interactor) ListPolicyVersions(
	ctx context.Context,
	ref entity.PolicyRef,
	query collection.Query,
) (collection.Page[entity.PolicyVersion], error) {
	ref, err := normalizePolicyRef(ref)
	if err != nil {
		return collection.Page[entity.PolicyVersion]{}, err
	}
	policy, err := s.policyQueries.GetPolicy(ctx, ref)
	if err != nil {
		return collection.Page[entity.PolicyVersion]{}, err
	}
	return s.policyQueries.ListPolicyVersions(ctx, policy.ID, policy.Name, query.Normalize())
}

func (s *interactor) SetDefaultPolicyVersion(
	ctx context.Context,
	ref entity.PolicyRef,
	version string,
) error {
	ref, err := normalizePolicyRef(ref)
	if err != nil {
		return err
	}
	policy, err := s.policyQueries.GetPolicy(ctx, ref)
	if err != nil {
		return err
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return entity.ErrInvalidPolicyName
	}
	return s.policyCommands.SetDefaultPolicyVersion(ctx, policy.ID, version)
}

func (s *interactor) ListPolicyAttachments(
	ctx context.Context,
	ref entity.PolicyRef,
	query collection.Query,
) (collection.Page[entity.PolicyAttachment], error) {
	ref, err := normalizePolicyRef(ref)
	if err != nil {
		return collection.Page[entity.PolicyAttachment]{}, err
	}
	policy, err := s.policyQueries.GetPolicy(ctx, ref)
	if err != nil {
		return collection.Page[entity.PolicyAttachment]{}, err
	}
	return s.policyQueries.ListPolicyAttachments(ctx, policy.ID, query.Normalize())
}

func (s *interactor) DeletePolicy(ctx context.Context, ref entity.PolicyRef) error {
	ref, err := normalizePolicyRef(ref)
	if err != nil {
		return err
	}
	policy, err := s.policyQueries.GetPolicy(ctx, ref)
	if err != nil {
		return err
	}
	return s.policyCommands.DeletePolicy(ctx, policy.ID)
}

func (s *interactor) PutRoleTrustPolicy(ctx context.Context, input entity.PutRoleTrustPolicyInput) error {
	role, err := s.roleQueries.GetByName(ctx, strings.TrimSpace(input.RoleName))
	if err != nil {
		return err
	}
	if len(input.Statements) == 0 {
		return fmt.Errorf("iam: at least one trust statement is required")
	}
	now := time.Now().UTC()
	statements := make([]entity.RoleTrustStatement, 0, len(input.Statements))
	for _, statement := range input.Statements {
		effect := strings.ToLower(strings.TrimSpace(statement.Effect))
		if effect == "" {
			effect = entity.PolicyEffectAllow
		}
		principalType := strings.TrimSpace(statement.PrincipalType)
		principalPattern := strings.TrimSpace(statement.PrincipalPattern)
		if principalType == "" || principalPattern == "" {
			return entity.ErrInvalidAssumeRole
		}
		tenantPattern := strings.TrimSpace(statement.TenantPattern)
		if tenantPattern == "" {
			tenantPattern = "*"
		}
		statements = append(statements, entity.RoleTrustStatement{
			RoleID:            role.ID,
			Effect:            effect,
			PrincipalType:     principalType,
			PrincipalPattern:  principalPattern,
			TenantPattern:     tenantPattern,
			ExternalIDPattern: strings.TrimSpace(statement.ExternalIDPattern),
			CreatedAt:         now,
		})
	}
	return s.roleCommands.PutTrustPolicy(ctx, role.ID, statements)
}

func (s *interactor) GetRoleTrustPolicy(ctx context.Context, roleName string) ([]entity.RoleTrustStatement, error) {
	role, err := s.roleQueries.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return nil, err
	}
	return s.roleQueries.GetTrustPolicy(ctx, role.ID)
}

func (s *interactor) DeleteRoleTrustPolicy(ctx context.Context, roleName string) error {
	role, err := s.roleQueries.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return err
	}
	return s.roleCommands.DeleteTrustPolicy(ctx, role.ID)
}

func (s *interactor) PutRolePermissionBoundary(ctx context.Context, roleName string, policyName string) error {
	role, err := s.roleQueries.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return err
	}
	policy, err := s.policyQueries.GetPolicy(ctx, entity.PolicyRef{
		Scope: role.Scope,
		Name:  strings.TrimSpace(policyName),
	})
	if err != nil {
		return err
	}
	if role.Scope != policy.Scope {
		return entity.ErrPermissionDenied
	}
	return s.roleCommands.PutPermissionBoundary(ctx, role.ID, policy.ID)
}

func (s *interactor) GetRolePermissionBoundary(
	ctx context.Context,
	roleName string,
) (*entity.RolePermissionBoundary, error) {
	role, err := s.roleQueries.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return nil, err
	}
	return s.roleQueries.GetPermissionBoundary(ctx, role.ID)
}

func (s *interactor) DeleteRolePermissionBoundary(ctx context.Context, roleName string) error {
	role, err := s.roleQueries.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return err
	}
	return s.roleCommands.DeletePermissionBoundary(ctx, role.ID)
}

func normalizePolicyRef(ref entity.PolicyRef) (entity.PolicyRef, error) {
	ref.Scope = strings.TrimSpace(ref.Scope)
	if ref.Scope == "" {
		ref.Scope = entity.PolicyScopeTenant
	}
	ref.OrgID = strings.TrimSpace(ref.OrgID)
	ref.Name = strings.TrimSpace(ref.Name)
	if ref.Name == "" {
		return entity.PolicyRef{}, entity.ErrInvalidPolicyName
	}
	if err := validatePolicyOwner(ref.Scope, ref.OrgID); err != nil {
		return entity.PolicyRef{}, err
	}
	return ref, nil
}

func validatePolicyOwner(scope string, orgID string) error {
	switch strings.TrimSpace(scope) {
	case entity.PolicyScopeOrganization:
		if strings.TrimSpace(orgID) == "" {
			return entity.ErrInvalidPolicyOwner
		}
	case entity.PolicyScopePlatform, entity.PolicyScopeTenant:
		if strings.TrimSpace(orgID) != "" {
			return entity.ErrInvalidPolicyOwner
		}
	default:
		return entity.ErrInvalidPolicyOwner
	}
	return nil
}
