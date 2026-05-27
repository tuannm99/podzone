package grpchandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	iamconfig "github.com/tuannm99/podzone/internal/iam/config"
	iammapper "github.com/tuannm99/podzone/internal/iam/controller/mapper"
	iamdomain "github.com/tuannm99/podzone/internal/iam/entity"
	iaminputport "github.com/tuannm99/podzone/internal/iam/inputport"
	iamoutputport "github.com/tuannm99/podzone/internal/iam/outputport"
	"github.com/tuannm99/podzone/pkg/pdauthn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

type IAMServer struct {
	pbauthv1.UnimplementedIAMServiceServer
	iamUC          iaminputport.IAMUsecase
	auditRep       iamoutputport.AuditLogRepository
	userDirectory  iamoutputport.UserDirectory
	appRedirectURL string
	verifier       *pdauthn.Verifier
}

func NewIAMServer(
	iamUC iaminputport.IAMUsecase,
	auditRep iamoutputport.AuditLogRepository,
	userDirectory iamoutputport.UserDirectory,
	cfg iamconfig.ServerConfig,
) *IAMServer {
	return &IAMServer{
		iamUC:          iamUC,
		auditRep:       auditRep,
		userDirectory:  userDirectory,
		appRedirectURL: cfg.AppRedirectURL,
		verifier:       pdauthn.NewVerifier(cfg.Authn),
	}
}

func (s *IAMServer) CreateTenant(
	ctx context.Context,
	req *pbauthv1.CreateTenantRequest,
) (*pbauthv1.CreateTenantResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if req.OwnerUserId != 0 && uint64(actorUserID) != req.OwnerUserId {
		return nil, status.Error(codes.InvalidArgument, "owner_user_id must match authenticated user")
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "tenant:create"); err != nil {
		return nil, iamStatusError(err)
	}

	tenant, err := s.iamUC.CreateTenant(ctx, actorUserID, iamdomain.CreateTenantCmd{
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		return nil, iamStatusError(err)
	}

	membership, err := s.iamUC.GetMembership(ctx, tenant.ID, actorUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}

	s.recordAudit(ctx, actorUserID, "tenant.created", "tenant", tenant.ID, tenant.ID, map[string]any{
		"slug": tenant.Slug,
		"name": tenant.Name,
	})
	return &pbauthv1.CreateTenantResponse{
		Tenant:          iammapper.ToPBTenant(tenant),
		OwnerMembership: iammapper.ToPBMembership(membership),
	}, nil
}

func (s *IAMServer) AssumeRole(
	ctx context.Context,
	req *pbauthv1.IAMAssumeRoleRequest,
) (*pbauthv1.IAMAssumeRoleResponse, error) {
	claims, err := s.claimsFromAccessToken(req.AccessToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if claims.UserID == 0 {
		return nil, status.Error(codes.Unauthenticated, "access token missing user_id")
	}

	assumedRole, err := s.iamUC.AssumeRole(ctx, iamdomain.AssumeRoleInput{
		UserID:           uint(claims.UserID),
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
	return &pbauthv1.IAMAssumeRoleResponse{
		AssumedRole: iammapper.ToPBIAMAssumedRole(assumedRole),
	}, nil
}

func (s *IAMServer) CreateOrganization(
	ctx context.Context,
	req *pbauthv1.CreateOrganizationRequest,
) (*pbauthv1.CreateOrganizationResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	org, err := s.iamUC.CreateOrganization(ctx, req.Name, req.Slug)
	if err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "organization.created", "organization", org.ID, "", map[string]any{
		"slug": org.Slug,
		"name": org.Name,
	})
	return &pbauthv1.CreateOrganizationResponse{Organization: iammapper.ToPBOrganization(org)}, nil
}

func (s *IAMServer) ListOrganizations(
	ctx context.Context,
	_ *pbauthv1.ListOrganizationsRequest,
) (*pbauthv1.ListOrganizationsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListOrganizations(ctx)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbauthv1.Organization, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, iammapper.ToPBOrganization(&item))
	}
	return &pbauthv1.ListOrganizationsResponse{Organizations: out}, nil
}

func (s *IAMServer) AttachTenantToOrganization(
	ctx context.Context,
	req *pbauthv1.AttachTenantToOrganizationRequest,
) (*pbauthv1.AttachTenantToOrganizationResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.AttachTenantToOrganization(ctx, req.TenantId, req.OrgId); err != nil {
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
	return &pbauthv1.AttachTenantToOrganizationResponse{}, nil
}

func (s *IAMServer) DetachTenantFromOrganization(
	ctx context.Context,
	req *pbauthv1.DetachTenantFromOrganizationRequest,
) (*pbauthv1.DetachTenantFromOrganizationResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DetachTenantFromOrganization(ctx, req.TenantId); err != nil {
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
	return &pbauthv1.DetachTenantFromOrganizationResponse{}, nil
}

func (s *IAMServer) AttachServiceControlPolicy(
	ctx context.Context,
	req *pbauthv1.AttachServiceControlPolicyRequest,
) (*pbauthv1.AttachServiceControlPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.AttachServiceControlPolicy(ctx, req.OrgId, req.PolicyName); err != nil {
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
	return &pbauthv1.AttachServiceControlPolicyResponse{}, nil
}

func (s *IAMServer) DetachServiceControlPolicy(
	ctx context.Context,
	req *pbauthv1.DetachServiceControlPolicyRequest,
) (*pbauthv1.DetachServiceControlPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DetachServiceControlPolicy(ctx, req.OrgId, req.PolicyName); err != nil {
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
	return &pbauthv1.DetachServiceControlPolicyResponse{}, nil
}

func (s *IAMServer) ListServiceControlPolicies(
	ctx context.Context,
	req *pbauthv1.ListServiceControlPoliciesRequest,
) (*pbauthv1.ListServiceControlPoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListServiceControlPolicies(ctx, req.OrgId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.ListServiceControlPoliciesResponse{Policies: iammapper.ToPBPolicies(items)}, nil
}

func (s *IAMServer) AddTenantMember(
	ctx context.Context,
	req *pbauthv1.AddTenantMemberRequest,
) (*pbauthv1.AddTenantMemberResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}

	if err := s.iamUC.AddMember(ctx, req.TenantId, userID, req.RoleName); err != nil {
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
	return &pbauthv1.AddTenantMemberResponse{}, nil
}

func (s *IAMServer) AddTenantMemberByIdentity(
	ctx context.Context,
	req *pbauthv1.AddTenantMemberByIdentityRequest,
) (*pbauthv1.AddTenantMemberByIdentityResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
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

	if err := s.iamUC.AddMember(ctx, req.TenantId, user.ID, req.RoleName); err != nil {
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
	return &pbauthv1.AddTenantMemberByIdentityResponse{
		UserId:      uint64(user.ID),
		CreatedUser: createdUser,
	}, nil
}

func (s *IAMServer) CreateTenantInvite(
	ctx context.Context,
	req *pbauthv1.CreateTenantInviteRequest,
) (*pbauthv1.CreateTenantInviteResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	invite, rawToken, err := s.iamUC.CreateInvite(ctx, req.TenantId, req.Email, req.RoleName, actorUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	acceptURL := fmt.Sprintf("%s/auth/invite/accept?token=%s", inviteAcceptBaseURL(s.appRedirectURL), rawToken)
	s.recordAudit(ctx, actorUserID, "tenant.invite.created", "tenant_invite", invite.ID, req.TenantId, map[string]any{
		"email":     invite.Email,
		"role_name": invite.RoleName,
	})
	return &pbauthv1.CreateTenantInviteResponse{
		Invite:      iammapper.ToPBTenantInvite(invite),
		InviteToken: rawToken,
		AcceptUrl:   acceptURL,
	}, nil
}

func (s *IAMServer) ListTenantInvites(
	ctx context.Context,
	req *pbauthv1.ListTenantInvitesRequest,
) (*pbauthv1.ListTenantInvitesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListTenantInvites(ctx, req.TenantId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbauthv1.TenantInvite, 0, len(items))
	for i := range items {
		out = append(out, iammapper.ToPBTenantInvite(&items[i]))
	}
	return &pbauthv1.ListTenantInvitesResponse{Invites: out}, nil
}

func (s *IAMServer) RevokeTenantInvite(
	ctx context.Context,
	req *pbauthv1.RevokeTenantInviteRequest,
) (*pbauthv1.RevokeTenantInviteResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	invite, err := s.iamUC.GetInvite(ctx, req.InviteId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.RequirePermission(ctx, invite.TenantID, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.RevokeInvite(ctx, req.InviteId); err != nil {
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
	return &pbauthv1.RevokeTenantInviteResponse{}, nil
}

func (s *IAMServer) AcceptTenantInvite(
	ctx context.Context,
	req *pbauthv1.AcceptTenantInviteRequest,
) (*pbauthv1.AcceptTenantInviteResponse, error) {
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
	membership, err := s.iamUC.AcceptInvite(ctx, req.InviteToken, actorUserID, user.Email)
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
	return &pbauthv1.AcceptTenantInviteResponse{
		Membership: iammapper.ToPBMembership(membership),
	}, nil
}

func (s *IAMServer) GetTenantMembership(
	ctx context.Context,
	req *pbauthv1.GetTenantMembershipRequest,
) (*pbauthv1.GetTenantMembershipResponse, error) {
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}

	membership, err := s.iamUC.GetMembership(ctx, req.TenantId, userID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.GetTenantMembershipResponse{
		Membership: iammapper.ToPBMembership(membership),
	}, nil
}

func (s *IAMServer) CheckPermission(
	ctx context.Context,
	req *pbauthv1.CheckPermissionRequest,
) (*pbauthv1.CheckPermissionResponse, error) {
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}

	allowed, err := s.iamUC.CheckPermissionForResource(
		ctx,
		req.TenantId,
		userID,
		req.Permission,
		req.Resource,
	)
	if err != nil {
		if errors.Is(err, iamdomain.ErrPermissionDenied) || errors.Is(err, iamdomain.ErrInactiveMembership) {
			return &pbauthv1.CheckPermissionResponse{Allowed: false}, nil
		}
		return nil, iamStatusError(err)
	}
	return &pbauthv1.CheckPermissionResponse{Allowed: allowed}, nil
}

func (s *IAMServer) CheckPlatformPermission(
	ctx context.Context,
	req *pbauthv1.CheckPlatformPermissionRequest,
) (*pbauthv1.CheckPermissionResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	allowed, err := s.iamUC.CheckPlatformPermission(ctx, actorUserID, req.Permission)
	if err != nil {
		if errors.Is(err, iamdomain.ErrPermissionDenied) {
			return &pbauthv1.CheckPermissionResponse{Allowed: false}, nil
		}
		return nil, iamStatusError(err)
	}
	return &pbauthv1.CheckPermissionResponse{Allowed: allowed}, nil
}

func (s *IAMServer) ListUserTenants(
	ctx context.Context,
	req *pbauthv1.ListUserTenantsRequest,
) (*pbauthv1.ListUserTenantsResponse, error) {
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
	items, err := s.iamUC.ListUserTenants(ctx, userID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbauthv1.TenantMembership, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, iammapper.ToPBMembership(&item))
	}
	return &pbauthv1.ListUserTenantsResponse{Memberships: out}, nil
}

func (s *IAMServer) ListPlatformRoles(
	ctx context.Context,
	req *pbauthv1.ListPlatformRolesRequest,
) (*pbauthv1.ListPlatformRolesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListPlatformRoles(ctx, targetUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbauthv1.PlatformRoleMembership, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, iammapper.ToPBPlatformMembership(&item))
	}
	return &pbauthv1.ListPlatformRolesResponse{Memberships: out}, nil
}

func (s *IAMServer) AddPlatformRole(
	ctx context.Context,
	req *pbauthv1.AddPlatformRoleRequest,
) (*pbauthv1.AddPlatformRoleResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.AddPlatformRole(ctx, targetUserID, req.RoleName); err != nil {
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
	return &pbauthv1.AddPlatformRoleResponse{}, nil
}

func (s *IAMServer) CreatePolicy(
	ctx context.Context,
	req *pbauthv1.CreatePolicyRequest,
) (*pbauthv1.CreatePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	policy, statements, err := s.iamUC.CreatePolicy(ctx, iamdomain.CreatePolicyInput{
		Scope:       req.Scope,
		Name:        req.Name,
		Description: req.Description,
		Statements:  iammapper.FromPBPolicyStatements(req.Statements),
	})
	if err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "iam.policy.created", "iam_policy", policy.Name, "", map[string]any{
		"scope":      policy.Scope,
		"statements": len(statements),
	})
	return &pbauthv1.CreatePolicyResponse{
		Policy:     iammapper.ToPBPolicy(policy),
		Statements: iammapper.ToPBPolicyStatements(statements),
	}, nil
}

func (s *IAMServer) CreatePolicyVersion(
	ctx context.Context,
	req *pbauthv1.CreatePolicyVersionRequest,
) (*pbauthv1.CreatePolicyVersionResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	version, statements, err := s.iamUC.CreatePolicyVersion(ctx, iamdomain.CreatePolicyVersionInput{
		PolicyName:   req.Name,
		Statements:   iammapper.FromPBPolicyStatements(req.Statements),
		SetAsDefault: req.SetAsDefault,
	})
	if err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "iam.policy.version.created", "iam_policy_version", req.Name, "", map[string]any{
		"version":        version.Version,
		"set_as_default": req.SetAsDefault,
		"statements":     len(statements),
	})
	return &pbauthv1.CreatePolicyVersionResponse{
		PolicyVersion: iammapper.ToPBPolicyVersion(version),
		Statements:    iammapper.ToPBPolicyStatements(statements),
	}, nil
}

func (s *IAMServer) GetPolicy(
	ctx context.Context,
	req *pbauthv1.GetPolicyRequest,
) (*pbauthv1.GetPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	policy, statements, err := s.iamUC.GetPolicy(ctx, req.Name)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.GetPolicyResponse{
		Policy:     iammapper.ToPBPolicy(policy),
		Statements: iammapper.ToPBPolicyStatements(statements),
	}, nil
}

func (s *IAMServer) ListPolicyVersions(
	ctx context.Context,
	req *pbauthv1.ListPolicyVersionsRequest,
) (*pbauthv1.ListPolicyVersionsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListPolicyVersions(ctx, req.Name)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbauthv1.PolicyVersion, 0, len(items))
	for i := range items {
		out = append(out, iammapper.ToPBPolicyVersion(&items[i]))
	}
	return &pbauthv1.ListPolicyVersionsResponse{Versions: out}, nil
}

func (s *IAMServer) SetDefaultPolicyVersion(
	ctx context.Context,
	req *pbauthv1.SetDefaultPolicyVersionRequest,
) (*pbauthv1.SetDefaultPolicyVersionResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.SetDefaultPolicyVersion(ctx, req.Name, req.Version); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.policy.version.set_default",
		"iam_policy_version",
		req.Name,
		"",
		map[string]any{
			"version": req.Version,
		},
	)
	return &pbauthv1.SetDefaultPolicyVersionResponse{}, nil
}

func (s *IAMServer) DeletePolicyVersion(
	ctx context.Context,
	req *pbauthv1.DeletePolicyVersionRequest,
) (*pbauthv1.DeletePolicyVersionResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DeletePolicyVersion(ctx, req.Name, req.Version); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "iam.policy.version.deleted", "iam_policy_version", req.Name, "", map[string]any{
		"version": req.Version,
	})
	return &pbauthv1.DeletePolicyVersionResponse{}, nil
}

func (s *IAMServer) ListPolicies(
	ctx context.Context,
	req *pbauthv1.ListPoliciesRequest,
) (*pbauthv1.ListPoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListPolicies(ctx, req.Scope)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.ListPoliciesResponse{Policies: iammapper.ToPBPolicies(items)}, nil
}

func (s *IAMServer) ListPolicyAttachments(
	ctx context.Context,
	req *pbauthv1.ListPolicyAttachmentsRequest,
) (*pbauthv1.ListPolicyAttachmentsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListPolicyAttachments(ctx, req.Name)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.ListPolicyAttachmentsResponse{Attachments: iammapper.ToPBPolicyAttachments(items)}, nil
}

func (s *IAMServer) DeletePolicy(
	ctx context.Context,
	req *pbauthv1.DeletePolicyRequest,
) (*pbauthv1.DeletePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DeletePolicy(ctx, req.Name); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "iam.policy.deleted", "iam_policy", req.Name, "", map[string]any{
		"policy_name": req.Name,
	})
	return &pbauthv1.DeletePolicyResponse{}, nil
}

func (s *IAMServer) PutRoleTrustPolicy(
	ctx context.Context,
	req *pbauthv1.PutRoleTrustPolicyRequest,
) (*pbauthv1.PutRoleTrustPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.PutRoleTrustPolicy(ctx, iamdomain.PutRoleTrustPolicyInput{
		RoleName:   req.RoleName,
		Statements: iammapper.FromPBRoleTrustStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.PutRoleTrustPolicyResponse{}, nil
}

func (s *IAMServer) GetRoleTrustPolicy(
	ctx context.Context,
	req *pbauthv1.GetRoleTrustPolicyRequest,
) (*pbauthv1.GetRoleTrustPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.GetRoleTrustPolicy(ctx, req.RoleName)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.GetRoleTrustPolicyResponse{Statements: iammapper.ToPBRoleTrustStatements(items)}, nil
}

func (s *IAMServer) DeleteRoleTrustPolicy(
	ctx context.Context,
	req *pbauthv1.DeleteRoleTrustPolicyRequest,
) (*pbauthv1.DeleteRoleTrustPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DeleteRoleTrustPolicy(ctx, req.RoleName); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.DeleteRoleTrustPolicyResponse{}, nil
}

func (s *IAMServer) CreateGroup(
	ctx context.Context,
	req *pbauthv1.CreateGroupRequest,
) (*pbauthv1.CreateGroupResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if req.Scope == iamdomain.PolicyScopePlatform {
		if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
			return nil, iamStatusError(err)
		}
	} else if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	group, err := s.iamUC.CreateGroup(ctx, iamdomain.CreateGroupInput{
		Scope:       req.Scope,
		TenantID:    req.TenantId,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.CreateGroupResponse{Group: iammapper.ToPBGroup(group)}, nil
}

func (s *IAMServer) ListGroups(
	ctx context.Context,
	req *pbauthv1.ListGroupsRequest,
) (*pbauthv1.ListGroupsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if req.Scope == iamdomain.PolicyScopePlatform {
		if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
			return nil, iamStatusError(err)
		}
	} else if strings.TrimSpace(req.TenantId) != "" {
		if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
			return nil, iamStatusError(err)
		}
	}
	items, err := s.iamUC.ListGroups(ctx, req.Scope, req.TenantId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbauthv1.Group, 0, len(items))
	for i := range items {
		out = append(out, iammapper.ToPBGroup(&items[i]))
	}
	return &pbauthv1.ListGroupsResponse{Groups: out}, nil
}

func (s *IAMServer) DeleteGroup(
	ctx context.Context,
	req *pbauthv1.DeleteGroupRequest,
) (*pbauthv1.DeleteGroupResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DeleteGroup(ctx, req.GroupId); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"iam.group.deleted",
		"iam_group",
		fmt.Sprintf("%d", req.GroupId),
		"",
		map[string]any{
			"group_id": req.GroupId,
		},
	)
	return &pbauthv1.DeleteGroupResponse{}, nil
}

func (s *IAMServer) AddGroupMember(
	ctx context.Context,
	req *pbauthv1.AddGroupMemberRequest,
) (*pbauthv1.AddGroupMemberResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.AddGroupMember(ctx, req.GroupId, userID); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.AddGroupMemberResponse{}, nil
}

func (s *IAMServer) ListGroupMembers(
	ctx context.Context,
	req *pbauthv1.ListGroupMembersRequest,
) (*pbauthv1.ListGroupMembersResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListGroupMembers(ctx, req.GroupId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]uint64, 0, len(items))
	for _, item := range items {
		out = append(out, uint64(item))
	}
	return &pbauthv1.ListGroupMembersResponse{UserIds: out}, nil
}

func (s *IAMServer) RemoveGroupMember(
	ctx context.Context,
	req *pbauthv1.RemoveGroupMemberRequest,
) (*pbauthv1.RemoveGroupMemberResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RemoveGroupMember(ctx, req.GroupId, userID); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.RemoveGroupMemberResponse{}, nil
}

func (s *IAMServer) AttachGroupPolicy(
	ctx context.Context,
	req *pbauthv1.AttachGroupPolicyRequest,
) (*pbauthv1.AttachGroupPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.AttachGroupPolicy(ctx, req.GroupId, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.AttachGroupPolicyResponse{}, nil
}

func (s *IAMServer) ListGroupPolicies(
	ctx context.Context,
	req *pbauthv1.ListGroupPoliciesRequest,
) (*pbauthv1.ListGroupPoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListGroupPolicies(ctx, req.GroupId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.ListGroupPoliciesResponse{Policies: iammapper.ToPBPolicies(items)}, nil
}

func (s *IAMServer) DetachGroupPolicy(
	ctx context.Context,
	req *pbauthv1.DetachGroupPolicyRequest,
) (*pbauthv1.DetachGroupPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DetachGroupPolicy(ctx, req.GroupId, req.PolicyName); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.DetachGroupPolicyResponse{}, nil
}

func (s *IAMServer) PutGroupInlinePolicy(
	ctx context.Context,
	req *pbauthv1.PutGroupInlinePolicyRequest,
) (*pbauthv1.PutGroupInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.PutGroupInlinePolicy(ctx, iamdomain.PutGroupInlinePolicyInput{
		GroupID:     req.GroupId,
		Name:        req.Name,
		Description: req.Description,
		Statements:  iammapper.FromPBPolicyStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.PutGroupInlinePolicyResponse{}, nil
}

func (s *IAMServer) GetGroupInlinePolicy(
	ctx context.Context,
	req *pbauthv1.GetGroupInlinePolicyRequest,
) (*pbauthv1.GetGroupInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	item, err := s.iamUC.GetGroupInlinePolicy(ctx, req.GroupId, req.Name)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.GetGroupInlinePolicyResponse{Policy: iammapper.ToPBGroupInlinePolicy(item)}, nil
}

func (s *IAMServer) ListGroupInlinePolicies(
	ctx context.Context,
	req *pbauthv1.ListGroupInlinePoliciesRequest,
) (*pbauthv1.ListGroupInlinePoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListGroupInlinePolicies(ctx, req.GroupId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbauthv1.GroupInlinePolicy, 0, len(items))
	for i := range items {
		out = append(out, iammapper.ToPBGroupInlinePolicy(&items[i]))
	}
	return &pbauthv1.ListGroupInlinePoliciesResponse{Policies: out}, nil
}

func (s *IAMServer) DeleteGroupInlinePolicy(
	ctx context.Context,
	req *pbauthv1.DeleteGroupInlinePolicyRequest,
) (*pbauthv1.DeleteGroupInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DeleteGroupInlinePolicy(ctx, req.GroupId, req.Name); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.DeleteGroupInlinePolicyResponse{}, nil
}

func (s *IAMServer) ListPlatformUserPolicies(
	ctx context.Context,
	req *pbauthv1.ListPlatformUserPoliciesRequest,
) (*pbauthv1.ListPlatformUserPoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListPlatformUserPolicies(ctx, targetUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.ListPlatformUserPoliciesResponse{Policies: iammapper.ToPBPolicies(items)}, nil
}

func (s *IAMServer) PutPlatformUserInlinePolicy(
	ctx context.Context,
	req *pbauthv1.PutPlatformUserInlinePolicyRequest,
) (*pbauthv1.PutPlatformUserInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.PutPlatformUserInlinePolicy(ctx, iamdomain.PutPlatformUserInlinePolicyInput{
		UserID:      targetUserID,
		Name:        req.Name,
		Description: req.Description,
		Statements:  iammapper.FromPBPolicyStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.PutPlatformUserInlinePolicyResponse{}, nil
}

func (s *IAMServer) GetPlatformUserInlinePolicy(
	ctx context.Context,
	req *pbauthv1.GetPlatformUserInlinePolicyRequest,
) (*pbauthv1.GetPlatformUserInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	item, err := s.iamUC.GetPlatformUserInlinePolicy(ctx, targetUserID, req.Name)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.GetPlatformUserInlinePolicyResponse{Policy: iammapper.ToPBUserInlinePolicy(item)}, nil
}

func (s *IAMServer) ListPlatformUserInlinePolicies(
	ctx context.Context,
	req *pbauthv1.ListPlatformUserInlinePoliciesRequest,
) (*pbauthv1.ListPlatformUserInlinePoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListPlatformUserInlinePolicies(ctx, targetUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbauthv1.UserInlinePolicy, 0, len(items))
	for i := range items {
		out = append(out, iammapper.ToPBUserInlinePolicy(&items[i]))
	}
	return &pbauthv1.ListPlatformUserInlinePoliciesResponse{Policies: out}, nil
}

func (s *IAMServer) DeletePlatformUserInlinePolicy(
	ctx context.Context,
	req *pbauthv1.DeletePlatformUserInlinePolicyRequest,
) (*pbauthv1.DeletePlatformUserInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DeletePlatformUserInlinePolicy(ctx, targetUserID, req.Name); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.DeletePlatformUserInlinePolicyResponse{}, nil
}

func (s *IAMServer) AttachPlatformUserPolicy(
	ctx context.Context,
	req *pbauthv1.AttachPlatformUserPolicyRequest,
) (*pbauthv1.AttachPlatformUserPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.AttachPlatformUserPolicy(ctx, targetUserID, req.PolicyName); err != nil {
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
	return &pbauthv1.AttachPlatformUserPolicyResponse{}, nil
}

func (s *IAMServer) DetachPlatformUserPolicy(
	ctx context.Context,
	req *pbauthv1.DetachPlatformUserPolicyRequest,
) (*pbauthv1.DetachPlatformUserPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DetachPlatformUserPolicy(ctx, targetUserID, req.PolicyName); err != nil {
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
	return &pbauthv1.DetachPlatformUserPolicyResponse{}, nil
}

func (s *IAMServer) PutPlatformUserPermissionBoundary(
	ctx context.Context,
	req *pbauthv1.PutPlatformUserPermissionBoundaryRequest,
) (*pbauthv1.PutPlatformUserPermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.PutPlatformUserPermissionBoundary(ctx, targetUserID, req.PolicyName); err != nil {
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
	return &pbauthv1.PutPlatformUserPermissionBoundaryResponse{}, nil
}

func (s *IAMServer) GetPlatformUserPermissionBoundary(
	ctx context.Context,
	req *pbauthv1.GetPlatformUserPermissionBoundaryRequest,
) (*pbauthv1.GetPlatformUserPermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	item, err := s.iamUC.GetPlatformUserPermissionBoundary(ctx, targetUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.GetPlatformUserPermissionBoundaryResponse{Boundary: iammapper.ToPBPermissionBoundary(item)}, nil
}

func (s *IAMServer) DeletePlatformUserPermissionBoundary(
	ctx context.Context,
	req *pbauthv1.DeletePlatformUserPermissionBoundaryRequest,
) (*pbauthv1.DeletePlatformUserPermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DeletePlatformUserPermissionBoundary(ctx, targetUserID); err != nil {
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
	return &pbauthv1.DeletePlatformUserPermissionBoundaryResponse{}, nil
}

func (s *IAMServer) RemovePlatformRole(
	ctx context.Context,
	req *pbauthv1.RemovePlatformRoleRequest,
) (*pbauthv1.RemovePlatformRoleResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	targetUserID, err := toUint(req.TargetUserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.RemovePlatformRole(ctx, targetUserID, req.RoleName); err != nil {
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
	return &pbauthv1.RemovePlatformRoleResponse{}, nil
}

func (s *IAMServer) ListTenantMembers(
	ctx context.Context,
	req *pbauthv1.ListTenantMembersRequest,
) (*pbauthv1.ListTenantMembersResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListTenantMembers(ctx, req.TenantId)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbauthv1.TenantMembership, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, iammapper.ToPBMembership(&item))
	}
	return &pbauthv1.ListTenantMembersResponse{Memberships: out}, nil
}

func (s *IAMServer) RemoveTenantMember(
	ctx context.Context,
	req *pbauthv1.RemoveTenantMemberRequest,
) (*pbauthv1.RemoveTenantMemberResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RemoveMember(ctx, req.TenantId, userID); err != nil {
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
	return &pbauthv1.RemoveTenantMemberResponse{}, nil
}

func (s *IAMServer) ListTenantUserPolicies(
	ctx context.Context,
	req *pbauthv1.ListTenantUserPoliciesRequest,
) (*pbauthv1.ListTenantUserPoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListTenantUserPolicies(ctx, req.TenantId, userID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.ListTenantUserPoliciesResponse{Policies: iammapper.ToPBPolicies(items)}, nil
}

func (s *IAMServer) PutTenantUserInlinePolicy(
	ctx context.Context,
	req *pbauthv1.PutTenantUserInlinePolicyRequest,
) (*pbauthv1.PutTenantUserInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.PutTenantUserInlinePolicy(ctx, iamdomain.PutTenantUserInlinePolicyInput{
		TenantID:    req.TenantId,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Statements:  iammapper.FromPBPolicyStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.PutTenantUserInlinePolicyResponse{}, nil
}

func (s *IAMServer) GetTenantUserInlinePolicy(
	ctx context.Context,
	req *pbauthv1.GetTenantUserInlinePolicyRequest,
) (*pbauthv1.GetTenantUserInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	item, err := s.iamUC.GetTenantUserInlinePolicy(ctx, req.TenantId, userID, req.Name)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.GetTenantUserInlinePolicyResponse{Policy: iammapper.ToPBUserInlinePolicy(item)}, nil
}

func (s *IAMServer) ListTenantUserInlinePolicies(
	ctx context.Context,
	req *pbauthv1.ListTenantUserInlinePoliciesRequest,
) (*pbauthv1.ListTenantUserInlinePoliciesResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	items, err := s.iamUC.ListTenantUserInlinePolicies(ctx, req.TenantId, userID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	out := make([]*pbauthv1.UserInlinePolicy, 0, len(items))
	for i := range items {
		out = append(out, iammapper.ToPBUserInlinePolicy(&items[i]))
	}
	return &pbauthv1.ListTenantUserInlinePoliciesResponse{Policies: out}, nil
}

func (s *IAMServer) DeleteTenantUserInlinePolicy(
	ctx context.Context,
	req *pbauthv1.DeleteTenantUserInlinePolicyRequest,
) (*pbauthv1.DeleteTenantUserInlinePolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DeleteTenantUserInlinePolicy(ctx, req.TenantId, userID, req.Name); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.DeleteTenantUserInlinePolicyResponse{}, nil
}

func (s *IAMServer) AttachTenantUserPolicy(
	ctx context.Context,
	req *pbauthv1.AttachTenantUserPolicyRequest,
) (*pbauthv1.AttachTenantUserPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.AttachTenantUserPolicy(ctx, req.TenantId, userID, req.PolicyName); err != nil {
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
	return &pbauthv1.AttachTenantUserPolicyResponse{}, nil
}

func (s *IAMServer) DetachTenantUserPolicy(
	ctx context.Context,
	req *pbauthv1.DetachTenantUserPolicyRequest,
) (*pbauthv1.DetachTenantUserPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DetachTenantUserPolicy(ctx, req.TenantId, userID, req.PolicyName); err != nil {
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
	return &pbauthv1.DetachTenantUserPolicyResponse{}, nil
}

func (s *IAMServer) PutTenantUserPermissionBoundary(
	ctx context.Context,
	req *pbauthv1.PutTenantUserPermissionBoundaryRequest,
) (*pbauthv1.PutTenantUserPermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.PutTenantUserPermissionBoundary(ctx, req.TenantId, userID, req.PolicyName); err != nil {
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
	return &pbauthv1.PutTenantUserPermissionBoundaryResponse{}, nil
}

func (s *IAMServer) GetTenantUserPermissionBoundary(
	ctx context.Context,
	req *pbauthv1.GetTenantUserPermissionBoundaryRequest,
) (*pbauthv1.GetTenantUserPermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	item, err := s.iamUC.GetTenantUserPermissionBoundary(ctx, req.TenantId, userID)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.GetTenantUserPermissionBoundaryResponse{Boundary: iammapper.ToPBPermissionBoundary(item)}, nil
}

func (s *IAMServer) DeleteTenantUserPermissionBoundary(
	ctx context.Context,
	req *pbauthv1.DeleteTenantUserPermissionBoundaryRequest,
) (*pbauthv1.DeleteTenantUserPermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, req.TenantId, actorUserID, "tenant:manage_members"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DeleteTenantUserPermissionBoundary(ctx, req.TenantId, userID); err != nil {
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
	return &pbauthv1.DeleteTenantUserPermissionBoundaryResponse{}, nil
}

func (s *IAMServer) PutRolePermissionBoundary(
	ctx context.Context,
	req *pbauthv1.PutRolePermissionBoundaryRequest,
) (*pbauthv1.PutRolePermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.PutRolePermissionBoundary(ctx, req.RoleName, req.PolicyName); err != nil {
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
	return &pbauthv1.PutRolePermissionBoundaryResponse{}, nil
}

func (s *IAMServer) GetRolePermissionBoundary(
	ctx context.Context,
	req *pbauthv1.GetRolePermissionBoundaryRequest,
) (*pbauthv1.GetRolePermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	item, err := s.iamUC.GetRolePermissionBoundary(ctx, req.RoleName)
	if err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.GetRolePermissionBoundaryResponse{Boundary: iammapper.ToPBRolePermissionBoundary(item)}, nil
}

func (s *IAMServer) DeleteRolePermissionBoundary(
	ctx context.Context,
	req *pbauthv1.DeleteRolePermissionBoundaryRequest,
) (*pbauthv1.DeleteRolePermissionBoundaryResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	if err := s.iamUC.DeleteRolePermissionBoundary(ctx, req.RoleName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "iam.role_boundary.deleted", "iam_role_permission_boundary", req.RoleName, "", nil)
	return &pbauthv1.DeleteRolePermissionBoundaryResponse{}, nil
}

func (s *IAMServer) SimulateAccess(
	ctx context.Context,
	req *pbauthv1.SimulateAccessRequest,
) (*pbauthv1.SimulateAccessResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if err := s.iamUC.RequirePlatformPermission(ctx, actorUserID, "platform:manage_roles"); err != nil {
		return nil, iamStatusError(err)
	}
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	var assumedRole *iamdomain.AssumedRole
	if req.UseAssumedRole && req.AssumedRoleSession != nil && req.AssumedRoleSession.AssumedRoleId != 0 {
		assumedRole = &iamdomain.AssumedRole{
			RoleID:           req.AssumedRoleSession.AssumedRoleId,
			RoleScope:        req.AssumedRoleSession.AssumedRoleScope,
			RoleName:         req.AssumedRoleSession.AssumedRoleName,
			TenantID:         req.AssumedRoleSession.AssumedRoleTenantId,
			ServicePrincipal: req.AssumedRoleSession.AssumedRoleServicePrincipal,
			SessionName:      req.AssumedRoleSession.AssumedRoleSessionName,
			SourceIdentity:   req.AssumedRoleSession.AssumedRoleSourceIdentity,
			SessionTags:      req.AssumedRoleSession.SessionTags,
		}
	}
	result, err := s.iamUC.SimulateAccess(ctx, iamdomain.SimulateAccessInput{
		Scope:            req.Scope,
		TenantID:         req.TenantId,
		UserID:           userID,
		Action:           req.Action,
		Resource:         req.Resource,
		UseAssumedRole:   req.UseAssumedRole,
		AssumedRole:      assumedRole,
		SessionPolicy:    iammapper.FromPBPolicyStatements(req.SessionPolicy),
		Attributes:       mergeStringMaps(req.Attributes, toPrincipalTagAttributes(req.SessionTags)),
		ServicePrincipal: req.ServicePrincipal,
	})
	if err != nil {
		return nil, iamStatusError(err)
	}
	return iammapper.ToPBSimulateAccessResponse(result), nil
}

func toUint(v uint64) (uint, error) {
	if v == 0 {
		return 0, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if v > math.MaxUint {
		return 0, status.Error(codes.InvalidArgument, "user_id is out of range")
	}
	return uint(v), nil
}

func mergeStringMaps(items ...map[string]string) map[string]string {
	merged := make(map[string]string)
	for _, item := range items {
		for k, v := range item {
			merged[k] = v
		}
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func toPrincipalTagAttributes(tags map[string]string) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	out := make(map[string]string, len(tags)*2)
	for k, v := range tags {
		out["principal_tag:"+k] = v
		out["request_tag:"+k] = v
	}
	return out
}

func inviteAcceptBaseURL(appRedirectURL string) string {
	base := strings.TrimSpace(appRedirectURL)
	if base == "" {
		return ""
	}
	base = strings.TrimRight(base, "/")
	base = strings.TrimSuffix(base, "/auth/google/callback")
	return base
}

func (s *IAMServer) recordAudit(
	ctx context.Context,
	actorUserID uint,
	action string,
	resourceType string,
	resourceID string,
	tenantID string,
	payload map[string]any,
) {
	if s.auditRep == nil {
		return
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		payloadJSON = []byte("{}")
	}
	_ = s.auditRep.Create(ctx, iamdomain.AuditLog{
		ID:           uuid.NewString(),
		ActorUserID:  actorUserID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		TenantID:     tenantID,
		Status:       "success",
		PayloadJSON:  string(payloadJSON),
		CreatedAt:    time.Now().UTC(),
	})
}

func iamStatusError(err error) error {
	switch {
	case errors.Is(err, iamdomain.ErrInvalidTenantName),
		errors.Is(err, iamdomain.ErrInvalidTenantSlug),
		errors.Is(err, iamdomain.ErrInvalidOrganizationName),
		errors.Is(err, iamdomain.ErrInvalidOrganizationSlug),
		errors.Is(err, iamdomain.ErrInvalidUserID),
		errors.Is(err, iamdomain.ErrInvalidRoleName),
		errors.Is(err, iamdomain.ErrInvalidPolicyName),
		errors.Is(err, iamdomain.ErrInvalidAssumeRole),
		errors.Is(err, iamdomain.ErrInvalidServicePrincipal),
		errors.Is(err, iamdomain.ErrInvalidPolicyStatement):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, iamdomain.ErrTenantNotFound),
		errors.Is(err, iamdomain.ErrOrganizationNotFound),
		errors.Is(err, iamdomain.ErrMembershipNotFound),
		errors.Is(err, iamdomain.ErrRoleNotFound),
		errors.Is(err, iamdomain.ErrPolicyNotFound),
		errors.Is(err, iamdomain.ErrPolicyVersionNotFound),
		errors.Is(err, iamdomain.ErrGroupNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, iamdomain.ErrTenantSlugTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, iamdomain.ErrPermissionDenied),
		errors.Is(err, iamdomain.ErrAssumeRoleDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, iamdomain.ErrInactiveMembership):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, iamdomain.ErrImmutablePolicy),
		errors.Is(err, iamdomain.ErrImmutableGroup),
		errors.Is(err, iamdomain.ErrPolicyInUse),
		errors.Is(err, iamdomain.ErrDefaultPolicyVersion):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *IAMServer) authorizedContext(ctx context.Context) (context.Context, uint, error) {
	claims, err := s.claimsFromContext(ctx)
	if err != nil {
		return ctx, 0, err
	}
	ctx = iamdomain.WithSessionPolicyStatements(ctx, iammapper.ToIAMSessionPolicyStatements(claims.SessionPolicy))
	ctx = iamdomain.WithSessionTags(ctx, claims.SessionTags)
	if claims.AssumedRoleID != 0 && claims.AssumedRoleName != "" {
		ctx = iamdomain.WithAssumedRole(ctx, iamdomain.AssumedRole{
			RoleID:           claims.AssumedRoleID,
			RoleScope:        claims.AssumedRoleScope,
			RoleName:         claims.AssumedRoleName,
			TenantID:         claims.AssumedRoleTenantID,
			ServicePrincipal: claims.AssumedRoleServicePrincipal,
			SessionName:      claims.AssumedRoleSessionName,
			SourceIdentity:   claims.AssumedRoleSourceIdentity,
			SessionTags:      claims.SessionTags,
		})
	}
	if claims.UserID == 0 {
		return ctx, 0, errors.New("access token missing user_id")
	}
	return ctx, claims.UserID, nil
}

func (s *IAMServer) claimsFromContext(ctx context.Context) (*pdauthn.Claims, error) {
	return s.verifier.ClaimsFromContext(ctx)
}

func (s *IAMServer) claimsFromAccessToken(accessToken string) (*pdauthn.Claims, error) {
	return s.verifier.ClaimsFromAccessToken(accessToken)
}
