package interactor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tuannm99/podzone/internal/iam/entity"
)

func (s *interactor) CreatePolicy(
	ctx context.Context,
	input entity.CreatePolicyInput,
) (*entity.Policy, []entity.PolicyStatement, error) {
	scope := strings.TrimSpace(input.Scope)
	if scope == "" {
		scope = entity.PolicyScopeTenant
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

	return s.policies.CreatePolicy(ctx, policy, statements)
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
	policy, err := s.policies.GetPolicyByName(ctx, name)
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
	return s.policies.CreatePolicyVersion(ctx, policy.ID, policy.Name, statements, input.SetAsDefault)
}

func (s *interactor) DeletePolicyVersion(ctx context.Context, name string, version string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return entity.ErrInvalidPolicyName
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return entity.ErrPolicyVersionNotFound
	}
	policy, err := s.policies.GetPolicyByName(ctx, name)
	if err != nil {
		return err
	}
	return s.policies.DeletePolicyVersion(ctx, policy.ID, version)
}

func (s *interactor) ListPolicies(ctx context.Context, scope string) ([]entity.Policy, error) {
	return s.policies.ListPolicies(ctx, strings.TrimSpace(scope))
}

func (s *interactor) GetPolicy(ctx context.Context, name string) (*entity.Policy, []entity.PolicyStatement, error) {
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

func (s *interactor) ListPolicyVersions(ctx context.Context, name string) ([]entity.PolicyVersion, error) {
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return nil, err
	}
	return s.policies.ListPolicyVersions(ctx, policy.ID, policy.Name)
}

func (s *interactor) SetDefaultPolicyVersion(ctx context.Context, name string, version string) error {
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return err
	}
	version = strings.TrimSpace(version)
	if version == "" {
		return entity.ErrInvalidPolicyName
	}
	return s.policies.SetDefaultPolicyVersion(ctx, policy.ID, version)
}

func (s *interactor) ListPolicyAttachments(ctx context.Context, name string) ([]entity.PolicyAttachment, error) {
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return nil, err
	}
	return s.policies.ListPolicyAttachments(ctx, policy.ID)
}

func (s *interactor) DeletePolicy(ctx context.Context, name string) error {
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return err
	}
	return s.policies.DeletePolicy(ctx, policy.ID)
}

func (s *interactor) PutRoleTrustPolicy(ctx context.Context, input entity.PutRoleTrustPolicyInput) error {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(input.RoleName))
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
	return s.roles.PutTrustPolicy(ctx, role.ID, statements)
}

func (s *interactor) GetRoleTrustPolicy(ctx context.Context, roleName string) ([]entity.RoleTrustStatement, error) {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return nil, err
	}
	return s.roles.GetTrustPolicy(ctx, role.ID)
}

func (s *interactor) DeleteRoleTrustPolicy(ctx context.Context, roleName string) error {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return err
	}
	return s.roles.DeleteTrustPolicy(ctx, role.ID)
}

func (s *interactor) PutRolePermissionBoundary(ctx context.Context, roleName string, policyName string) error {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return err
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	if role.Scope != policy.Scope {
		return entity.ErrPermissionDenied
	}
	return s.roles.PutPermissionBoundary(ctx, role.ID, policy.ID)
}

func (s *interactor) GetRolePermissionBoundary(ctx context.Context, roleName string) (*entity.RolePermissionBoundary, error) {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return nil, err
	}
	return s.roles.GetPermissionBoundary(ctx, role.ID)
}

func (s *interactor) DeleteRolePermissionBoundary(ctx context.Context, roleName string) error {
	role, err := s.roles.GetByName(ctx, strings.TrimSpace(roleName))
	if err != nil {
		return err
	}
	return s.roles.DeletePermissionBoundary(ctx, role.ID)
}
