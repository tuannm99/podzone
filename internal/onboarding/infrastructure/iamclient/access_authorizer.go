package iamclient

import (
	"context"
	"fmt"
	"strconv"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	onboardingconfig "github.com/tuannm99/podzone/internal/onboarding/config"
	storeoutputport "github.com/tuannm99/podzone/internal/onboarding/domain/store/outputport"
	"github.com/tuannm99/podzone/internal/onboarding/runtime/identity"
	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
)

const storeApprovalPermission = "store:approve"

type AccessAuthorizer struct {
	client pbiamv1.IAMQueryServiceClient
}

var _ storeoutputport.AccessAuthorizer = (*AccessAuthorizer)(nil)

type AccessAuthorizerParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    pdlog.Logger
	Config    onboardingconfig.AuthConfig
}

func NewAccessAuthorizer(params AccessAuthorizerParams) (*AccessAuthorizer, error) {
	addr := params.Config.IAM.GRPCHost + ":" + params.Config.IAM.GRPCPort
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect onboarding IAM client %s: %w", addr, err)
	}
	params.Lifecycle.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return conn.Close()
		},
	})
	params.Logger.Info("onboarding IAM gRPC client connected", "addr", addr)
	return &AccessAuthorizer{client: pbiamv1.NewIAMQueryServiceClient(conn)}, nil
}

func (a *AccessAuthorizer) AuthorizeStoreRequest(
	ctx context.Context,
	workspaceID string,
	requestedBy string,
) error {
	if identity.IsTrustedService(ctx) {
		return nil
	}
	userID, err := parseUserID(requestedBy)
	if err != nil {
		return err
	}
	response, err := a.client.CheckPermission(ctx, &pbiamv1.CheckPermissionRequest{
		TenantId:   workspaceID,
		UserId:     userID,
		Permission: "store:create",
		Resource:   "tenant/" + workspaceID + "/stores/*",
	})
	if err != nil {
		return fmt.Errorf("check store creation permission: %w", err)
	}
	if !response.GetAllowed() {
		return fmt.Errorf("permission denied: store:create")
	}
	return nil
}

func (a *AccessAuthorizer) AuthorizeStoreRead(
	ctx context.Context,
	workspaceID string,
	requestedBy string,
) error {
	if identity.IsTrustedService(ctx) {
		return nil
	}
	userID, err := parseUserID(requestedBy)
	if err != nil {
		return err
	}
	response, err := a.client.CheckPermission(ctx, &pbiamv1.CheckPermissionRequest{
		TenantId:   workspaceID,
		UserId:     userID,
		Permission: "store:read",
		Resource:   "tenant/" + workspaceID + "/stores/*",
	})
	if err != nil {
		return fmt.Errorf("check store read permission: %w", err)
	}
	if !response.GetAllowed() {
		return fmt.Errorf("permission denied: store:read")
	}
	return nil
}

func (a *AccessAuthorizer) AuthorizeStoreApproval(ctx context.Context, requestedBy string) error {
	if identity.IsTrustedService(ctx) {
		return nil
	}
	userID, err := parseUserID(requestedBy)
	if err != nil {
		return err
	}
	response, err := a.client.CheckPlatformPermission(ctx, &pbiamv1.CheckPlatformPermissionRequest{
		UserId:     userID,
		Permission: storeApprovalPermission,
	})
	if err != nil {
		return fmt.Errorf("check store approval permission: %w", err)
	}
	if !response.GetAllowed() {
		return fmt.Errorf("permission denied: %s", storeApprovalPermission)
	}
	return nil
}

func parseUserID(value string) (uint64, error) {
	userID, err := strconv.ParseUint(value, 10, 64)
	if err != nil || userID == 0 {
		return 0, fmt.Errorf("invalid user_id")
	}
	return userID, nil
}
