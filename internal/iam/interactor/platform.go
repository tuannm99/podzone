package interactor

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func (s *interactor) CheckPlatformPermission(ctx context.Context, userID uint, permission string) (bool, error) {
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

func (s *interactor) RequirePlatformPermission(ctx context.Context, userID uint, permission string) error {
	allowed, err := s.CheckPlatformPermission(ctx, userID, permission)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

func (s *interactor) AddPlatformRole(ctx context.Context, userID uint, roleName string) error {
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

func (s *interactor) ListPlatformRoles(ctx context.Context, userID uint) ([]PlatformMembership, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.platformMemberships.ListByUser(ctx, userID)
}

func (s *interactor) RemovePlatformRole(ctx context.Context, userID uint, roleName string) error {
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

func (s *interactor) PutPlatformUserInlinePolicy(ctx context.Context, input PutPlatformUserInlinePolicyInput) error {
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

func (s *interactor) GetPlatformUserInlinePolicy(
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

func (s *interactor) ListPlatformUserInlinePolicies(ctx context.Context, userID uint) ([]UserInlinePolicy, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.policies.ListPlatformUserInlinePolicies(ctx, userID)
}

func (s *interactor) DeletePlatformUserInlinePolicy(ctx context.Context, userID uint, name string) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	if strings.TrimSpace(name) == "" {
		return ErrInvalidPolicyName
	}
	return s.policies.DeletePlatformUserInlinePolicy(ctx, userID, strings.TrimSpace(name))
}

func (s *interactor) AttachPlatformUserPolicy(ctx context.Context, userID uint, policyName string) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	if err := s.policies.AttachPlatformUserPolicy(ctx, userID, policy.ID); err != nil {
		return err
	}
	now := time.Now().UTC()
	record, err := newIAMEventOutboxRecord(now, "policy.attached", "", policy.Name, policy.Name, map[string]any{
		"user_id":          userID,
		"policy_id":        policy.ID,
		"policy_name":      policy.Name,
		"policy_scope":     policy.Scope,
		"attachment_type":  "platform_user",
		"attachment_scope": PolicyScopePlatform,
	})
	if err != nil {
		return err
	}
	return s.appendOutboxRecord(ctx, now, record)
}

func (s *interactor) DetachPlatformUserPolicy(ctx context.Context, userID uint, policyName string) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	policy, err := s.policies.GetPolicyByName(ctx, strings.TrimSpace(policyName))
	if err != nil {
		return err
	}
	return s.policies.DetachPlatformUserPolicy(ctx, userID, policy.ID)
}

func (s *interactor) ListPlatformUserPolicies(ctx context.Context, userID uint) ([]Policy, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.policies.ListPlatformUserPolicies(ctx, userID)
}

func (s *interactor) PutPlatformUserPermissionBoundary(ctx context.Context, userID uint, policyName string) error {
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

func (s *interactor) GetPlatformUserPermissionBoundary(ctx context.Context, userID uint) (*PermissionBoundary, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	return s.policies.GetPlatformUserPermissionBoundary(ctx, userID)
}

func (s *interactor) DeletePlatformUserPermissionBoundary(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	return s.policies.DeletePlatformUserPermissionBoundary(ctx, userID)
}
