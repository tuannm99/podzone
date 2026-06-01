package grpchandler

import (
	"context"
	"fmt"

	iammapper "github.com/tuannm99/podzone/internal/iam/controller/mapper"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
)

func (s *IAMCommandServer) PutPlatformUserInlinePolicy(
	ctx context.Context,
	req *pbiamv1.PutPlatformUserInlinePolicyRequest,
) (*pbiamv1.PutPlatformUserInlinePolicyResponse, error) {
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
	if err := s.commands.PutPlatformUserInlinePolicy(ctx, iamdomain.PutPlatformUserInlinePolicyInput{
		UserID:      targetUserID,
		Name:        req.Name,
		Description: req.Description,
		Statements:  iammapper.FromPBPolicyStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.PutPlatformUserInlinePolicyResponse{}, nil
}

func (s *IAMCommandServer) DeletePlatformUserInlinePolicy(
	ctx context.Context,
	req *pbiamv1.DeletePlatformUserInlinePolicyRequest,
) (*pbiamv1.DeletePlatformUserInlinePolicyResponse, error) {
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
	if err := s.commands.DeletePlatformUserInlinePolicy(ctx, targetUserID, req.Name); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.DeletePlatformUserInlinePolicyResponse{}, nil
}

func (s *IAMCommandServer) AttachPlatformUserPolicy(
	ctx context.Context,
	req *pbiamv1.AttachPlatformUserPolicyRequest,
) (*pbiamv1.AttachPlatformUserPolicyResponse, error) {
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
	if err := s.commands.AttachPlatformUserPolicy(ctx, targetUserID, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.platform_user_policy.attached",
		"iam_policy_attachment",
		req.PolicyName,
		"",
		map[string]any{
			"target_user_id": targetUserID,
			"policy_name":    req.PolicyName,
		},
	)
	return &pbiamv1.AttachPlatformUserPolicyResponse{}, nil
}

func (s *IAMCommandServer) DetachPlatformUserPolicy(
	ctx context.Context,
	req *pbiamv1.DetachPlatformUserPolicyRequest,
) (*pbiamv1.DetachPlatformUserPolicyResponse, error) {
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
	if err := s.commands.DetachPlatformUserPolicy(ctx, targetUserID, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.platform_user_policy.detached",
		"iam_policy_attachment",
		req.PolicyName,
		"",
		map[string]any{
			"target_user_id": targetUserID,
			"policy_name":    req.PolicyName,
		},
	)
	return &pbiamv1.DetachPlatformUserPolicyResponse{}, nil
}

func (s *IAMCommandServer) PutPlatformUserPermissionBoundary(
	ctx context.Context,
	req *pbiamv1.PutPlatformUserPermissionBoundaryRequest,
) (*pbiamv1.PutPlatformUserPermissionBoundaryResponse, error) {
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
	if err := s.commands.PutPlatformUserPermissionBoundary(ctx, targetUserID, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.platform_user_boundary.put",
		"iam_permission_boundary",
		req.PolicyName,
		"",
		map[string]any{
			"target_user_id": targetUserID,
			"policy_name":    req.PolicyName,
		},
	)
	return &pbiamv1.PutPlatformUserPermissionBoundaryResponse{}, nil
}

func (s *IAMCommandServer) DeletePlatformUserPermissionBoundary(
	ctx context.Context,
	req *pbiamv1.DeletePlatformUserPermissionBoundaryRequest,
) (*pbiamv1.DeletePlatformUserPermissionBoundaryResponse, error) {
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
	if err := s.commands.DeletePlatformUserPermissionBoundary(ctx, targetUserID); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.platform_user_boundary.deleted",
		"iam_permission_boundary",
		fmt.Sprintf("%d", targetUserID),
		"",
		map[string]any{
			"target_user_id": targetUserID,
		},
	)
	return &pbiamv1.DeletePlatformUserPermissionBoundaryResponse{}, nil
}

func (s *IAMQueryServer) ListPlatformUserPolicies(
	ctx context.Context,
	req *pbiamv1.ListPlatformUserPoliciesRequest,
) (*pbiamv1.ListPlatformUserPoliciesResponse, error) {
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
	items, err := s.queries.ListPlatformUserPolicies(ctx, targetUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.ListPlatformUserPoliciesResponse{Policies: iammapper.ToPBPolicies(items)}, nil
}

func (s *IAMQueryServer) GetPlatformUserInlinePolicy(
	ctx context.Context,
	req *pbiamv1.GetPlatformUserInlinePolicyRequest,
) (*pbiamv1.GetPlatformUserInlinePolicyResponse, error) {
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
	item, err := s.queries.GetPlatformUserInlinePolicy(ctx, targetUserID, req.Name)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.GetPlatformUserInlinePolicyResponse{Policy: iammapper.ToPBUserInlinePolicy(item)}, nil
}

func (s *IAMQueryServer) ListPlatformUserInlinePolicies(
	ctx context.Context,
	req *pbiamv1.ListPlatformUserInlinePoliciesRequest,
) (*pbiamv1.ListPlatformUserInlinePoliciesResponse, error) {
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
	items, err := s.queries.ListPlatformUserInlinePolicies(ctx, targetUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbiamv1.UserInlinePolicy, 0, len(items))
	for i := range items {
		out = append(out, iammapper.ToPBUserInlinePolicy(&items[i]))
	}
	return &pbiamv1.ListPlatformUserInlinePoliciesResponse{Policies: out}, nil
}

func (s *IAMQueryServer) GetPlatformUserPermissionBoundary(
	ctx context.Context,
	req *pbiamv1.GetPlatformUserPermissionBoundaryRequest,
) (*pbiamv1.GetPlatformUserPermissionBoundaryResponse, error) {
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
	item, err := s.queries.GetPlatformUserPermissionBoundary(ctx, targetUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.GetPlatformUserPermissionBoundaryResponse{Boundary: iammapper.ToPBPermissionBoundary(item)}, nil
}
