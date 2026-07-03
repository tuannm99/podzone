package mapper

import (
	"time"

	iamdomain "github.com/tuannm99/podzone/internal/iam/domain/entity"
	pbcommonv1 "github.com/tuannm99/podzone/pkg/api/proto/common/v1"
	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
	"github.com/tuannm99/podzone/pkg/pdauthn"
)

func ToPBTenant(t *iamdomain.Tenant) *pbiamv1.Tenant {
	if t == nil {
		return nil
	}
	return &pbiamv1.Tenant{
		Id:        t.ID,
		Slug:      t.Slug,
		Name:      t.Name,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
		OrgId:     t.OrgID,
	}
}

func ToPBOrganization(org *iamdomain.Organization) *pbiamv1.Organization {
	if org == nil {
		return nil
	}
	return &pbiamv1.Organization{
		Id:         org.ID,
		Slug:       org.Slug,
		Name:       org.Name,
		CreatedAt:  org.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  org.UpdatedAt.Format(time.RFC3339),
		RootUserId: uint64(org.RootUserID),
	}
}

func ToPBMembership(m *iamdomain.Membership) *pbiamv1.TenantMembership {
	if m == nil {
		return nil
	}
	return &pbiamv1.TenantMembership{
		TenantId:  m.TenantID,
		UserId:    uint64(m.UserID),
		RoleId:    m.RoleID,
		RoleName:  m.RoleName,
		Status:    m.Status,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

func ToPBPlatformMembership(m *iamdomain.PlatformMembership) *pbiamv1.PlatformRoleMembership {
	if m == nil {
		return nil
	}
	return &pbiamv1.PlatformRoleMembership{
		UserId:    uint64(m.UserID),
		RoleId:    m.RoleID,
		RoleName:  m.RoleName,
		Status:    m.Status,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

func ToPBPolicy(policy *iamdomain.Policy) *pbiamv1.Policy {
	if policy == nil {
		return nil
	}
	return &pbiamv1.Policy{
		Id:             policy.ID,
		Scope:          policy.Scope,
		Name:           policy.Name,
		Description:    policy.Description,
		IsSystem:       policy.IsSystem,
		DefaultVersion: policy.DefaultVersion,
		CreatedAt:      policy.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      policy.UpdatedAt.Format(time.RFC3339),
	}
}

func ToPBPolicyVersion(version *iamdomain.PolicyVersion) *pbiamv1.PolicyVersion {
	if version == nil {
		return nil
	}
	return &pbiamv1.PolicyVersion{
		Id:         version.ID,
		PolicyId:   version.PolicyID,
		PolicyName: version.PolicyName,
		Version:    version.Version,
		IsDefault:  version.IsDefault,
		CreatedAt:  version.CreatedAt.Format(time.RFC3339),
	}
}

func ToPBGroup(group *iamdomain.Group) *pbiamv1.Group {
	if group == nil {
		return nil
	}
	return &pbiamv1.Group{
		Id:          group.ID,
		Scope:       group.Scope,
		TenantId:    group.TenantID,
		Name:        group.Name,
		Description: group.Description,
		IsSystem:    group.IsSystem,
		CreatedAt:   group.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   group.UpdatedAt.Format(time.RFC3339),
	}
}

func ToPBGroupInlinePolicy(policy *iamdomain.GroupInlinePolicy) *pbiamv1.GroupInlinePolicy {
	if policy == nil {
		return nil
	}
	return &pbiamv1.GroupInlinePolicy{
		GroupId:     policy.GroupID,
		Name:        policy.Name,
		Description: policy.Description,
		Statements:  ToPBPolicyStatements(policy.Statements),
		CreatedAt:   policy.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   policy.UpdatedAt.Format(time.RFC3339),
	}
}

func ToPBUserInlinePolicy(policy *iamdomain.UserInlinePolicy) *pbiamv1.UserInlinePolicy {
	if policy == nil {
		return nil
	}
	return &pbiamv1.UserInlinePolicy{
		Scope:       policy.Scope,
		TenantId:    policy.TenantID,
		UserId:      uint64(policy.UserID),
		Name:        policy.Name,
		Description: policy.Description,
		Statements:  ToPBPolicyStatements(policy.Statements),
		CreatedAt:   policy.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   policy.UpdatedAt.Format(time.RFC3339),
	}
}

func ToPBPermissionBoundary(boundary *iamdomain.PermissionBoundary) *pbiamv1.PermissionBoundary {
	if boundary == nil {
		return nil
	}
	return &pbiamv1.PermissionBoundary{
		Scope:      boundary.Scope,
		TenantId:   boundary.TenantID,
		UserId:     uint64(boundary.UserID),
		PolicyId:   boundary.PolicyID,
		PolicyName: boundary.PolicyName,
		CreatedAt:  boundary.CreatedAt.Format(time.RFC3339),
	}
}

func ToPBRolePermissionBoundary(boundary *iamdomain.RolePermissionBoundary) *pbiamv1.RolePermissionBoundary {
	if boundary == nil {
		return nil
	}
	return &pbiamv1.RolePermissionBoundary{
		RoleId:     boundary.RoleID,
		RoleName:   boundary.RoleName,
		PolicyId:   boundary.PolicyID,
		PolicyName: boundary.PolicyName,
		CreatedAt:  boundary.CreatedAt.Format(time.RFC3339),
	}
}

func ToPBSimulateAccessResponse(result *iamdomain.SimulateAccessResult) *pbiamv1.SimulateAccessResponse {
	if result == nil {
		return nil
	}
	out := make([]*pbiamv1.SimulateMatchedStatement, 0, len(result.MatchedStatements))
	for i := range result.MatchedStatements {
		item := result.MatchedStatements[i]
		out = append(out, &pbiamv1.SimulateMatchedStatement{
			PolicyName:      item.PolicyName,
			Effect:          item.Effect,
			ActionPattern:   item.ActionPattern,
			ResourcePattern: item.ResourcePattern,
			Conditions:      ToPBPolicyConditions(item.Conditions),
			Source:          item.Source,
		})
	}
	layers := make([]*pbiamv1.SimulateDecisionLayer, 0, len(result.Layers))
	for i := range result.Layers {
		layer := result.Layers[i]
		statements := make([]*pbiamv1.SimulateMatchedStatement, 0, len(layer.MatchedStatements))
		for j := range layer.MatchedStatements {
			item := layer.MatchedStatements[j]
			statements = append(statements, &pbiamv1.SimulateMatchedStatement{
				PolicyName:      item.PolicyName,
				Effect:          item.Effect,
				ActionPattern:   item.ActionPattern,
				ResourcePattern: item.ResourcePattern,
				Conditions:      ToPBPolicyConditions(item.Conditions),
				Source:          item.Source,
			})
		}
		layers = append(layers, &pbiamv1.SimulateDecisionLayer{
			Layer:             layer.Layer,
			Allowed:           layer.Allowed,
			Reason:            layer.Reason,
			MatchedStatements: statements,
		})
	}
	return &pbiamv1.SimulateAccessResponse{
		Allowed:           result.Allowed,
		DecisionSource:    result.DecisionSource,
		Reason:            result.Reason,
		MatchedStatements: out,
		Layers:            layers,
	}
}

func ToPBPolicies(items []iamdomain.Policy) []*pbiamv1.Policy {
	out := make([]*pbiamv1.Policy, 0, len(items))
	for i := range items {
		out = append(out, ToPBPolicy(&items[i]))
	}
	return out
}

func ToPBPolicyAttachments(items []iamdomain.PolicyAttachment) []*pbiamv1.PolicyAttachment {
	out := make([]*pbiamv1.PolicyAttachment, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, &pbiamv1.PolicyAttachment{
			AttachmentType: item.AttachmentType,
			Scope:          item.Scope,
			TenantId:       item.TenantID,
			RoleId:         item.RoleID,
			RoleName:       item.RoleName,
			UserId:         uint64(item.UserID),
			GroupId:        item.GroupID,
			GroupName:      item.GroupName,
			CreatedAt:      item.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}

func ToPBRoleTrustStatements(items []iamdomain.RoleTrustStatement) []*pbiamv1.RoleTrustStatement {
	out := make([]*pbiamv1.RoleTrustStatement, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, &pbiamv1.RoleTrustStatement{
			Id:                item.ID,
			RoleId:            item.RoleID,
			Effect:            item.Effect,
			PrincipalType:     item.PrincipalType,
			PrincipalPattern:  item.PrincipalPattern,
			TenantPattern:     item.TenantPattern,
			ExternalIdPattern: item.ExternalIDPattern,
			CreatedAt:         item.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}

func ToPBPolicyStatements(items []iamdomain.PolicyStatement) []*pbcommonv1.PolicyStatement {
	out := make([]*pbcommonv1.PolicyStatement, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, &pbcommonv1.PolicyStatement{
			Id:              item.ID,
			PolicyId:        item.PolicyID,
			PolicyName:      item.PolicyName,
			Effect:          item.Effect,
			ActionPattern:   item.ActionPattern,
			ResourcePattern: item.ResourcePattern,
			Conditions:      ToPBPolicyConditions(item.Conditions),
			CreatedAt:       item.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}

func FromPBPolicyStatements(items []*pbcommonv1.PolicyStatement) []iamdomain.PolicyStatement {
	out := make([]iamdomain.PolicyStatement, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, iamdomain.PolicyStatement{
			Effect:          item.Effect,
			ActionPattern:   item.ActionPattern,
			ResourcePattern: item.ResourcePattern,
			Conditions:      FromPBPolicyConditions(item.Conditions),
		})
	}
	return out
}

func FromPBRoleTrustStatements(items []*pbiamv1.RoleTrustStatement) []iamdomain.RoleTrustStatement {
	out := make([]iamdomain.RoleTrustStatement, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, iamdomain.RoleTrustStatement{
			Effect:            item.Effect,
			PrincipalType:     item.PrincipalType,
			PrincipalPattern:  item.PrincipalPattern,
			TenantPattern:     item.TenantPattern,
			ExternalIDPattern: item.ExternalIdPattern,
		})
	}
	return out
}

func ToPBTenantInvite(invite *iamdomain.TenantInvite) *pbiamv1.TenantInvite {
	if invite == nil {
		return nil
	}
	resp := &pbiamv1.TenantInvite{
		Id:              invite.ID,
		TenantId:        invite.TenantID,
		Email:           invite.Email,
		RoleId:          invite.RoleID,
		RoleName:        invite.RoleName,
		Status:          invite.Status,
		InvitedByUserId: uint64(invite.InvitedByUserID),
		CreatedAt:       invite.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       invite.UpdatedAt.Format(time.RFC3339),
		ExpiresAt:       invite.ExpiresAt.Format(time.RFC3339),
	}
	if invite.AcceptedByUserID != nil {
		resp.AcceptedByUserId = uint64(*invite.AcceptedByUserID)
	}
	if invite.AcceptedAt != nil {
		resp.AcceptedAt = invite.AcceptedAt.Format(time.RFC3339)
	}
	if invite.RevokedAt != nil {
		resp.RevokedAt = invite.RevokedAt.Format(time.RFC3339)
	}
	return resp
}

func ToIAMSessionPolicyStatements(items []pdauthn.PolicyStatement) []iamdomain.PolicyStatement {
	out := make([]iamdomain.PolicyStatement, 0, len(items))
	for _, item := range items {
		out = append(out, iamdomain.PolicyStatement{
			Effect:          item.Effect,
			ActionPattern:   item.ActionPattern,
			ResourcePattern: item.ResourcePattern,
			Conditions:      toIAMSessionPolicyConditions(item.Conditions),
		})
	}
	return out
}

func ToPBPolicyConditions(items []iamdomain.PolicyCondition) []*pbcommonv1.PolicyCondition {
	out := make([]*pbcommonv1.PolicyCondition, 0, len(items))
	for _, item := range items {
		out = append(out, &pbcommonv1.PolicyCondition{
			Operator: item.Operator,
			Key:      item.Key,
			Value:    item.Value,
		})
	}
	return out
}

func ToPBIAMAssumedRole(item *iamdomain.AssumedRole) *pbiamv1.IAMAssumedRole {
	if item == nil {
		return nil
	}
	return &pbiamv1.IAMAssumedRole{
		RoleId:           item.RoleID,
		RoleScope:        item.RoleScope,
		RoleName:         item.RoleName,
		TenantId:         item.TenantID,
		ServicePrincipal: item.ServicePrincipal,
		SessionName:      item.SessionName,
		SourceIdentity:   item.SourceIdentity,
		SessionTags:      CloneStringMap(item.SessionTags),
		ExpiresAt:        item.ExpiresAt.Format(time.RFC3339),
	}
}

func CloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func FromPBPolicyConditions(items []*pbcommonv1.PolicyCondition) []iamdomain.PolicyCondition {
	out := make([]iamdomain.PolicyCondition, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, iamdomain.PolicyCondition{
			Operator: item.Operator,
			Key:      item.Key,
			Value:    item.Value,
		})
	}
	return out
}

func toIAMSessionPolicyConditions(items []pdauthn.PolicyCondition) []iamdomain.PolicyCondition {
	out := make([]iamdomain.PolicyCondition, 0, len(items))
	for _, item := range items {
		out = append(out, iamdomain.PolicyCondition{
			Operator: item.Operator,
			Key:      item.Key,
			Value:    item.Value,
		})
	}
	return out
}
