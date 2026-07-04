package grpchandler

import (
	"context"

	iammapper "github.com/tuannm99/podzone/internal/iam/controller/mapper"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain/entity"
	"github.com/tuannm99/podzone/pkg/collection"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
)

func (s *IAMCommandServer) CreateOrganization(
	ctx context.Context,
	req *pbiamv1.CreateOrganizationRequest,
) (*pbiamv1.CreateOrganizationResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	org, err := s.commands.CreateOrganization(ctx, req.Name, req.Slug)
	if err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "organization.created", "organization", org.ID, "", map[string]any{
		"slug": org.Slug,
		"name": org.Name,
	})
	return &pbiamv1.CreateOrganizationResponse{Organization: iammapper.ToPBOrganization(org)}, nil
}

func (s *IAMCommandServer) EnsureRootOrganization(
	ctx context.Context,
	req *pbiamv1.EnsureRootOrganizationRequest,
) (*pbiamv1.EnsureRootOrganizationResponse, error) {
	claims, err := s.claimsFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if claims.IdentitySource != "podzone" {
		return nil, status.Error(codes.PermissionDenied, "organization root bootstrap requires a self-service identity")
	}
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	org, err := s.commands.EnsureRootOrganization(ctx, actorUserID, req.Name, req.Slug)
	if err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "organization.root.bootstrapped", "organization", org.ID, "", map[string]any{
		"root_user_id": actorUserID,
	})
	return &pbiamv1.EnsureRootOrganizationResponse{
		Organization: iammapper.ToPBOrganization(org),
	}, nil
}

func (s *IAMCommandServer) AttachTenantToOrganization(
	ctx context.Context,
	req *pbiamv1.AttachTenantToOrganizationRequest,
) (*pbiamv1.AttachTenantToOrganizationResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.AttachTenantToOrganization(ctx, req.TenantId, req.OrgId); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"organization.tenant.attached",
		"organization_tenant",
		req.OrgId+":"+req.TenantId,
		req.TenantId,
		map[string]any{
			"org_id":    req.OrgId,
			"tenant_id": req.TenantId,
		},
	)
	return &pbiamv1.AttachTenantToOrganizationResponse{}, nil
}

func (s *IAMCommandServer) DetachTenantFromOrganization(
	ctx context.Context,
	req *pbiamv1.DetachTenantFromOrganizationRequest,
) (*pbiamv1.DetachTenantFromOrganizationResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DetachTenantFromOrganization(ctx, req.TenantId); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"organization.tenant.detached",
		"organization_tenant",
		req.OrgId+":"+req.TenantId,
		req.TenantId,
		map[string]any{
			"org_id":    req.OrgId,
			"tenant_id": req.TenantId,
		},
	)
	return &pbiamv1.DetachTenantFromOrganizationResponse{}, nil
}

func (s *IAMCommandServer) AttachServiceControlPolicy(
	ctx context.Context,
	req *pbiamv1.AttachServiceControlPolicyRequest,
) (*pbiamv1.AttachServiceControlPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.AttachServiceControlPolicy(ctx, req.OrgId, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"organization.scp.attached",
		"organization_policy",
		req.OrgId+":"+req.PolicyName,
		"",
		map[string]any{
			"org_id":      req.OrgId,
			"policy_name": req.PolicyName,
		},
	)
	return &pbiamv1.AttachServiceControlPolicyResponse{}, nil
}

func (s *IAMCommandServer) DetachServiceControlPolicy(
	ctx context.Context,
	req *pbiamv1.DetachServiceControlPolicyRequest,
) (*pbiamv1.DetachServiceControlPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.commands.DetachServiceControlPolicy(ctx, req.OrgId, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"organization.scp.detached",
		"organization_policy",
		req.OrgId+":"+req.PolicyName,
		"",
		map[string]any{
			"org_id":      req.OrgId,
			"policy_name": req.PolicyName,
		},
	)
	return &pbiamv1.DetachServiceControlPolicyResponse{}, nil
}

func (s *IAMQueryServer) ListOrganizations(
	ctx context.Context,
	req *pbiamv1.ListOrganizationsRequest,
) (*pbiamv1.ListOrganizationsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	canManagePlatform, err := s.queries.CheckPlatformPermission(ctx, actorUserID, "platform:manage_roles")
	if err != nil {
		return nil, iamStatusError(err)
	}
	query := iammapper.ToCollectionQuery(req.Collection)
	var page collection.Page[iamdomain.Organization]
	if canManagePlatform {
		page, err = s.queries.ListOrganizations(ctx, query)
	} else {
		page, err = s.queries.ListOrganizationsForUser(ctx, actorUserID, query)
	}
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbiamv1.Organization, 0, len(page.Items))
	for i := range page.Items {
		item := page.Items[i]
		out = append(out, iammapper.ToPBOrganization(&item))
	}
	return &pbiamv1.ListOrganizationsResponse{
		Organizations:     out,
		PageInfo:          iammapper.ToPBPageInfo(page),
		CanManagePlatform: canManagePlatform,
	}, nil
}

func (s *IAMQueryServer) ListServiceControlPolicies(
	ctx context.Context,
	req *pbiamv1.ListServiceControlPoliciesRequest,
) (*pbiamv1.ListServiceControlPoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	canManagePlatform, err := s.queries.CheckPlatformPermission(ctx, actorUserID, "platform:manage_roles")
	if err != nil {
		return nil, iamStatusError(err)
	}
	if !canManagePlatform {
		if err := s.queries.RequireOrganizationPermission(
			ctx,
			req.OrgId,
			actorUserID,
			"organization:read",
		); err != nil {
			return nil, iamStatusError(err)
		}
	}
	items, err := s.queries.ListServiceControlPolicies(ctx, req.OrgId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.ListServiceControlPoliciesResponse{Policies: iammapper.ToPBPolicies(items)}, nil
}

func (s *IAMCommandServer) AddOrganizationMember(
	ctx context.Context,
	req *pbiamv1.AddOrganizationMemberRequest,
) (*pbiamv1.AddOrganizationMemberResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequireOrganizationPermission(
		ctx,
		req.OrgId,
		actorUserID,
		"organization:manage_members",
	); err != nil {
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
	if err := s.commands.AddOrganizationMember(ctx, req.OrgId, userID, req.RoleName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "organization.member.added", "organization_member", req.OrgId, "", map[string]any{
		"user_id":   userID,
		"role_name": req.RoleName,
	})
	return &pbiamv1.AddOrganizationMemberResponse{}, nil
}

func (s *IAMCommandServer) RemoveOrganizationMember(
	ctx context.Context,
	req *pbiamv1.RemoveOrganizationMemberRequest,
) (*pbiamv1.RemoveOrganizationMemberResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequireOrganizationPermission(
		ctx,
		req.OrgId,
		actorUserID,
		"organization:manage_members",
	); err != nil {
		return nil, iamStatusError(err)
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.commands.RemoveOrganizationMember(ctx, req.OrgId, userID); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "organization.member.removed", "organization_member", req.OrgId, "", map[string]any{
		"user_id": userID,
	})
	return &pbiamv1.RemoveOrganizationMemberResponse{}, nil
}

func (s *IAMQueryServer) ListOrganizationMembers(
	ctx context.Context,
	req *pbiamv1.ListOrganizationMembersRequest,
) (*pbiamv1.ListOrganizationMembersResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.queries.RequireOrganizationPermission(
		ctx,
		req.OrgId,
		actorUserID,
		"organization:read",
	); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListOrganizationMembers(ctx, req.OrgId, iammapper.ToCollectionQuery(req.Collection))
	if err != nil {
		return nil, iamStatusError(err)
	}
	items := make([]*pbiamv1.OrganizationMembership, 0, len(page.Items))
	for i := range page.Items {
		items = append(items, iammapper.ToPBOrganizationMembership(&page.Items[i]))
	}
	return &pbiamv1.ListOrganizationMembersResponse{
		Memberships: items,
		PageInfo:    iammapper.ToPBPageInfo(page),
	}, nil
}
