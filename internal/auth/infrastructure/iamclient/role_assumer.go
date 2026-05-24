package iamclient

import (
	"context"
	"fmt"
	"time"

	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RoleAssumer struct {
	client pbauthv1.IAMServiceClient
}

var _ outputport.RoleAssumer = (*RoleAssumer)(nil)

type RoleAssumerParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    pdlog.Logger
	Config    config.AuthConfig
}

func NewRoleAssumer(p RoleAssumerParams) (outputport.RoleAssumer, error) {
	addr := fmt.Sprintf("%s:%s", p.Config.IAM.GRPCHost, p.Config.IAM.GRPCPort)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Closing IAM role-assumer gRPC client connection")
			return conn.Close()
		},
	})
	return &RoleAssumer{client: pbauthv1.NewIAMServiceClient(conn)}, nil
}

func (r *RoleAssumer) AssumeRole(
	ctx context.Context,
	input outputport.AssumeRoleInput,
) (*outputport.AssumedRole, error) {
	resp, err := r.client.AssumeRole(ctx, &pbauthv1.IAMAssumeRoleRequest{
		AccessToken:      input.AccessToken,
		RoleName:         input.RoleName,
		TenantId:         input.TenantID,
		ExternalId:       input.ExternalID,
		SessionName:      input.SessionName,
		SourceIdentity:   input.SourceIdentity,
		DurationSeconds:  input.DurationSeconds,
		ServicePrincipal: input.ServicePrincipal,
		SessionTags:      cloneStringMap(input.SessionTags),
	})
	if err != nil {
		return nil, err
	}
	assumedRole := resp.GetAssumedRole()
	if assumedRole == nil {
		return nil, fmt.Errorf("iam assume-role returned empty assumed_role")
	}
	return &outputport.AssumedRole{
		RoleID:           assumedRole.RoleId,
		RoleScope:        assumedRole.RoleScope,
		RoleName:         assumedRole.RoleName,
		TenantID:         assumedRole.TenantId,
		ServicePrincipal: assumedRole.ServicePrincipal,
		SessionName:      assumedRole.SessionName,
		SourceIdentity:   assumedRole.SourceIdentity,
		SessionTags:      cloneStringMap(assumedRole.SessionTags),
		ExpiresAt:        parseTimestamp(assumedRole.ExpiresAt),
	}, nil
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func parseTimestamp(raw string) time.Time {
	if raw == "" {
		return time.Time{}
	}
	ts, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return ts
}
