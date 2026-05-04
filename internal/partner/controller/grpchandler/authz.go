package grpchandler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt"
	partnerconfig "github.com/tuannm99/podzone/internal/partner/config"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TenantAuthorizer interface {
	AuthorizeTenant(ctx context.Context, tenantID, permission string) (string, error)
}

type partnerJWTClaims struct {
	UserID         uint   `json:"user_id"`
	ActiveTenantID string `json:"active_tenant_id"`
	SessionID      string `json:"session_id"`
	Key            string `json:"key"`
	jwt.StandardClaims
}

type authzClientParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    pdlog.Logger
	Config    partnerconfig.Config
}

type authTenantAuthorizer struct {
	cfg        partnerconfig.Config
	authClient pbauthv1.AuthServiceClient
	iamClient  pbauthv1.IAMServiceClient
}

func NewTenantAuthorizer(p authzClientParams) (TenantAuthorizer, error) {
	authAddr := p.Config.Auth.GRPCHost + ":" + p.Config.Auth.GRPCPort
	authConn, err := grpc.NewClient(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect auth grpc %s: %w", authAddr, err)
	}

	iamAddr := p.Config.IAM.GRPCHost + ":" + p.Config.IAM.GRPCPort
	iamConn, err := grpc.NewClient(iamAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = authConn.Close()
		return nil, fmt.Errorf("connect iam grpc %s: %w", iamAddr, err)
	}

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := authConn.Close(); err != nil {
				_ = iamConn.Close()
				return err
			}
			return iamConn.Close()
		},
	})

	p.Logger.Info("partner auth gRPC client connected", "addr", authAddr)
	p.Logger.Info("partner iam gRPC client connected", "addr", iamAddr)

	return &authTenantAuthorizer{
		cfg:        p.Config,
		authClient: pbauthv1.NewAuthServiceClient(authConn),
		iamClient:  pbauthv1.NewIAMServiceClient(iamConn),
	}, nil
}

func (a *authTenantAuthorizer) AuthorizeTenant(ctx context.Context, tenantID, permission string) (string, error) {
	userID, activeTenantID, sessionID, err := a.identityFromContext(ctx)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(tenantID) == "" {
		return "", fmt.Errorf("tenant id is required")
	}
	if activeTenantID != tenantID {
		return "", fmt.Errorf("authorization token tenant mismatch")
	}

	sessionResp, err := a.authClient.GetSession(ctx, &pbauthv1.GetSessionRequest{SessionId: sessionID})
	if err != nil {
		return "", mapAuthzError("session validation failed", err)
	}
	session := sessionResp.GetSession()
	if session == nil {
		return "", fmt.Errorf("session not found")
	}
	if session.Status != "active" {
		return "", fmt.Errorf("session is not active")
	}
	if session.UserId != mustParseUserID(userID) {
		return "", fmt.Errorf("session user mismatch")
	}
	if session.ActiveTenantId != tenantID {
		return "", fmt.Errorf("session tenant mismatch")
	}

	userIDNum, err := parseUserID(userID)
	if err != nil {
		return "", err
	}

	membershipResp, err := a.iamClient.GetTenantMembership(ctx, &pbauthv1.GetTenantMembershipRequest{
		TenantId: tenantID,
		UserId:   userIDNum,
	})
	if err != nil {
		return "", mapAuthzError("tenant membership check failed", err)
	}
	if membershipResp.GetMembership() == nil {
		return "", fmt.Errorf("tenant membership not found")
	}
	if membershipResp.GetMembership().Status != "active" {
		return "", fmt.Errorf("tenant membership is inactive")
	}

	permResp, err := a.iamClient.CheckPermission(ctx, &pbauthv1.CheckPermissionRequest{
		TenantId:   tenantID,
		UserId:     userIDNum,
		Permission: permission,
	})
	if err != nil {
		return "", mapAuthzError("permission check failed", err)
	}
	if !permResp.GetAllowed() {
		return "", fmt.Errorf("permission denied: %s", permission)
	}

	return userID, nil
}

func (a *authTenantAuthorizer) identityFromContext(ctx context.Context) (string, string, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", "", fmt.Errorf("missing request metadata")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return "", "", "", fmt.Errorf("missing authorization header")
	}
	header := values[0]
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(header, bearerPrefix) {
		return "", "", "", fmt.Errorf("invalid authorization header")
	}
	tokenStr := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
	if tokenStr == "" {
		return "", "", "", fmt.Errorf("missing authorization bearer token")
	}

	claims := &partnerJWTClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(tok *jwt.Token) (interface{}, error) {
		if tok.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(a.cfg.Auth.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", "", "", fmt.Errorf("invalid authorization token")
	}
	if a.cfg.Auth.JWTKey != "" && claims.Key != a.cfg.Auth.JWTKey {
		return "", "", "", fmt.Errorf("invalid authorization token")
	}
	if claims.UserID == 0 {
		return "", "", "", fmt.Errorf("authorization token missing user_id")
	}
	if claims.ActiveTenantID == "" {
		return "", "", "", fmt.Errorf("authorization token missing active_tenant_id")
	}
	if claims.SessionID == "" {
		return "", "", "", fmt.Errorf("authorization token missing session_id")
	}
	return strconv.FormatUint(uint64(claims.UserID), 10), claims.ActiveTenantID, claims.SessionID, nil
}

func parseUserID(userID string) (uint64, error) {
	if userID == "" {
		return 0, fmt.Errorf("authorization token missing user_id")
	}
	out, err := strconv.ParseUint(userID, 10, 64)
	if err != nil || out == 0 {
		return 0, fmt.Errorf("invalid user_id")
	}
	return out, nil
}

func mustParseUserID(userID string) uint64 {
	out, _ := strconv.ParseUint(userID, 10, 64)
	return out
}

func mapAuthzError(prefix string, err error) error {
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.InvalidArgument:
			return fmt.Errorf("%s: invalid request", prefix)
		case codes.NotFound:
			return fmt.Errorf("%s: tenant membership not found", prefix)
		case codes.PermissionDenied:
			return fmt.Errorf("%s: permission denied", prefix)
		case codes.FailedPrecondition:
			return fmt.Errorf("%s: tenant membership is inactive", prefix)
		default:
			return fmt.Errorf("%s: %s", prefix, st.Message())
		}
	}
	return fmt.Errorf("%s: %w", prefix, err)
}
