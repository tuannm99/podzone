package interactor

import (
	"context"
	"fmt"
	"strings"

	"github.com/tuannm99/podzone/internal/iam/domain/entity"
)

func (s *interactor) CheckPermission(
	ctx context.Context,
	tenantID string,
	userID uint,
	permission string,
) (bool, error) {
	return s.CheckPermissionForResource(ctx, tenantID, userID, permission, "*")
}

func (s *interactor) CheckPermissionForResource(
	ctx context.Context,
	tenantID string,
	userID uint,
	permission string,
	resource string,
) (bool, error) {
	resource = strings.TrimSpace(resource)
	if resource == "" {
		resource = "*"
	}
	if assumedRole, ok := entity.GetAssumedRole(ctx); ok {
		if assumedRole.RoleScope == entity.PolicyScopeTenant && assumedRole.TenantID != strings.TrimSpace(tenantID) {
			return false, nil
		}
		return s.evaluateAssumedRolePermission(ctx, entity.AccessRequest{
			TenantID:   tenantID,
			UserID:     userID,
			Action:     permission,
			Resource:   resource,
			Attributes: requestAttributesFromContext(ctx),
		}, assumedRole.RoleID, permission)
	}
	tenant, err := s.tenantQueries.GetByID(ctx, tenantID)
	if err != nil {
		return false, err
	}
	membership, err := s.GetMembership(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	if membership.Status != entity.MembershipStatusActive {
		return false, entity.ErrInactiveMembership
	}
	request := entity.AccessRequest{
		TenantID:   tenantID,
		OrgID:      tenant.OrgID,
		UserID:     userID,
		Action:     permission,
		Resource:   resource,
		Attributes: requestAttributesFromContext(ctx),
	}
	statements, err := s.policyQueries.ListTenantUserStatements(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	groupStatements, err := s.policyQueries.ListTenantGroupStatements(ctx, tenantID, userID)
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
		sessionStatements := entity.GetSessionPolicyStatements(ctx)
		if len(sessionStatements) > 0 {
			return evaluatePolicyStatements(request, sessionStatements), nil
		}
		return true, nil
	}
tenantRoleEvaluation:
	roleStatements, err := s.policyQueries.ListRoleStatements(ctx, membership.RoleID)
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
			sessionStatements := entity.GetSessionPolicyStatements(ctx)
			if len(sessionStatements) > 0 {
				return evaluatePolicyStatements(request, sessionStatements), nil
			}
			return true, nil
		}
	}
	allowed, err := s.roleQueries.RoleHasPermission(ctx, membership.RoleID, permission)
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
	sessionStatements := entity.GetSessionPolicyStatements(ctx)
	if len(sessionStatements) > 0 {
		return evaluatePolicyStatements(request, sessionStatements), nil
	}
	return true, nil
}

func (s *interactor) RequirePermission(ctx context.Context, tenantID string, userID uint, permission string) error {
	allowed, err := s.CheckPermission(ctx, tenantID, userID, permission)
	if err != nil {
		return err
	}
	if !allowed {
		return entity.ErrPermissionDenied
	}
	return nil
}

func (s *interactor) SimulateAccess(
	ctx context.Context,
	input entity.SimulateAccessInput,
) (*entity.SimulateAccessResult, error) {
	scope := strings.TrimSpace(input.Scope)
	if scope == "" {
		scope = entity.PolicyScopeTenant
	}
	request := entity.AccessRequest{
		TenantID:   strings.TrimSpace(input.TenantID),
		UserID:     input.UserID,
		Action:     strings.TrimSpace(input.Action),
		Resource:   strings.TrimSpace(input.Resource),
		Attributes: input.Attributes,
	}
	result := entity.SimulateAccessResult{}

	if scope == entity.PolicyScopeTenant && request.TenantID != "" {
		tenant, err := s.tenantQueries.GetByID(ctx, request.TenantID)
		if err != nil {
			return nil, err
		}
		if tenant != nil {
			request.OrgID = tenant.OrgID
		}
	}

	if input.UseAssumedRole && input.AssumedRole != nil {
		statements, err := s.policyQueries.ListRoleStatements(ctx, input.AssumedRole.RoleID)
		if err != nil {
			return nil, err
		}
		identity := explainPolicyStatements("assumed_role", request, statements)
		result.Layers = append(result.Layers, decisionLayerFromResult("assumed_role", identity))
		if !identity.Allowed {
			identity.Layers = result.Layers
			return &identity, nil
		}
		boundaryStatements, err := s.roleQueries.GetPermissionBoundaryStatements(ctx, input.AssumedRole.RoleID)
		if err != nil {
			return nil, err
		}
		if len(boundaryStatements) > 0 {
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
		if scope == entity.PolicyScopeTenant && request.OrgID != "" {
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

	var statements []entity.PolicyStatement
	var err error
	switch scope {
	case entity.PolicyScopePlatform:
		statements, err = s.policyQueries.ListPlatformUserStatements(ctx, input.UserID)
		if err != nil {
			return nil, err
		}
		groupStatements, err := s.policyQueries.ListPlatformGroupStatements(ctx, input.UserID)
		if err != nil {
			return nil, err
		}
		statements = append(statements, groupStatements...)
		roleIDs, err := s.platformMembershipQueries.ListRoleIDsByUser(ctx, input.UserID)
		if err != nil {
			return nil, err
		}
		for _, roleID := range roleIDs {
			roleStatements, roleErr := s.policyQueries.ListRoleStatements(ctx, roleID)
			if roleErr != nil {
				return nil, roleErr
			}
			statements = append(statements, roleStatements...)
		}
	default:
		statements, err = s.policyQueries.ListTenantUserStatements(ctx, request.TenantID, input.UserID)
		if err != nil {
			return nil, err
		}
		groupStatements, err := s.policyQueries.ListTenantGroupStatements(ctx, request.TenantID, input.UserID)
		if err != nil {
			return nil, err
		}
		statements = append(statements, groupStatements...)
		membership, membershipErr := s.membershipQueries.GetByTenantAndUser(ctx, request.TenantID, input.UserID)
		if membershipErr == nil && membership != nil {
			roleStatements, roleErr := s.policyQueries.ListRoleStatements(ctx, membership.RoleID)
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

	if scope == entity.PolicyScopePlatform {
		boundaryStatements, boundaryErr := s.policyQueries.GetPlatformUserPermissionBoundaryStatements(
			ctx,
			input.UserID,
		)
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
		boundaryStatements, boundaryErr := s.policyQueries.GetTenantUserPermissionBoundaryStatements(
			ctx,
			request.TenantID,
			input.UserID,
		)
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

func (s *interactor) evaluateAssumedRolePermission(
	ctx context.Context,
	request entity.AccessRequest,
	roleID uint64,
	permission string,
) (bool, error) {
	statements, err := s.policyQueries.ListRoleStatements(ctx, roleID)
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
		sessionStatements := entity.GetSessionPolicyStatements(ctx)
		if len(sessionStatements) > 0 {
			return evaluatePolicyStatements(request, sessionStatements), nil
		}
		return true, nil
	}
	allowed, err := s.roleQueries.RoleHasPermission(ctx, roleID, permission)
	if err != nil || !allowed {
		return allowed, err
	}
	if !s.evaluateRoleBoundary(ctx, request, roleID) {
		return false, nil
	}
	sessionStatements := entity.GetSessionPolicyStatements(ctx)
	if len(sessionStatements) > 0 {
		return evaluatePolicyStatements(request, sessionStatements), nil
	}
	return true, nil
}

func (s *interactor) evaluatePlatformUserBoundary(ctx context.Context, request entity.AccessRequest, userID uint) bool {
	statements, err := s.policyQueries.GetPlatformUserPermissionBoundaryStatements(ctx, userID)
	if err != nil {
		return false
	}
	if len(statements) == 0 {
		return true
	}
	return evaluatePolicyStatements(request, statements)
}

func (s *interactor) evaluateTenantUserBoundary(
	ctx context.Context,
	request entity.AccessRequest,
	tenantID string,
	userID uint,
) bool {
	statements, err := s.policyQueries.GetTenantUserPermissionBoundaryStatements(ctx, tenantID, userID)
	if err != nil {
		return false
	}
	if len(statements) == 0 {
		return true
	}
	return evaluatePolicyStatements(request, statements)
}

func (s *interactor) evaluateRoleBoundary(ctx context.Context, request entity.AccessRequest, roleID uint64) bool {
	statements, err := s.roleQueries.GetPermissionBoundaryStatements(ctx, roleID)
	if err != nil {
		return false
	}
	if len(statements) == 0 {
		return true
	}
	return evaluatePolicyStatements(request, statements)
}

func (s *interactor) evaluateOrganizationSCP(ctx context.Context, request entity.AccessRequest, orgID string) bool {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return true
	}
	statements, err := s.orgQueries.ListServiceControlPolicyStatements(ctx, orgID)
	if err != nil {
		return false
	}
	if len(statements) == 0 {
		return true
	}
	return evaluatePolicyStatements(request, statements)
}

func (s *interactor) explainOrganizationSCP(
	ctx context.Context,
	request entity.AccessRequest,
	orgID string,
) (*entity.SimulateAccessResult, error) {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return nil, nil
	}
	statements, err := s.orgQueries.ListServiceControlPolicyStatements(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if len(statements) == 0 {
		return nil, nil
	}
	result := explainPolicyStatements("service_control_policy", request, statements)
	return &result, nil
}

func decisionLayerFromResult(layer string, result entity.SimulateAccessResult) entity.SimulateDecisionLayer {
	return entity.SimulateDecisionLayer{
		Layer:             layer,
		Allowed:           result.Allowed,
		Reason:            result.Reason,
		MatchedStatements: result.MatchedStatements,
	}
}

func (s *interactor) canAssumeRole(
	ctx context.Context,
	userID uint,
	tenantID string,
	externalID string,
	servicePrincipal string,
	statements []entity.RoleTrustStatement,
) bool {
	if len(statements) == 0 {
		return false
	}

	platformMemberships, _ := s.platformMembershipQueries.ListByUser(ctx, userID)
	var tenantMembership *entity.Membership
	if tenantID != "" {
		tenantMembership, _ = s.membershipQueries.GetByTenantAndUser(ctx, tenantID, userID)
	}

	allowed := false
	for _, statement := range statements {
		if !matchesTrustStatement(
			statement,
			userID,
			tenantID,
			externalID,
			servicePrincipal,
			platformMemberships,
			tenantMembership,
		) {
			continue
		}
		if statement.Effect == entity.PolicyEffectDeny {
			return false
		}
		if statement.Effect == entity.PolicyEffectAllow {
			allowed = true
		}
	}
	return allowed
}

func matchesTrustStatement(
	statement entity.RoleTrustStatement,
	userID uint,
	tenantID string,
	externalID string,
	servicePrincipal string,
	platformMemberships []entity.PlatformMembership,
	tenantMembership *entity.Membership,
) bool {
	if !matchesPattern(statement.TenantPattern, tenantID) {
		return false
	}
	if statement.ExternalIDPattern != "" && !matchesPattern(statement.ExternalIDPattern, externalID) {
		return false
	}
	switch statement.PrincipalType {
	case entity.TrustPrincipalUser:
		return matchesPattern(statement.PrincipalPattern, fmt.Sprintf("%d", userID))
	case entity.TrustPrincipalService:
		return servicePrincipal != "" && matchesPattern(statement.PrincipalPattern, servicePrincipal)
	case entity.TrustPrincipalPlatformRole:
		for _, membership := range platformMemberships {
			if matchesPattern(statement.PrincipalPattern, membership.RoleName) {
				return true
			}
		}
	case entity.TrustPrincipalTenantRole:
		if tenantMembership == nil || tenantMembership.Status != entity.MembershipStatusActive {
			return false
		}
		return matchesPattern(statement.PrincipalPattern, tenantMembership.RoleName)
	}
	return false
}
