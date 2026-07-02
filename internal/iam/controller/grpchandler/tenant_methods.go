package grpchandler

import (
	"context"
	"errors"
	"fmt"
	"strings"

	iammapper "github.com/tuannm99/podzone/internal/iam/controller/mapper"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
)

func (s *IAMCommandServer) CreateTenant(
	ctx context.Context,
	req *pbiamv1.CreateTenantRequest,
) (*pbiamv1.CreateTenantResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if req.OwnerUserId != 0 && uint64(actorUserID) != req.OwnerUserId {
		return nil, status.Error(codes.InvalidArgument, "owner_user_id must match authenticated user")
	}
	memberships, err := s.queries.ListUserTenants(ctx, actorUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	canCreateTenant := len(memberships) == 0
	if !canCreateTenant {
		allowed, err := s.queries.CheckPlatformPermission(ctx, actorUserID, "tenant:create")
		if err != nil {
			return nil, iamStatusError(err)
		}
		canCreateTenant = allowed
	}
	if !canCreateTenant {
		return nil, iamStatusError(iamdomain.ErrPermissionDenied)
	}

	tenant, err := s.commands.CreateTenant(ctx, actorUserID, iamdomain.CreateTenantCmd{
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		return nil, iamStatusError(err)
	}

	membership, err := s.queries.GetMembership(ctx, tenant.ID, actorUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}

	s.recordAudit(ctx, actorUserID, "tenant.created", "tenant", tenant.ID, tenant.ID, map[string]any{
		"slug": tenant.Slug,
		"name": tenant.Name,
	})
	return &pbiamv1.CreateTenantResponse{
		Tenant:          iammapper.ToPBTenant(tenant),
		OwnerMembership: iammapper.ToPBMembership(membership),
	}, nil
}

func (s *IAMCommandServer) AddTenantMember(
	ctx context.Context,
	req *pbiamv1.AddTenantMemberRequest,
) (*pbiamv1.AddTenantMemberResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}

	if err := s.commands.AddMember(ctx, req.TenantId, userID, req.RoleName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"tenant.member.added",
		"tenant_member",
		fmt.Sprintf("%s:%d", req.TenantId, req.UserId),
		req.TenantId,
		map[string]any{
			"user_id":   userID,
			"role_name": req.RoleName,
		},
	)
	return &pbiamv1.AddTenantMemberResponse{}, nil
}

func (s *IAMCommandServer) AddTenantMemberByIdentity(
	ctx context.Context,
	req *pbiamv1.AddTenantMemberByIdentityRequest,
) (*pbiamv1.AddTenantMemberByIdentityResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	identity := strings.TrimSpace(req.Identity)
	if identity == "" {
		return nil, status.Error(codes.InvalidArgument, "identity is required")
	}
	if s.userDirectory == nil {
		return nil, status.Error(codes.Internal, "user directory is not configured")
	}

	createdUser := false
	user, err := s.userDirectory.GetByIdentity(ctx, identity)
	if err != nil {
		if strings.Contains(identity, "@") && errors.Is(err, iamdomain.ErrUserNotFound) {
			user, createdUser, err = s.userDirectory.EnsureByEmail(ctx, identity)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		} else if errors.Is(err, iamdomain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if user == nil || user.ID == 0 {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if err := s.commands.AddMember(ctx, req.TenantId, user.ID, req.RoleName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"tenant.member.identity_added",
		"tenant_member",
		fmt.Sprintf("%s:%d", req.TenantId, user.ID),
		req.TenantId,
		map[string]any{
			"user_id":      user.ID,
			"identity":     identity,
			"role_name":    req.RoleName,
			"created_user": createdUser,
		},
	)
	return &pbiamv1.AddTenantMemberByIdentityResponse{
		UserId:      uint64(user.ID),
		CreatedUser: createdUser,
	}, nil
}

func (s *IAMCommandServer) CreateTenantInvite(
	ctx context.Context,
	req *pbiamv1.CreateTenantInviteRequest,
) (*pbiamv1.CreateTenantInviteResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	invite, rawToken, err := s.commands.CreateInvite(ctx, req.TenantId, req.Email, req.RoleName, actorUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	acceptURL := fmt.Sprintf("%s/auth/invite/accept?token=%s", inviteAcceptBaseURL(s.appRedirectURL), rawToken)
	s.recordAudit(ctx, actorUserID, "tenant.invite.created", "tenant_invite", invite.ID, req.TenantId, map[string]any{
		"email":     invite.Email,
		"role_name": invite.RoleName,
	})
	return &pbiamv1.CreateTenantInviteResponse{
		Invite:      iammapper.ToPBTenantInvite(invite),
		InviteToken: rawToken,
		AcceptUrl:   acceptURL,
	}, nil
}

func (s *IAMCommandServer) RevokeTenantInvite(
	ctx context.Context,
	req *pbiamv1.RevokeTenantInviteRequest,
) (*pbiamv1.RevokeTenantInviteResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	invite, err := s.queries.GetInvite(ctx, req.InviteId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.queries.RequirePermission(ctx, invite.TenantID, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.RevokeInvite(ctx, req.InviteId); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"tenant.invite.revoked",
		"tenant_invite",
		req.InviteId,
		invite.TenantID,
		map[string]any{
			"email": invite.Email,
		},
	)
	return &pbiamv1.RevokeTenantInviteResponse{}, nil
}

func (s *IAMCommandServer) AcceptTenantInvite(
	ctx context.Context,
	req *pbiamv1.AcceptTenantInviteRequest,
) (*pbiamv1.AcceptTenantInviteResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if s.userDirectory == nil {
		return nil, status.Error(codes.Internal, "user directory is not configured")
	}
	user, err := s.userDirectory.GetByID(ctx, actorUserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	membership, err := s.commands.AcceptInvite(ctx, req.InviteToken, actorUserID, user.Email)
	if err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"tenant.invite.accepted",
		"tenant_invite",
		membership.TenantID,
		membership.TenantID,
		map[string]any{
			"role_name": membership.RoleName,
		},
	)
	return &pbiamv1.AcceptTenantInviteResponse{
		Membership: iammapper.ToPBMembership(membership),
	}, nil
}

func (s *IAMCommandServer) RemoveTenantMember(
	ctx context.Context,
	req *pbiamv1.RemoveTenantMemberRequest,
) (*pbiamv1.RemoveTenantMemberResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.commands.RemoveMember(ctx, req.TenantId, userID); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"tenant.member.removed",
		"tenant_member",
		fmt.Sprintf("%s:%d", req.TenantId, req.UserId),
		req.TenantId,
		map[string]any{
			"user_id": userID,
		},
	)
	return &pbiamv1.RemoveTenantMemberResponse{}, nil
}

func (s *IAMCommandServer) PutTenantUserInlinePolicy(
	ctx context.Context,
	req *pbiamv1.PutTenantUserInlinePolicyRequest,
) (*pbiamv1.PutTenantUserInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.PutTenantUserInlinePolicy(ctx, iamdomain.PutTenantUserInlinePolicyInput{
		TenantID:    req.TenantId,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Statements:  iammapper.FromPBPolicyStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.PutTenantUserInlinePolicyResponse{}, nil
}

func (s *IAMCommandServer) DeleteTenantUserInlinePolicy(
	ctx context.Context,
	req *pbiamv1.DeleteTenantUserInlinePolicyRequest,
) (*pbiamv1.DeleteTenantUserInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DeleteTenantUserInlinePolicy(ctx, req.TenantId, userID, req.Name); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.DeleteTenantUserInlinePolicyResponse{}, nil
}

func (s *IAMCommandServer) AttachTenantUserPolicy(
	ctx context.Context,
	req *pbiamv1.AttachTenantUserPolicyRequest,
) (*pbiamv1.AttachTenantUserPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.AttachTenantUserPolicy(ctx, req.TenantId, userID, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.tenant_user_policy.attached",
		"iam_policy_attachment",
		req.PolicyName,
		req.TenantId,
		map[string]any{
			"user_id":     userID,
			"tenant_id":   req.TenantId,
			"policy_name": req.PolicyName,
		},
	)
	return &pbiamv1.AttachTenantUserPolicyResponse{}, nil
}

func (s *IAMCommandServer) DetachTenantUserPolicy(
	ctx context.Context,
	req *pbiamv1.DetachTenantUserPolicyRequest,
) (*pbiamv1.DetachTenantUserPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DetachTenantUserPolicy(ctx, req.TenantId, userID, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.tenant_user_policy.detached",
		"iam_policy_attachment",
		req.PolicyName,
		req.TenantId,
		map[string]any{
			"user_id":     userID,
			"tenant_id":   req.TenantId,
			"policy_name": req.PolicyName,
		},
	)
	return &pbiamv1.DetachTenantUserPolicyResponse{}, nil
}

func (s *IAMCommandServer) PutTenantUserPermissionBoundary(
	ctx context.Context,
	req *pbiamv1.PutTenantUserPermissionBoundaryRequest,
) (*pbiamv1.PutTenantUserPermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.PutTenantUserPermissionBoundary(ctx, req.TenantId, userID, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.tenant_user_boundary.put",
		"iam_permission_boundary",
		req.PolicyName,
		req.TenantId,
		map[string]any{
			"user_id":     userID,
			"tenant_id":   req.TenantId,
			"policy_name": req.PolicyName,
		},
	)
	return &pbiamv1.PutTenantUserPermissionBoundaryResponse{}, nil
}

func (s *IAMCommandServer) DeleteTenantUserPermissionBoundary(
	ctx context.Context,
	req *pbiamv1.DeleteTenantUserPermissionBoundaryRequest,
) (*pbiamv1.DeleteTenantUserPermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DeleteTenantUserPermissionBoundary(ctx, req.TenantId, userID); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.tenant_user_boundary.deleted",
		"iam_permission_boundary",
		fmt.Sprintf("%s:%d", req.TenantId, userID),
		req.TenantId,
		map[string]any{
			"user_id":   userID,
			"tenant_id": req.TenantId,
		},
	)
	return &pbiamv1.DeleteTenantUserPermissionBoundaryResponse{}, nil
}

func (s *IAMQueryServer) ListTenantInvites(
	ctx context.Context,
	req *pbiamv1.ListTenantInvitesRequest,
) (*pbiamv1.ListTenantInvitesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.queries.ListTenantInvites(ctx, req.TenantId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbiamv1.TenantInvite, 0, len(items))
	for i := range items {
		out = append(out, iammapper.ToPBTenantInvite(&items[i]))
	}
	return &pbiamv1.ListTenantInvitesResponse{Invites: out}, nil
}

func (s *IAMQueryServer) GetTenantMembership(
	ctx context.Context,
	req *pbiamv1.GetTenantMembershipRequest,
) (*pbiamv1.GetTenantMembershipResponse, error) {
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}

	membership, err := s.queries.GetMembership(ctx, req.TenantId, userID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.GetTenantMembershipResponse{
		Membership: iammapper.ToPBMembership(membership),
	}, nil
}

func (s *IAMQueryServer) ListUserTenants(
	ctx context.Context,
	req *pbiamv1.ListUserTenantsRequest,
) (*pbiamv1.ListUserTenantsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if userID != actorUserID {
		return nil, status.Error(codes.PermissionDenied, "cannot list another user's tenant memberships")
	}
	items, err := s.queries.ListUserTenants(ctx, userID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbiamv1.TenantMembership, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, iammapper.ToPBMembership(&item))
	}
	return &pbiamv1.ListUserTenantsResponse{Memberships: out}, nil
}

func (s *IAMQueryServer) ListTenantMembers(
	ctx context.Context,
	req *pbiamv1.ListTenantMembersRequest,
) (*pbiamv1.ListTenantMembersResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.queries.ListTenantMembers(ctx, req.TenantId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbiamv1.TenantMembership, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, iammapper.ToPBMembership(&item))
	}
	return &pbiamv1.ListTenantMembersResponse{Memberships: out}, nil
}

func (s *IAMQueryServer) ListTenantUserPolicies(
	ctx context.Context,
	req *pbiamv1.ListTenantUserPoliciesRequest,
) (*pbiamv1.ListTenantUserPoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListTenantUserPolicies(
		ctx,
		req.TenantId,
		userID,
		iammapper.ToCollectionQuery(req.Collection),
	)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.ListTenantUserPoliciesResponse{
		Policies: iammapper.ToPBPolicies(page.Items),
		PageInfo: iammapper.ToPBPageInfo(page),
	}, nil
}

func (s *IAMQueryServer) GetTenantUserInlinePolicy(
	ctx context.Context,
	req *pbiamv1.GetTenantUserInlinePolicyRequest,
) (*pbiamv1.GetTenantUserInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	item, err := s.queries.GetTenantUserInlinePolicy(ctx, req.TenantId, userID, req.Name)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.GetTenantUserInlinePolicyResponse{Policy: iammapper.ToPBUserInlinePolicy(item)}, nil
}

func (s *IAMQueryServer) ListTenantUserInlinePolicies(
	ctx context.Context,
	req *pbiamv1.ListTenantUserInlinePoliciesRequest,
) (*pbiamv1.ListTenantUserInlinePoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListTenantUserInlinePolicies(
		ctx,
		req.TenantId,
		userID,
		iammapper.ToCollectionQuery(req.Collection),
	)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbiamv1.UserInlinePolicy, 0, len(page.Items))
	for i := range page.Items {
		out = append(out, iammapper.ToPBUserInlinePolicy(&page.Items[i]))
	}
	return &pbiamv1.ListTenantUserInlinePoliciesResponse{
		Policies: out,
		PageInfo: iammapper.ToPBPageInfo(page),
	}, nil
}

func (s *IAMQueryServer) GetTenantUserPermissionBoundary(
	ctx context.Context,
	req *pbiamv1.GetTenantUserPermissionBoundaryRequest,
) (*pbiamv1.GetTenantUserPermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.queries.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	item, err := s.queries.GetTenantUserPermissionBoundary(ctx, req.TenantId, userID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.GetTenantUserPermissionBoundaryResponse{Boundary: iammapper.ToPBPermissionBoundary(item)}, nil
}
