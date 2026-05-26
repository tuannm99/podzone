package iamclient

import (
	"context"
	"fmt"

	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/entity"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type TenantAccessChecker struct {
	client     pbauthv1.IAMServiceClient
	projection outputport.IAMProjectionRepository
}

var _ outputport.TenantAccessChecker = (*TenantAccessChecker)(nil)

type TenantAccessCheckerParams struct {
	fx.In
	Lifecycle  fx.Lifecycle
	Logger     pdlog.Logger
	Config     config.AuthConfig
	Projection outputport.IAMProjectionRepository `optional:"true"`
}

func NewTenantAccessChecker(p TenantAccessCheckerParams) (outputport.TenantAccessChecker, error) {
	addr := fmt.Sprintf("%s:%s", p.Config.IAM.GRPCHost, p.Config.IAM.GRPCPort)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Closing IAM gRPC client connection")
			return conn.Close()
		},
	})
	return &TenantAccessChecker{
		client:     pbauthv1.NewIAMServiceClient(conn),
		projection: p.Projection,
	}, nil
}

func (c *TenantAccessChecker) EnsureActiveMembership(ctx context.Context, tenantID string, userID uint) error {
	if c.projection != nil {
		membership, err := c.projection.GetTenantMembership(ctx, tenantID, userID)
		if err != nil {
			return err
		}
		if membership != nil {
			if membership.Status != entity.MembershipStatusActive {
				return entity.ErrInactiveMembership
			}
			return nil
		}
	}

	resp, err := c.client.GetTenantMembership(ctx, &pbauthv1.GetTenantMembershipRequest{
		TenantId: tenantID,
		UserId:   uint64(userID),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return entity.ErrMembershipNotFound
		}
		return err
	}
	membership := resp.GetMembership()
	if membership == nil {
		return entity.ErrMembershipNotFound
	}
	if membership.Status != entity.MembershipStatusActive {
		return entity.ErrInactiveMembership
	}
	return nil
}
