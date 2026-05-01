package backoffice

import (
	"context"
	"fmt"
	"strconv"

	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type TenantAuthorizer interface {
	AuthorizeTenant(ctx context.Context, sessionID, userID, tenantID string) error
	RequirePermission(ctx context.Context, userID, tenantID, permission string) error
}

type authTenantAuthorizer struct {
	client pbauthv1.AuthServiceClient
}

type authClientParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    pdlog.Logger
	Config    boconfig.Config
}

func NewTenantAuthorizer(p authClientParams) (TenantAuthorizer, error) {
	addr := p.Config.Auth.GRPCHost + ":" + p.Config.Auth.GRPCPort
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect auth grpc %s: %w", addr, err)
	}

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return conn.Close()
		},
	})

	p.Logger.Info("backoffice auth gRPC client connected", "addr", addr)
	return &authTenantAuthorizer{client: pbauthv1.NewAuthServiceClient(conn)}, nil
}

func (a *authTenantAuthorizer) AuthorizeTenant(ctx context.Context, sessionID, userID, tenantID string) error {
	sessionResp, err := a.client.GetSession(ctx, &pbauthv1.GetSessionRequest{SessionId: sessionID})
	if err != nil {
		return mapAuthzError("session validation failed", err)
	}
	session := sessionResp.GetSession()
	if session == nil {
		return fmt.Errorf("session not found")
	}
	if session.Status != "active" {
		return fmt.Errorf("session is not active")
	}
	if session.UserId != mustParseUserID(userID) {
		return fmt.Errorf("session user mismatch")
	}
	if session.ActiveTenantId != tenantID {
		return fmt.Errorf("session tenant mismatch")
	}

	userIDNum, err := parseUserID(userID)
	if err != nil {
		return err
	}

	resp, err := a.client.GetTenantMembership(ctx, &pbauthv1.GetTenantMembershipRequest{
		TenantId: tenantID,
		UserId:   userIDNum,
	})
	if err != nil {
		return mapAuthzError("tenant membership check failed", err)
	}
	if resp.GetMembership() == nil {
		return fmt.Errorf("tenant membership not found")
	}
	if resp.GetMembership().Status != "active" {
		return fmt.Errorf("tenant membership is inactive")
	}
	return nil
}

func (a *authTenantAuthorizer) RequirePermission(ctx context.Context, userID, tenantID, permission string) error {
	userIDNum, err := parseUserID(userID)
	if err != nil {
		return err
	}

	resp, err := a.client.CheckPermission(ctx, &pbauthv1.CheckPermissionRequest{
		TenantId:   tenantID,
		UserId:     userIDNum,
		Permission: permission,
	})
	if err != nil {
		return mapAuthzError("permission check failed", err)
	}
	if !resp.GetAllowed() {
		return fmt.Errorf("permission denied: %s", permission)
	}
	return nil
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
