package grpchandler

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/inputport"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	iamdomain "github.com/tuannm99/podzone/internal/iam/domain"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

type AuthServer struct {
	pbauthv1.UnimplementedAuthServiceServer
	authUC     inputport.AuthUsecase
	iamUC      iamdomain.IAMUsecase
	sessionRep outputport.SessionRepository
}

func NewAuthServer(authUC inputport.AuthUsecase, iamUC iamdomain.IAMUsecase, sessionRep outputport.SessionRepository) *AuthServer {
	return &AuthServer{
		authUC:     authUC,
		iamUC:      iamUC,
		sessionRep: sessionRep,
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
	ownerUserID, err := toUint(req.OwnerUserId)
	if err != nil {
		return nil, err
	}

	tenant, err := s.iamUC.CreateTenant(ctx, ownerUserID, iamdomain.CreateTenantCmd{
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		return nil, iamStatusError(err)
	}

	membership, err := s.iamUC.GetMembership(ctx, tenant.ID, ownerUserID)
	if err != nil {
		return nil, iamStatusError(err)
	}

	return &pbauthv1.CreateTenantResponse{
		Tenant:          toProtoTenant(tenant),
		OwnerMembership: toProtoMembership(membership),
	}, nil
}

func (s *AuthServer) AddTenantMember(
	ctx context.Context,
	req *pbauthv1.AddTenantMemberRequest,
) (*pbauthv1.AddTenantMemberResponse, error) {
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}

	if err := s.iamUC.AddMember(ctx, req.TenantId, userID, req.RoleName); err != nil {
		return nil, iamStatusError(err)
	}
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

func (s *AuthServer) SwitchActiveTenant(
	ctx context.Context,
	req *pbauthv1.SwitchActiveTenantRequest,
) (*pbauthv1.SwitchActiveTenantResponse, error) {
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}

	authResp, err := s.authUC.SwitchActiveTenant(ctx, userID, req.TenantId, req.AccessToken)
	if err != nil {
		return nil, authStatusError(err)
	}

	resp, err := toolkit.MapStruct[inputport.AuthResult, pbauthv1.SwitchActiveTenantResponse](*authResp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *AuthServer) GetSession(ctx context.Context, req *pbauthv1.GetSessionRequest) (*pbauthv1.GetSessionResponse, error) {
	session, err := s.sessionRep.GetByID(ctx, req.SessionId)
	if err != nil {
		return nil, authStatusError(err)
	}
	return &pbauthv1.GetSessionResponse{Session: toProtoSession(session)}, nil
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
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
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

func (s *AuthServer) ListTenantMembers(ctx context.Context, req *pbauthv1.ListTenantMembersRequest) (*pbauthv1.ListTenantMembersResponse, error) {
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
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.iamUC.RemoveMember(ctx, req.TenantId, userID); err != nil {
		return nil, iamStatusError(err)
	}
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
