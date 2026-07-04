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

func (s *IAMCommandServer) CreateGroup(
	ctx context.Context,
	req *pbiamv1.CreateGroupRequest,
) (*pbiamv1.CreateGroupResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if req.Scope == iamdomain.PolicyScopeOrganization {
		if err := s.queries.RequireOrganizationPermission(
			ctx,
			req.OrgId,
			actorUserID,
			"organization:manage_iam",
		); err != nil {
			return nil, iamStatusError(err)
		}
	} else if req.Scope == iamdomain.PolicyScopePlatform {
		if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
			return nil, iamStatusError(err)
		}
	} else if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	group, err := s.commands.CreateGroup(ctx, iamdomain.CreateGroupInput{
		Scope:       req.Scope,
		OrgID:       req.OrgId,
		TenantID:    req.TenantId,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.CreateGroupResponse{Group: iammapper.ToPBGroup(group)}, nil
}

func (s *IAMCommandServer) DeleteGroup(
	ctx context.Context,
	req *pbiamv1.DeleteGroupRequest,
) (*pbiamv1.DeleteGroupResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	group, err := requireGroupAccess(ctx, s.queries, actorUserID, req.GroupId, true)
	if err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DeleteGroup(ctx, req.GroupId); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.group.deleted",
		"iam_group",
		fmt.Sprintf("%d", req.GroupId),
		group.OrgID,
		map[string]any{
			"group_id": req.GroupId,
		},
	)
	return &pbiamv1.DeleteGroupResponse{}, nil
}

func (s *IAMCommandServer) AddGroupMember(
	ctx context.Context,
	req *pbiamv1.AddGroupMemberRequest,
) (*pbiamv1.AddGroupMemberResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if _, err := requireGroupAccess(ctx, s.queries, actorUserID, req.GroupId, true); err != nil {
		return nil, iamStatusError(err)
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if s.userDirectory == nil {
		return nil, status.Error(codes.Internal, "user directory is not configured")
	}
	user, err := s.userDirectory.GetByID(ctx, userID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	if user == nil || user.ID == 0 {
		return nil, iamStatusError(iamdomain.ErrUserNotFound)
	}
	if err := s.commands.AddGroupMember(ctx, req.GroupId, userID); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.AddGroupMemberResponse{}, nil
}

func (s *IAMCommandServer) RemoveGroupMember(
	ctx context.Context,
	req *pbiamv1.RemoveGroupMemberRequest,
) (*pbiamv1.RemoveGroupMemberResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if _, err := requireGroupAccess(ctx, s.queries, actorUserID, req.GroupId, true); err != nil {
		return nil, iamStatusError(err)
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.commands.RemoveGroupMember(ctx, req.GroupId, userID); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.RemoveGroupMemberResponse{}, nil
}

func (s *IAMCommandServer) AttachGroupPolicy(
	ctx context.Context,
	req *pbiamv1.AttachGroupPolicyRequest,
) (*pbiamv1.AttachGroupPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if _, err := requireGroupAccess(ctx, s.queries, actorUserID, req.GroupId, true); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.AttachGroupPolicy(ctx, req.GroupId, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.AttachGroupPolicyResponse{}, nil
}

func (s *IAMCommandServer) DetachGroupPolicy(
	ctx context.Context,
	req *pbiamv1.DetachGroupPolicyRequest,
) (*pbiamv1.DetachGroupPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if _, err := requireGroupAccess(ctx, s.queries, actorUserID, req.GroupId, true); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DetachGroupPolicy(ctx, req.GroupId, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.DetachGroupPolicyResponse{}, nil
}

func (s *IAMCommandServer) PutGroupInlinePolicy(
	ctx context.Context,
	req *pbiamv1.PutGroupInlinePolicyRequest,
) (*pbiamv1.PutGroupInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if _, err := requireGroupAccess(ctx, s.queries, actorUserID, req.GroupId, true); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.PutGroupInlinePolicy(ctx, iamdomain.PutGroupInlinePolicyInput{
		GroupID:     req.GroupId,
		Name:        req.Name,
		Description: req.Description,
		Statements:  iammapper.FromPBPolicyStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.PutGroupInlinePolicyResponse{}, nil
}

func (s *IAMCommandServer) DeleteGroupInlinePolicy(
	ctx context.Context,
	req *pbiamv1.DeleteGroupInlinePolicyRequest,
) (*pbiamv1.DeleteGroupInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if _, err := requireGroupAccess(ctx, s.queries, actorUserID, req.GroupId, true); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DeleteGroupInlinePolicy(ctx, req.GroupId, req.Name); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.DeleteGroupInlinePolicyResponse{}, nil
}

func (s *IAMQueryServer) ListGroups(
	ctx context.Context,
	req *pbiamv1.ListGroupsRequest,
) (*pbiamv1.ListGroupsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if req.Scope == iamdomain.PolicyScopeOrganization {
		if err := s.queries.RequireOrganizationPermission(
			ctx,
			req.OrgId,
			actorUserID,
			"organization:read",
		); err != nil {
			return nil, iamStatusError(err)
		}
	} else if req.Scope == iamdomain.PolicyScopePlatform {
		if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
			return nil, iamStatusError(err)
		}
	} else if strings.TrimSpace(req.TenantId) != "" {
		if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
			return nil, iamStatusError(err)
		}
	}
	page, err := s.queries.ListGroups(
		ctx,
		req.Scope,
		req.OrgId,
		req.TenantId,
		iammapper.ToCollectionQuery(req.Collection),
	)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbiamv1.Group, 0, len(page.Items))
	for i := range page.Items {
		out = append(out, iammapper.ToPBGroup(&page.Items[i]))
	}
	return &pbiamv1.ListGroupsResponse{
		Groups:   out,
		PageInfo: iammapper.ToPBPageInfo(page),
	}, nil
}

func (s *IAMQueryServer) ListGroupMembers(
	ctx context.Context,
	req *pbiamv1.ListGroupMembersRequest,
) (*pbiamv1.ListGroupMembersResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if _, err := requireGroupAccess(ctx, s.queries, actorUserID, req.GroupId, false); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListGroupMembers(ctx, req.GroupId, iammapper.ToCollectionQuery(req.Collection))
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]uint64, 0, len(page.Items))
	for _, item := range page.Items {
		out = append(out, uint64(item))
	}
	return &pbiamv1.ListGroupMembersResponse{
		UserIds:  out,
		PageInfo: iammapper.ToPBPageInfo(page),
	}, nil
}

func (s *IAMQueryServer) ListGroupPolicies(
	ctx context.Context,
	req *pbiamv1.ListGroupPoliciesRequest,
) (*pbiamv1.ListGroupPoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if _, err := requireGroupAccess(ctx, s.queries, actorUserID, req.GroupId, false); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListGroupPolicies(ctx, req.GroupId, iammapper.ToCollectionQuery(req.Collection))
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.ListGroupPoliciesResponse{
		Policies: iammapper.ToPBPolicies(page.Items),
		PageInfo: iammapper.ToPBPageInfo(page),
	}, nil
}

func (s *IAMQueryServer) GetGroupInlinePolicy(
	ctx context.Context,
	req *pbiamv1.GetGroupInlinePolicyRequest,
) (*pbiamv1.GetGroupInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if _, err := requireGroupAccess(ctx, s.queries, actorUserID, req.GroupId, false); err != nil {
		return nil, iamStatusError(err)
	}
	item, err := s.queries.GetGroupInlinePolicy(ctx, req.GroupId, req.Name)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.GetGroupInlinePolicyResponse{Policy: iammapper.ToPBGroupInlinePolicy(item)}, nil
}

func (s *IAMQueryServer) ListGroupInlinePolicies(
	ctx context.Context,
	req *pbiamv1.ListGroupInlinePoliciesRequest,
) (*pbiamv1.ListGroupInlinePoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if _, err := requireGroupAccess(ctx, s.queries, actorUserID, req.GroupId, false); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListGroupInlinePolicies(ctx, req.GroupId, iammapper.ToCollectionQuery(req.Collection))
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbiamv1.GroupInlinePolicy, 0, len(page.Items))
	for i := range page.Items {
		out = append(out, iammapper.ToPBGroupInlinePolicy(&page.Items[i]))
	}
	return &pbiamv1.ListGroupInlinePoliciesResponse{
		Policies: out,
		PageInfo: iammapper.ToPBPageInfo(page),
	}, nil
}
