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
	authUC     inputport.AuthUsecase
	iamUC      iamdomain.IAMUsecase
	sessionRep outputport.SessionRepository
	auditRep   outputport.AuditLogRepository
	jwtSecret  string
	jwtKey     string
}

func NewAuthServer(authUC inputport.AuthUsecase, sessionRep outputport.SessionRepository, auditRep outputport.AuditLogRepository, cfg config.AuthConfig) *AuthServer {
	return &AuthServer{
		authUC:     authUC,
		sessionRep: sessionRep,
		auditRep:   auditRep,
		jwtSecret:  cfg.JWTSecret,
		jwtKey:     cfg.JWTKey,
	}
}

func NewIAMServer(iamUC iamdomain.IAMUsecase, auditRep outputport.AuditLogRepository, cfg config.AuthConfig) *AuthServer {
	return &AuthServer{
		iamUC:     iamUC,
		auditRep:  auditRep,
		jwtSecret: cfg.JWTSecret,
		jwtKey:    cfg.JWTKey,
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
	actorUserID, err := s.actorUserIDFromContext(ctx)
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
	actorUserID, err := s.actorUserIDFromContext(ctx)
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
	s.recordAudit(ctx, actorUserID, "tenant.member.added", "tenant_member", fmt.Sprintf("%s:%d", req.TenantId, req.UserId), req.TenantId, map[string]any{
		"user_id":   userID,
		"role_name": req.RoleName,
	})
	return &pbauthv1.AddTenantMemberResponse{}, nil
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
	actorUserID, err := s.actorUserIDFromContext(ctx)
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
	actorUserID, err := s.actorUserIDFromContext(ctx)
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

func (s *AuthServer) GetSession(ctx context.Context, req *pbauthv1.GetSessionRequest) (*pbauthv1.GetSessionResponse, error) {
	session, err := s.sessionRep.GetByID(ctx, req.SessionId)
	if err != nil {
		return nil, authStatusError(err)
	}
	return &pbauthv1.GetSessionResponse{Session: toProtoSession(session)}, nil
}

func (s *AuthServer) ListSessions(ctx context.Context, req *pbauthv1.ListSessionsRequest) (*pbauthv1.ListSessionsResponse, error) {
	actorUserID, err := s.actorUserIDFromContext(ctx)
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

func (s *AuthServer) RevokeSession(ctx context.Context, req *pbauthv1.RevokeSessionRequest) (*pbauthv1.RevokeSessionResponse, error) {
	actorUserID, err := s.actorUserIDFromContext(ctx)
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

func (s *AuthServer) ListAuditLogs(ctx context.Context, req *pbauthv1.ListAuditLogsRequest) (*pbauthv1.ListAuditLogsResponse, error) {
	actorUserID, err := s.actorUserIDFromContext(ctx)
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

func (s *AuthServer) RefreshToken(ctx context.Context, req *pbauthv1.RefreshTokenRequest) (*pbauthv1.RefreshTokenResponse, error) {
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

func (s *AuthServer) ListUserTenants(ctx context.Context, req *pbauthv1.ListUserTenantsRequest) (*pbauthv1.ListUserTenantsResponse, error) {
	actorUserID, err := s.actorUserIDFromContext(ctx)
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

func (s *AuthServer) ListPlatformRoles(ctx context.Context, req *pbauthv1.ListPlatformRolesRequest) (*pbauthv1.ListPlatformRolesResponse, error) {
	actorUserID, err := s.actorUserIDFromContext(ctx)
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

func (s *AuthServer) AddPlatformRole(ctx context.Context, req *pbauthv1.AddPlatformRoleRequest) (*pbauthv1.AddPlatformRoleResponse, error) {
	actorUserID, err := s.actorUserIDFromContext(ctx)
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
	s.recordAudit(ctx, actorUserID, "platform.role.added", "platform_role_membership", fmt.Sprintf("%s:%d", req.RoleName, req.TargetUserId), "", map[string]any{
		"target_user_id": targetUserID,
		"role_name":      req.RoleName,
	})
	return &pbauthv1.AddPlatformRoleResponse{}, nil
}

func (s *AuthServer) RemovePlatformRole(ctx context.Context, req *pbauthv1.RemovePlatformRoleRequest) (*pbauthv1.RemovePlatformRoleResponse, error) {
	actorUserID, err := s.actorUserIDFromContext(ctx)
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
	s.recordAudit(ctx, actorUserID, "platform.role.removed", "platform_role_membership", fmt.Sprintf("%s:%d", req.RoleName, req.TargetUserId), "", map[string]any{
		"target_user_id": targetUserID,
		"role_name":      req.RoleName,
	})
	return &pbauthv1.RemovePlatformRoleResponse{}, nil
}

func (s *AuthServer) ListTenantMembers(ctx context.Context, req *pbauthv1.ListTenantMembersRequest) (*pbauthv1.ListTenantMembersResponse, error) {
	actorUserID, err := s.actorUserIDFromContext(ctx)
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

func (s *AuthServer) RemoveTenantMember(ctx context.Context, req *pbauthv1.RemoveTenantMemberRequest) (*pbauthv1.RemoveTenantMemberResponse, error) {
	actorUserID, err := s.actorUserIDFromContext(ctx)
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
	s.recordAudit(ctx, actorUserID, "tenant.member.removed", "tenant_member", fmt.Sprintf("%s:%d", req.TenantId, req.UserId), req.TenantId, map[string]any{
		"user_id": userID,
	})
	return &pbauthv1.RemoveTenantMemberResponse{}, nil
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

func toProtoSession(s *entity.Session) *pbauthv1.Session {
	if s == nil {
		return nil
	}
	resp := &pbauthv1.Session{
		Id:             s.ID,
		UserId:         uint64(s.UserID),
		ActiveTenantId: s.ActiveTenantID,
		Status:         s.Status,
		CreatedAt:      s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      s.UpdatedAt.Format(time.RFC3339),
		ExpiresAt:      s.ExpiresAt.Format(time.RFC3339),
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
		errors.Is(err, iamdomain.ErrInvalidRoleName):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, iamdomain.ErrTenantNotFound),
		errors.Is(err, iamdomain.ErrMembershipNotFound),
		errors.Is(err, iamdomain.ErrRoleNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, iamdomain.ErrTenantSlugTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, iamdomain.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, iamdomain.ErrInactiveMembership):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *AuthServer) actorUserIDFromContext(ctx context.Context) (uint, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, errors.New("missing request metadata")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return 0, errors.New("missing authorization header")
	}
	raw := strings.TrimSpace(values[0])
	if !strings.HasPrefix(strings.ToLower(raw), "bearer ") {
		return 0, errors.New("invalid authorization header")
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
		return 0, errors.New("invalid access token")
	}
	if s.jwtKey != "" && claims.Key != s.jwtKey {
		return 0, errors.New("invalid access token")
	}
	if claims.UserID == 0 {
		return 0, errors.New("access token missing user_id")
	}
	return uint(claims.UserID), nil
}
