package grpchandler

import (
	"context"

	iammapper "github.com/tuannm99/podzone/internal/iam/controller/mapper"
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
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	page, err := s.queries.ListOrganizations(ctx, iammapper.ToCollectionQuery(req.Collection))
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbiamv1.Organization, 0, len(page.Items))
	for i := range page.Items {
		item := page.Items[i]
		out = append(out, iammapper.ToPBOrganization(&item))
	}
	return &pbiamv1.ListOrganizationsResponse{
		Organizations: out,
		PageInfo:      iammapper.ToPBPageInfo(page),
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
	if err := s.queries.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.queries.ListServiceControlPolicies(ctx, req.OrgId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbiamv1.ListServiceControlPoliciesResponse{Policies: iammapper.ToPBPolicies(items)}, nil
}
