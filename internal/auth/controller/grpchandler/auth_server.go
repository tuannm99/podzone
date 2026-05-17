package grpchandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

type AuthServer struct {
	pbauthv1.UnimplementedAuthServiceServer
	pbauthv1.UnimplementedIAMServiceServer
	authUC         inputport.AuthUsecase
	iamUC          iamdomain.IAMUsecase
	sessionRep     outputport.SessionRepository
	auditRep       outputport.AuditLogRepository
	userRepo       outputport.UserRepository
	jwtSecret      string
	jwtKey         string
	appRedirectURL string
}

func NewAuthServer(
	authUC inputport.AuthUsecase,
	sessionRep outputport.SessionRepository,
	auditRep outputport.AuditLogRepository,
	userRepo outputport.UserRepository,
	cfg config.AuthConfig,
) *AuthServer {
	return &AuthServer{
		authUC:         authUC,
		sessionRep:     sessionRep,
		auditRep:       auditRep,
		userRepo:       userRepo,
		jwtSecret:      cfg.JWTSecret,
		jwtKey:         cfg.JWTKey,
		appRedirectURL: cfg.AppRedirectURL,
	}
}

func NewIAMServer(
	iamUC iamdomain.IAMUsecase,
	auditRep outputport.AuditLogRepository,
	userRepo outputport.UserRepository,
	cfg config.AuthConfig,
) *AuthServer {
	return &AuthServer{
		iamUC:          iamUC,
		auditRep:       auditRep,
		userRepo:       userRepo,
		jwtSecret:      cfg.JWTSecret,
		jwtKey:         cfg.JWTKey,
		appRedirectURL: cfg.AppRedirectURL,
	}
}

func (s *AuthServer) GoogleLogin(
	ctx context.Context,
	req *pbauthv1.GoogleLoginRequest,
) (*pbauthv1.GoogleLoginResponse, error) {
	authURL, err := s.authUC.GenerateOAuthURL(ctx)
	if err != nil {
		return nil, err
	}
	return &pbauthv1.GoogleLoginResponse{
		RedirectUrl: authURL,
	}, nil
}

func (s *AuthServer) GoogleCallback(
	ctx context.Context,
	req *pbauthv1.GoogleCallbackRequest,
) (*pbauthv1.GoogleCallbackResponse, error) {
	callbackResp, err := s.authUC.HandleOAuthCallback(ctx, req.Code, req.State)
	if err != nil {
		return nil, err
	}
	resp, err := toolkit.MapStruct[inputport.GoogleCallbackResult, pbauthv1.GoogleCallbackResponse](*callbackResp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *AuthServer) ExchangeGoogleLogin(
	ctx context.Context,
	req *pbauthv1.ExchangeGoogleLoginRequest,
) (*pbauthv1.LoginResponse, error) {
	authResp, err := s.authUC.ExchangeOAuthLogin(ctx, req.ExchangeCode)
	if err != nil {
		return nil, authStatusError(err)
	}
	resp, err := toolkit.MapStruct[inputport.AuthResult, pbauthv1.LoginResponse](*authResp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pbauthv1.LogoutRequest) (*pbauthv1.LogoutResponse, error) {
	redirectUrl, err := s.authUC.Logout(ctx, req.Token)
	if err != nil {
		return nil, authStatusError(err)
	}
	return &pbauthv1.LogoutResponse{
		Success:     true,
		RedirectUrl: redirectUrl,
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pbauthv1.LoginRequest) (*pbauthv1.LoginResponse, error) {
	loginResp, err := s.authUC.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	resp, err := toolkit.MapStruct[inputport.AuthResult, pbauthv1.LoginResponse](*loginResp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *AuthServer) Register(ctx context.Context, req *pbauthv1.RegisterRequest) (*pbauthv1.RegisterResponse, error) {
	registerCmd, err := toolkit.MapStruct[*pbauthv1.RegisterRequest, inputport.RegisterCmd](req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	registerResp, err := s.authUC.Register(ctx, *registerCmd)
	if err != nil {
		return nil, err
	}
	resp, err := toolkit.MapStruct[inputport.AuthResult, pbauthv1.RegisterResponse](*registerResp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *AuthServer) CreateTenant(
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
		Tenant:          toProtoTenant(tenant),
		OwnerMembership: toProtoMembership(membership),
	}, nil
}

func (s *AuthServer) AddTenantMember(
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

func (s *AuthServer) AddTenantMemberByIdentity(
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
	if s.userRepo == nil {
		return nil, status.Error(codes.Internal, "user repository is not configured")
	}

	createdUser := false
	user, err := s.userRepo.GetByUsernameOrEmail(identity)
	if err != nil {
		if strings.Contains(identity, "@") && errors.Is(err, entity.ErrUserNotFound) {
			user, err = s.userRepo.CreateByEmailIfNotExisted(identity)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			createdUser = true
		} else if errors.Is(err, entity.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	if user == nil || user.Id == 0 {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if err := s.iamUC.AddMember(ctx, req.TenantId, user.Id, req.RoleName); err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(
		ctx,
		actorUserID,
		"tenant.member.identity_added",
		"tenant_member",
		fmt.Sprintf("%s:%d", req.TenantId, user.Id),
		req.TenantId,
		map[string]any{
			"user_id":      user.Id,
			"identity":     identity,
			"role_name":    req.RoleName,
			"created_user": createdUser,
		},
	)
	return &pbauthv1.AddTenantMemberByIdentityResponse{
		UserId:      uint64(user.Id),
		CreatedUser: createdUser,
	}, nil
}

func (s *AuthServer) CreateTenantInvite(
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
		Invite:      toProtoInvite(invite),
		InviteToken: rawToken,
		AcceptUrl:   acceptURL,
	}, nil
}

func (s *AuthServer) ListTenantInvites(
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
		out = append(out, toProtoInvite(&items[i]))
	}
	return &pbauthv1.ListTenantInvitesResponse{Invites: out}, nil
}

func (s *AuthServer) RevokeTenantInvite(
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

func (s *AuthServer) AcceptTenantInvite(
	ctx context.Context,
	req *pbauthv1.AcceptTenantInviteRequest,
) (*pbauthv1.AcceptTenantInviteResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if s.userRepo == nil {
		return nil, status.Error(codes.Internal, "user repository is not configured")
	}
	user, err := s.userRepo.GetByID(fmt.Sprintf("%d", actorUserID))
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
		Membership: toProtoMembership(membership),
	}, nil
}

func (s *AuthServer) GetTenantMembership(
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
		Membership: toProtoMembership(membership),
	}, nil
}

func (s *AuthServer) CheckPermission(
	ctx context.Context,
	req *pbauthv1.CheckPermissionRequest,
) (*pbauthv1.CheckPermissionResponse, error) {
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}

	allowed, err := s.iamUC.CheckPermission(ctx, req.TenantId, userID, req.Permission)
	if err != nil {
		if errors.Is(err, iamdomain.ErrPermissionDenied) || errors.Is(err, iamdomain.ErrInactiveMembership) {
			return &pbauthv1.CheckPermissionResponse{Allowed: false}, nil
		}
		return nil, iamStatusError(err)
	}
	return &pbauthv1.CheckPermissionResponse{Allowed: allowed}, nil
}

func (s *AuthServer) CheckPlatformPermission(
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

func (s *AuthServer) SwitchActiveTenant(
	ctx context.Context,
	req *pbauthv1.SwitchActiveTenantRequest,
) (*pbauthv1.SwitchActiveTenantResponse, error) {
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	if userID != actorUserID {
		return nil, status.Error(codes.PermissionDenied, "cannot switch another user's active tenant")
	}

	authResp, err := s.authUC.SwitchActiveTenant(ctx, userID, req.TenantId, req.AccessToken)
	if err != nil {
		return nil, authStatusError(err)
	}

	resp, err := toolkit.MapStruct[inputport.AuthResult, pbauthv1.SwitchActiveTenantResponse](*authResp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.recordAudit(ctx, actorUserID, "session.tenant_switched", "session", "", req.TenantId, map[string]any{
		"tenant_id": req.TenantId,
	})
	return resp, nil
}

func (s *AuthServer) AssumeSessionPolicy(
	ctx context.Context,
	req *pbauthv1.AssumeSessionPolicyRequest,
) (*pbauthv1.AssumeSessionPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	authResp, err := s.authUC.AssumeSessionPolicy(
		ctx,
		actorUserID,
		req.AccessToken,
		fromProtoSessionPolicyStatements(req.Statements),
	)
	if err != nil {
		return nil, authStatusError(err)
	}
	session, err := s.currentSessionFromAccessToken(req.AccessToken)
	if err != nil {
		return nil, authStatusError(err)
	}
	return &pbauthv1.AssumeSessionPolicyResponse{
		JwtToken: authResp.JwtToken,
		Session:  toProtoSession(session),
	}, nil
}

func (s *AuthServer) ClearSessionPolicy(
	ctx context.Context,
	req *pbauthv1.ClearSessionPolicyRequest,
) (*pbauthv1.ClearSessionPolicyResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	authResp, err := s.authUC.ClearSessionPolicy(ctx, actorUserID, req.AccessToken)
	if err != nil {
		return nil, authStatusError(err)
	}
	session, err := s.currentSessionFromAccessToken(req.AccessToken)
	if err != nil {
		return nil, authStatusError(err)
	}
	return &pbauthv1.ClearSessionPolicyResponse{
		JwtToken: authResp.JwtToken,
		Session:  toProtoSession(session),
	}, nil
}

func (s *AuthServer) AssumeRole(
	ctx context.Context,
	req *pbauthv1.AssumeRoleRequest,
) (*pbauthv1.AssumeRoleResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	authResp, err := s.authUC.AssumeRole(
		ctx,
		actorUserID,
		req.AccessToken,
		req.RoleName,
		req.TenantId,
		fromProtoSessionPolicyStatements(req.SessionPolicy),
	)
	if err != nil {
		return nil, authStatusError(err)
	}
	session, err := s.currentSessionFromAccessToken(req.AccessToken)
	if err != nil {
		return nil, authStatusError(err)
	}
	return &pbauthv1.AssumeRoleResponse{
		JwtToken: authResp.JwtToken,
		Session:  toProtoSession(session),
	}, nil
}

func (s *AuthServer) ClearAssumedRole(
	ctx context.Context,
	req *pbauthv1.ClearAssumedRoleRequest,
) (*pbauthv1.ClearAssumedRoleResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	authResp, err := s.authUC.ClearAssumedRole(ctx, actorUserID, req.AccessToken)
	if err != nil {
		return nil, authStatusError(err)
	}
	session, err := s.currentSessionFromAccessToken(req.AccessToken)
	if err != nil {
		return nil, authStatusError(err)
	}
	return &pbauthv1.ClearAssumedRoleResponse{
		JwtToken: authResp.JwtToken,
		Session:  toProtoSession(session),
	}, nil
}

func (s *AuthServer) GetSession(
	ctx context.Context,
	req *pbauthv1.GetSessionRequest,
) (*pbauthv1.GetSessionResponse, error) {
	session, err := s.sessionRep.GetByID(ctx, req.SessionId)
	if err != nil {
		return nil, authStatusError(err)
	}
	return &pbauthv1.GetSessionResponse{Session: toProtoSession(session)}, nil
}

func (s *AuthServer) ListSessions(
	ctx context.Context,
	req *pbauthv1.ListSessionsRequest,
) (*pbauthv1.ListSessionsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	items, err := s.sessionRep.ListByUser(ctx, actorUserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*pbauthv1.Session, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, toProtoSession(&item))
	}
	return &pbauthv1.ListSessionsResponse{Sessions: out}, nil
}

func (s *AuthServer) RevokeSession(
	ctx context.Context,
	req *pbauthv1.RevokeSessionRequest,
) (*pbauthv1.RevokeSessionResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	session, err := s.sessionRep.GetByID(ctx, req.SessionId)
	if err != nil {
		return nil, authStatusError(err)
	}
	if session.UserID != actorUserID {
		return nil, status.Error(codes.PermissionDenied, "cannot revoke another user's session")
	}
	now := time.Now().UTC()
	if err := s.sessionRep.Revoke(ctx, session.ID, now); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	s.recordAudit(ctx, actorUserID, "session.revoked", "session", session.ID, session.ActiveTenantID, map[string]any{
		"target_user_id":   session.UserID,
		"active_tenant_id": session.ActiveTenantID,
		"revoked_at":       now.Format(time.RFC3339),
	})
	return &pbauthv1.RevokeSessionResponse{}, nil
}

func (s *AuthServer) ListAuditLogs(
	ctx context.Context,
	req *pbauthv1.ListAuditLogsRequest,
) (*pbauthv1.ListAuditLogsResponse, error) {
	ctx, actorUserID, err := s.authorizedContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	limit := 20
	if req.PageSize > 0 {
		limit = int(req.PageSize)
	}
	if limit > 100 {
		limit = 100
	}
	items, err := s.auditRep.ListByActor(ctx, actorUserID, limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*pbauthv1.AuditLog, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, toProtoAuditLog(&item))
	}
	return &pbauthv1.ListAuditLogsResponse{Logs: out}, nil
}

func (s *AuthServer) RefreshToken(
	ctx context.Context,
	req *pbauthv1.RefreshTokenRequest,
) (*pbauthv1.RefreshTokenResponse, error) {
	authResp, err := s.authUC.RefreshAccessToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, authStatusError(err)
	}
	resp, err := toolkit.MapStruct[inputport.AuthResult, pbauthv1.RefreshTokenResponse](*authResp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *AuthServer) ListUserTenants(
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
		out = append(out, toProtoMembership(&item))
	}
	return &pbauthv1.ListUserTenantsResponse{Memberships: out}, nil
}

func (s *AuthServer) ListPlatformRoles(
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
		out = append(out, toProtoPlatformMembership(&item))
	}
	return &pbauthv1.ListPlatformRolesResponse{Memberships: out}, nil
}

func (s *AuthServer) AddPlatformRole(
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

func (s *AuthServer) CreatePolicy(
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
		Statements:  fromProtoPolicyStatements(req.Statements),
	})
	if err != nil {
		return nil, iamStatusError(err)
	}
	s.recordAudit(ctx, actorUserID, "iam.policy.created", "iam_policy", policy.Name, "", map[string]any{
		"scope":      policy.Scope,
		"statements": len(statements),
	})
	return &pbauthv1.CreatePolicyResponse{
		Policy:     toProtoPolicy(policy),
		Statements: toProtoPolicyStatements(statements),
	}, nil
}

func (s *AuthServer) GetPolicy(
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
		Policy:     toProtoPolicy(policy),
		Statements: toProtoPolicyStatements(statements),
	}, nil
}

func (s *AuthServer) ListPolicies(
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
	return &pbauthv1.ListPoliciesResponse{Policies: toProtoPolicies(items)}, nil
}

func (s *AuthServer) ListPolicyAttachments(
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
	return &pbauthv1.ListPolicyAttachmentsResponse{Attachments: toProtoPolicyAttachments(items)}, nil
}

func (s *AuthServer) DeletePolicy(
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

func (s *AuthServer) PutRoleTrustPolicy(
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
		Statements: fromProtoRoleTrustStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.PutRoleTrustPolicyResponse{}, nil
}

func (s *AuthServer) GetRoleTrustPolicy(
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
	return &pbauthv1.GetRoleTrustPolicyResponse{Statements: toProtoRoleTrustStatements(items)}, nil
}

func (s *AuthServer) DeleteRoleTrustPolicy(
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

func (s *AuthServer) CreateGroup(
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
	return &pbauthv1.CreateGroupResponse{Group: toProtoGroup(group)}, nil
}

func (s *AuthServer) ListGroups(
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
		out = append(out, toProtoGroup(&items[i]))
	}
	return &pbauthv1.ListGroupsResponse{Groups: out}, nil
}

func (s *AuthServer) DeleteGroup(
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
	s.recordAudit(ctx, actorUserID, "iam.group.deleted", "iam_group", fmt.Sprintf("%d", req.GroupId), "", map[string]any{
		"group_id": req.GroupId,
	})
	return &pbauthv1.DeleteGroupResponse{}, nil
}

func (s *AuthServer) AddGroupMember(
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

func (s *AuthServer) ListGroupMembers(
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

func (s *AuthServer) RemoveGroupMember(
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

func (s *AuthServer) AttachGroupPolicy(
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

func (s *AuthServer) ListGroupPolicies(
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
	return &pbauthv1.ListGroupPoliciesResponse{Policies: toProtoPolicies(items)}, nil
}

func (s *AuthServer) DetachGroupPolicy(
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

func (s *AuthServer) PutGroupInlinePolicy(
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
		Statements:  fromProtoPolicyStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.PutGroupInlinePolicyResponse{}, nil
}

func (s *AuthServer) GetGroupInlinePolicy(
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
	return &pbauthv1.GetGroupInlinePolicyResponse{Policy: toProtoGroupInlinePolicy(item)}, nil
}

func (s *AuthServer) ListGroupInlinePolicies(
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
		out = append(out, toProtoGroupInlinePolicy(&items[i]))
	}
	return &pbauthv1.ListGroupInlinePoliciesResponse{Policies: out}, nil
}

func (s *AuthServer) DeleteGroupInlinePolicy(
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

func (s *AuthServer) ListPlatformUserPolicies(
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
	return &pbauthv1.ListPlatformUserPoliciesResponse{Policies: toProtoPolicies(items)}, nil
}

func (s *AuthServer) PutPlatformUserInlinePolicy(
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
		Statements:  fromProtoPolicyStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.PutPlatformUserInlinePolicyResponse{}, nil
}

func (s *AuthServer) GetPlatformUserInlinePolicy(
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
	return &pbauthv1.GetPlatformUserInlinePolicyResponse{Policy: toProtoUserInlinePolicy(item)}, nil
}

func (s *AuthServer) ListPlatformUserInlinePolicies(
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
		out = append(out, toProtoUserInlinePolicy(&items[i]))
	}
	return &pbauthv1.ListPlatformUserInlinePoliciesResponse{Policies: out}, nil
}

func (s *AuthServer) DeletePlatformUserInlinePolicy(
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

func (s *AuthServer) AttachPlatformUserPolicy(
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
	s.recordAudit(ctx, actorUserID, "iam.platform_user_policy.attached", "iam_policy_attachment", req.PolicyName, "", map[string]any{
		"target_user_id": targetUserID,
		"policy_name":    req.PolicyName,
	})
	return &pbauthv1.AttachPlatformUserPolicyResponse{}, nil
}

func (s *AuthServer) DetachPlatformUserPolicy(
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
	s.recordAudit(ctx, actorUserID, "iam.platform_user_policy.detached", "iam_policy_attachment", req.PolicyName, "", map[string]any{
		"target_user_id": targetUserID,
		"policy_name":    req.PolicyName,
	})
	return &pbauthv1.DetachPlatformUserPolicyResponse{}, nil
}

func (s *AuthServer) RemovePlatformRole(
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

func (s *AuthServer) ListTenantMembers(
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
		out = append(out, toProtoMembership(&item))
	}
	return &pbauthv1.ListTenantMembersResponse{Memberships: out}, nil
}

func (s *AuthServer) RemoveTenantMember(
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

func (s *AuthServer) ListTenantUserPolicies(
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
	return &pbauthv1.ListTenantUserPoliciesResponse{Policies: toProtoPolicies(items)}, nil
}

func (s *AuthServer) PutTenantUserInlinePolicy(
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
		Statements:  fromProtoPolicyStatements(req.Statements),
	}); err != nil {
		return nil, iamStatusError(err)
	}
	return &pbauthv1.PutTenantUserInlinePolicyResponse{}, nil
}

func (s *AuthServer) GetTenantUserInlinePolicy(
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
	return &pbauthv1.GetTenantUserInlinePolicyResponse{Policy: toProtoUserInlinePolicy(item)}, nil
}

func (s *AuthServer) ListTenantUserInlinePolicies(
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
		out = append(out, toProtoUserInlinePolicy(&items[i]))
	}
	return &pbauthv1.ListTenantUserInlinePoliciesResponse{Policies: out}, nil
}

func (s *AuthServer) DeleteTenantUserInlinePolicy(
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

func (s *AuthServer) AttachTenantUserPolicy(
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
	s.recordAudit(ctx, actorUserID, "iam.tenant_user_policy.attached", "iam_policy_attachment", req.PolicyName, req.TenantId, map[string]any{
		"user_id":     userID,
		"tenant_id":   req.TenantId,
		"policy_name": req.PolicyName,
	})
	return &pbauthv1.AttachTenantUserPolicyResponse{}, nil
}

func (s *AuthServer) DetachTenantUserPolicy(
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
	s.recordAudit(ctx, actorUserID, "iam.tenant_user_policy.detached", "iam_policy_attachment", req.PolicyName, req.TenantId, map[string]any{
		"user_id":     userID,
		"tenant_id":   req.TenantId,
		"policy_name": req.PolicyName,
	})
	return &pbauthv1.DetachTenantUserPolicyResponse{}, nil
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

func toProtoTenant(t *iamdomain.Tenant) *pbauthv1.Tenant {
	if t == nil {
		return nil
	}

	return &pbauthv1.Tenant{
		Id:        t.ID,
		Slug:      t.Slug,
		Name:      t.Name,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
		UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
	}
}

func toProtoMembership(m *iamdomain.Membership) *pbauthv1.TenantMembership {
	if m == nil {
		return nil
	}
	return &pbauthv1.TenantMembership{
		TenantId:  m.TenantID,
		UserId:    uint64(m.UserID),
		RoleId:    m.RoleID,
		RoleName:  m.RoleName,
		Status:    m.Status,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

func toProtoPlatformMembership(m *iamdomain.PlatformMembership) *pbauthv1.PlatformRoleMembership {
	if m == nil {
		return nil
	}
	return &pbauthv1.PlatformRoleMembership{
		UserId:    uint64(m.UserID),
		RoleId:    m.RoleID,
		RoleName:  m.RoleName,
		Status:    m.Status,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339),
	}
}

func toProtoPolicy(policy *iamdomain.Policy) *pbauthv1.Policy {
	if policy == nil {
		return nil
	}
	return &pbauthv1.Policy{
		Id:          policy.ID,
		Scope:       policy.Scope,
		Name:        policy.Name,
		Description: policy.Description,
		IsSystem:    policy.IsSystem,
		CreatedAt:   policy.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   policy.UpdatedAt.Format(time.RFC3339),
	}
}

func toProtoGroup(group *iamdomain.Group) *pbauthv1.Group {
	if group == nil {
		return nil
	}
	return &pbauthv1.Group{
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

func toProtoGroupInlinePolicy(policy *iamdomain.GroupInlinePolicy) *pbauthv1.GroupInlinePolicy {
	if policy == nil {
		return nil
	}
	return &pbauthv1.GroupInlinePolicy{
		GroupId:     policy.GroupID,
		Name:        policy.Name,
		Description: policy.Description,
		Statements:  toProtoPolicyStatements(policy.Statements),
		CreatedAt:   policy.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   policy.UpdatedAt.Format(time.RFC3339),
	}
}

func toProtoUserInlinePolicy(policy *iamdomain.UserInlinePolicy) *pbauthv1.UserInlinePolicy {
	if policy == nil {
		return nil
	}
	return &pbauthv1.UserInlinePolicy{
		Scope:       policy.Scope,
		TenantId:    policy.TenantID,
		UserId:      uint64(policy.UserID),
		Name:        policy.Name,
		Description: policy.Description,
		Statements:  toProtoPolicyStatements(policy.Statements),
		CreatedAt:   policy.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   policy.UpdatedAt.Format(time.RFC3339),
	}
}

func toProtoPolicies(items []iamdomain.Policy) []*pbauthv1.Policy {
	out := make([]*pbauthv1.Policy, 0, len(items))
	for i := range items {
		out = append(out, toProtoPolicy(&items[i]))
	}
	return out
}

func toProtoPolicyAttachments(items []iamdomain.PolicyAttachment) []*pbauthv1.PolicyAttachment {
	out := make([]*pbauthv1.PolicyAttachment, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, &pbauthv1.PolicyAttachment{
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

func toProtoRoleTrustStatements(items []iamdomain.RoleTrustStatement) []*pbauthv1.RoleTrustStatement {
	out := make([]*pbauthv1.RoleTrustStatement, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, &pbauthv1.RoleTrustStatement{
			Id:               item.ID,
			RoleId:           item.RoleID,
			Effect:           item.Effect,
			PrincipalType:    item.PrincipalType,
			PrincipalPattern: item.PrincipalPattern,
			TenantPattern:    item.TenantPattern,
			CreatedAt:        item.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}

func toProtoPolicyStatements(items []iamdomain.PolicyStatement) []*pbauthv1.PolicyStatement {
	out := make([]*pbauthv1.PolicyStatement, 0, len(items))
	for i := range items {
		item := items[i]
		out = append(out, &pbauthv1.PolicyStatement{
			Id:              item.ID,
			PolicyId:        item.PolicyID,
			PolicyName:      item.PolicyName,
			Effect:          item.Effect,
			ActionPattern:   item.ActionPattern,
			ResourcePattern: item.ResourcePattern,
			CreatedAt:       item.CreatedAt.Format(time.RFC3339),
		})
	}
	return out
}

func fromProtoPolicyStatements(items []*pbauthv1.PolicyStatement) []iamdomain.PolicyStatement {
	out := make([]iamdomain.PolicyStatement, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, iamdomain.PolicyStatement{
			Effect:          item.Effect,
			ActionPattern:   item.ActionPattern,
			ResourcePattern: item.ResourcePattern,
		})
	}
	return out
}

func fromProtoRoleTrustStatements(items []*pbauthv1.RoleTrustStatement) []iamdomain.RoleTrustStatement {
	out := make([]iamdomain.RoleTrustStatement, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, iamdomain.RoleTrustStatement{
			Effect:           item.Effect,
			PrincipalType:    item.PrincipalType,
			PrincipalPattern: item.PrincipalPattern,
			TenantPattern:    item.TenantPattern,
		})
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

func toProtoInvite(invite *iamdomain.TenantInvite) *pbauthv1.TenantInvite {
	if invite == nil {
		return nil
	}
	resp := &pbauthv1.TenantInvite{
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

func toProtoSession(s *entity.Session) *pbauthv1.Session {
	if s == nil {
		return nil
	}
	resp := &pbauthv1.Session{
		Id:                  s.ID,
		UserId:              uint64(s.UserID),
		ActiveTenantId:      s.ActiveTenantID,
		Status:              s.Status,
		CreatedAt:           s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           s.UpdatedAt.Format(time.RFC3339),
		ExpiresAt:           s.ExpiresAt.Format(time.RFC3339),
		SessionPolicy:       toProtoSessionPolicyStatements(s.SessionPolicy),
		AssumedRoleId:       s.AssumedRoleID,
		AssumedRoleScope:    s.AssumedRoleScope,
		AssumedRoleName:     s.AssumedRoleName,
		AssumedRoleTenantId: s.AssumedRoleTenantID,
	}
	if s.RevokedAt != nil {
		resp.RevokedAt = s.RevokedAt.Format(time.RFC3339)
	}
	return resp
}

func toProtoAuditLog(a *entity.AuditLog) *pbauthv1.AuditLog {
	if a == nil {
		return nil
	}
	return &pbauthv1.AuditLog{
		Id:           a.ID,
		ActorUserId:  uint64(a.ActorUserID),
		Action:       a.Action,
		ResourceType: a.ResourceType,
		ResourceId:   a.ResourceID,
		TenantId:     a.TenantID,
		Status:       a.Status,
		PayloadJson:  a.PayloadJSON,
		CreatedAt:    a.CreatedAt.Format(time.RFC3339),
	}
}

func (s *AuthServer) recordAudit(
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
	_ = s.auditRep.Create(ctx, entity.AuditLog{
		ID:           uuid.NewString(),
		ActorUserID:  actorUserID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		TenantID:     tenantID,
		Status:       entity.AuditStatusSuccess,
		PayloadJSON:  string(payloadJSON),
		CreatedAt:    time.Now().UTC(),
	})
}

func authStatusError(err error) error {
	switch {
	case errors.Is(err, entity.ErrInvalidUserID),
		errors.Is(err, entity.ErrInvalidSessionPolicy):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, entity.ErrSessionNotFound),
		errors.Is(err, entity.ErrRefreshTokenInvalid):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, entity.ErrRefreshTokenExpired),
		errors.Is(err, entity.ErrSessionRevoked):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return iamStatusError(err)
	}
}

func iamStatusError(err error) error {
	switch {
	case errors.Is(err, iamdomain.ErrInvalidTenantName),
		errors.Is(err, iamdomain.ErrInvalidTenantSlug),
		errors.Is(err, iamdomain.ErrInvalidUserID),
		errors.Is(err, iamdomain.ErrInvalidRoleName),
		errors.Is(err, iamdomain.ErrInvalidPolicyName),
		errors.Is(err, iamdomain.ErrInvalidAssumeRole):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, iamdomain.ErrTenantNotFound),
		errors.Is(err, iamdomain.ErrMembershipNotFound),
		errors.Is(err, iamdomain.ErrRoleNotFound),
		errors.Is(err, iamdomain.ErrPolicyNotFound),
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
		errors.Is(err, iamdomain.ErrPolicyInUse):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *AuthServer) actorUserIDFromContext(ctx context.Context) (uint, error) {
	claims, err := s.claimsFromContext(ctx)
	if err != nil {
		return 0, err
	}
	if claims.UserID == 0 {
		return 0, errors.New("access token missing user_id")
	}
	return uint(claims.UserID), nil
}

func (s *AuthServer) authorizedContext(ctx context.Context) (context.Context, uint, error) {
	claims, err := s.claimsFromContext(ctx)
	if err != nil {
		return ctx, 0, err
	}
	ctx = iamdomain.WithSessionPolicyStatements(ctx, toIAMSessionPolicyStatements(claims.SessionPolicy))
	if claims.AssumedRoleID != 0 && claims.AssumedRoleName != "" {
		ctx = iamdomain.WithAssumedRole(ctx, iamdomain.AssumedRole{
			RoleID:    claims.AssumedRoleID,
			RoleScope: claims.AssumedRoleScope,
			RoleName:  claims.AssumedRoleName,
			TenantID:  claims.AssumedRoleTenantID,
		})
	}
	if claims.UserID == 0 {
		return ctx, 0, errors.New("access token missing user_id")
	}
	return ctx, uint(claims.UserID), nil
}

func (s *AuthServer) currentSessionFromAccessToken(accessToken string) (*entity.Session, error) {
	claims := &entity.JWTClaims{}
	token, err := jwt.ParseWithClaims(accessToken, claims, func(tok *jwt.Token) (interface{}, error) {
		if tok.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid access token")
	}
	if s.jwtKey != "" && claims.Key != s.jwtKey {
		return nil, errors.New("invalid access token")
	}
	return s.sessionRep.GetByID(context.Background(), claims.SessionID)
}

func (s *AuthServer) claimsFromContext(ctx context.Context) (*entity.JWTClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("missing request metadata")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, errors.New("missing authorization header")
	}
	raw := strings.TrimSpace(values[0])
	if !strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		return nil, errors.New("invalid authorization header")
	}
	tokenString := strings.TrimSpace(raw[len("Bearer "):])
	claims := &entity.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(tok *jwt.Token) (interface{}, error) {
		if tok.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid access token")
	}
	if s.jwtKey != "" && claims.Key != s.jwtKey {
		return nil, errors.New("invalid access token")
	}
	return claims, nil
}

func toProtoSessionPolicyStatements(items []entity.SessionPolicyStatement) []*pbauthv1.PolicyStatement {
	out := make([]*pbauthv1.PolicyStatement, 0, len(items))
	for _, item := range items {
		out = append(out, &pbauthv1.PolicyStatement{
			Effect:          item.Effect,
			ActionPattern:   item.ActionPattern,
			ResourcePattern: item.ResourcePattern,
		})
	}
	return out
}

func fromProtoSessionPolicyStatements(items []*pbauthv1.PolicyStatement) []entity.SessionPolicyStatement {
	out := make([]entity.SessionPolicyStatement, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, entity.SessionPolicyStatement{
			Effect:          item.Effect,
			ActionPattern:   item.ActionPattern,
			ResourcePattern: item.ResourcePattern,
		})
	}
	return out
}

func toIAMSessionPolicyStatements(items []entity.SessionPolicyStatement) []iamdomain.PolicyStatement {
	out := make([]iamdomain.PolicyStatement, 0, len(items))
	for _, item := range items {
		out = append(out, iamdomain.PolicyStatement{
			Effect:          item.Effect,
			ActionPattern:   item.ActionPattern,
			ResourcePattern: item.ResourcePattern,
		})
	}
	return out
}
