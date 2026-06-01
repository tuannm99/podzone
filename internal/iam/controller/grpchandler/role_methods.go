package grpchandler

import (
	"context"
	"fmt"
	"strings"

	iammapper "github.com/tuannm99/podzone/internal/iam/controller/mapper"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
)

func (s *IAMCommandServer) AssumeRole(
	ctx context.Context,
	req *pbiamv1.IAMAssumeRoleRequest,
) (*pbiamv1.IAMAssumeRoleResponse, error) {
	claims, err := s.claimsFromAccessToken(req.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if claims.UserID == 0 {
		return nil, status.Error(codes.Unauthenticated, "access token missing user_id")
	}

	assumedRole, err := s.commands.AssumeRole(ctx, iamdomain.AssumeRoleInput{
		UserID:           claims.UserID,
		RoleName:         req.RoleName,
		TenantID:         req.TenantId,
		ExternalID:       req.ExternalId,
		ServicePrincipal: strings.TrimSpace(req.ServicePrincipal),
		SessionName:      req.SessionName,
		SourceIdentity:   req.SourceIdentity,
		DurationSeconds:  req.DurationSeconds,
		SessionTags:      iammapper.CloneStringMap(req.SessionTags),
	})
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.IAMAssumeRoleResponse{
		AssumedRole: iammapper.ToPBIAMAssumedRole(assumedRole),
	}, nil
}

func (s *IAMCommandServer) AddPlatformRole(
	ctx context.Context,
	req *pbiamv1.AddPlatformRoleRequest,
) (*pbiamv1.AddPlatformRoleResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.AddPlatformRole(ctx, targetUserID, req.RoleName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"platform.role.added",
		"platform_role_membership",
		fmt.Sprintf("%s:%d", req.RoleName, req.TargetUserId),
		"",
		map[string]any{
			"target_user_id": targetUserID,
			"role_name":      req.RoleName,
		},
	)
	return &pbiamv1.AddPlatformRoleResponse{}, nil
}

func (s *IAMCommandServer) PutRoleTrustPolicy(
	ctx context.Context,
	req *pbiamv1.PutRoleTrustPolicyRequest,
) (*pbiamv1.PutRoleTrustPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.PutRoleTrustPolicy(ctx, iamdomain.PutRoleTrustPolicyInput{
		RoleName:   req.RoleName,
		Statements: iammapper.FromPBRoleTrustStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.PutRoleTrustPolicyResponse{}, nil
}

func (s *IAMCommandServer) DeleteRoleTrustPolicy(
	ctx context.Context,
	req *pbiamv1.DeleteRoleTrustPolicyRequest,
) (*pbiamv1.DeleteRoleTrustPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DeleteRoleTrustPolicy(ctx, req.RoleName); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.DeleteRoleTrustPolicyResponse{}, nil
}

func (s *IAMCommandServer) RemovePlatformRole(
	ctx context.Context,
	req *pbiamv1.RemovePlatformRoleRequest,
) (*pbiamv1.RemovePlatformRoleResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.RemovePlatformRole(ctx, targetUserID, req.RoleName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"platform.role.removed",
		"platform_role_membership",
		fmt.Sprintf("%s:%d", req.RoleName, req.TargetUserId),
		"",
		map[string]any{
			"target_user_id": targetUserID,
			"role_name":      req.RoleName,
		},
	)
	return &pbiamv1.RemovePlatformRoleResponse{}, nil
}

func (s *IAMCommandServer) PutRolePermissionBoundary(
	ctx context.Context,
	req *pbiamv1.PutRolePermissionBoundaryRequest,
) (*pbiamv1.PutRolePermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.PutRolePermissionBoundary(ctx, req.RoleName, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.role_boundary.put",
		"iam_role_permission_boundary",
		req.RoleName,
		"",
		map[string]any{
			"policy_name": req.PolicyName,
		},
	)
	return &pbiamv1.PutRolePermissionBoundaryResponse{}, nil
}

func (s *IAMCommandServer) DeleteRolePermissionBoundary(
	ctx context.Context,
	req *pbiamv1.DeleteRolePermissionBoundaryRequest,
) (*pbiamv1.DeleteRolePermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DeleteRolePermissionBoundary(ctx, req.RoleName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "iam.role_boundary.deleted", "iam_role_permission_boundary", req.RoleName, "", nil)
	return &pbiamv1.DeleteRolePermissionBoundaryResponse{}, nil
}

func (s *IAMQueryServer) ListPlatformRoles(
	ctx context.Context,
	req *pbiamv1.ListPlatformRolesRequest,
) (*pbiamv1.ListPlatformRolesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.queries.ListPlatformRoles(ctx, targetUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbiamv1.PlatformRoleMembership, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, iammapper.ToPBPlatformMembership(&item))
	}
	return &pbiamv1.ListPlatformRolesResponse{Memberships: out}, nil
}

func (s *IAMQueryServer) GetRoleTrustPolicy(
	ctx context.Context,
	req *pbiamv1.GetRoleTrustPolicyRequest,
) (*pbiamv1.GetRoleTrustPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.queries.GetRoleTrustPolicy(ctx, req.RoleName)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.GetRoleTrustPolicyResponse{Statements: iammapper.ToPBRoleTrustStatements(items)}, nil
}

func (s *IAMQueryServer) GetRolePermissionBoundary(
	ctx context.Context,
	req *pbiamv1.GetRolePermissionBoundaryRequest,
) (*pbiamv1.GetRolePermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	item, err := s.queries.GetRolePermissionBoundary(ctx, req.RoleName)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.GetRolePermissionBoundaryResponse{Boundary: iammapper.ToPBRolePermissionBoundary(item)}, nil
}
