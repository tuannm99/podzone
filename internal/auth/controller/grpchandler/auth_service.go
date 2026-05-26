package grpchandler

import (
	"context"
	"strconv"
	"time"

	authmapper "github.com/tuannm99/podzone/internal/auth/controller/mapper"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

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
	return authmapper.ToPBGoogleCallbackResponse(callbackResp), nil
}

func (s *AuthServer) ExchangeGoogleLogin(
	ctx context.Context,
	req *pbauthv1.ExchangeGoogleLoginRequest,
) (*pbauthv1.LoginResponse, error) {
	authResp, err := s.authUC.ExchangeOAuthLogin(ctx, req.ExchangeCode)
	if err != nil {
		return nil, authStatusError(err)
	}
	resp, err := authmapper.ToPBLoginResponse(authResp)
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
	resp, err := authmapper.ToPBLoginResponse(loginResp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *AuthServer) Register(ctx context.Context, req *pbauthv1.RegisterRequest) (*pbauthv1.RegisterResponse, error) {
	registerCmd, err := authmapper.ToRegisterCmd(req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	registerResp, err := s.authUC.Register(ctx, *registerCmd)
	if err != nil {
		return nil, err
	}
	resp, err := authmapper.ToPBRegisterResponse(registerResp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
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

	resp, err := authmapper.ToPBSwitchActiveTenantResponse(authResp)
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
		authmapper.FromPBSessionPolicyStatements(req.Statements),
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
		Session:  authmapper.ToPBSession(session),
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
		Session:  authmapper.ToPBSession(session),
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
		authmapper.FromPBSessionPolicyStatements(req.SessionPolicy),
		req.ExternalId,
		req.SessionName,
		req.SourceIdentity,
		req.DurationSeconds,
		req.ServicePrincipal,
		req.SessionTags,
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
		Session:  authmapper.ToPBSession(session),
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
		Session:  authmapper.ToPBSession(session),
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
	return &pbauthv1.GetSessionResponse{Session: authmapper.ToPBSession(session)}, nil
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
		out = append(out, authmapper.ToPBSession(&item))
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
		out = append(out, authmapper.ToPBAuditLog(&item))
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
	resp, err := authmapper.ToPBRefreshTokenResponse(authResp)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func (s *AuthServer) GetUserByIdentity(
	ctx context.Context,
	req *pbauthv1.GetUserByIdentityRequest,
) (*pbauthv1.GetUserByIdentityResponse, error) {
	user, err := s.userRepo.GetByUsernameOrEmail(req.Identity)
	if err != nil {
		return nil, authStatusError(err)
	}
	return &pbauthv1.GetUserByIdentityResponse{
		UserInfo: authmapper.ToPBUserInfo(user),
	}, nil
}

func (s *AuthServer) EnsureUserByEmail(
	ctx context.Context,
	req *pbauthv1.EnsureUserByEmailRequest,
) (*pbauthv1.EnsureUserByEmailResponse, error) {
	user, err := s.userRepo.GetByUsernameOrEmail(req.Email)
	created := false
	if err != nil {
		if err != entity.ErrUserNotFound {
			return nil, authStatusError(err)
		}
		user, err = s.userRepo.CreateByEmailIfNotExisted(req.Email)
		if err != nil {
			return nil, authStatusError(err)
		}
		created = true
	}
	return &pbauthv1.EnsureUserByEmailResponse{
		UserInfo: authmapper.ToPBUserInfo(user),
		Created:  created,
	}, nil
}

func (s *AuthServer) GetUserByID(
	ctx context.Context,
	req *pbauthv1.GetUserByIDRequest,
) (*pbauthv1.GetUserByIDResponse, error) {
	userID, err := toUint(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := s.userRepo.GetByID(strconv.FormatUint(uint64(userID), 10))
	if err != nil {
		return nil, authStatusError(err)
	}
	return &pbauthv1.GetUserByIDResponse{
		UserInfo: authmapper.ToPBUserInfo(user),
	}, nil
}
