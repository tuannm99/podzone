package domain

import (
	"context"
	"fmt"
	"strings"
)

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
