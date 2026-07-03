package iamclient

import (
	"context"
	"fmt"

	"github.com/tuannm99/podzone/internal/auth/config"
	"github.com/tuannm99/podzone/internal/auth/domain/outputport"
	pbiamv1 "github.com/tuannm99/podzone/pkg/api/proto/iam/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type AccountBootstrapper struct {
	client pbiamv1.IAMCommandServiceClient
}

var _ outputport.AccountBootstrapper = (*AccountBootstrapper)(nil)

type AccountBootstrapperParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Logger    pdlog.Logger
	Config    config.AuthConfig
}

func NewAccountBootstrapper(p AccountBootstrapperParams) (*AccountBootstrapper, error) {
	addr := fmt.Sprintf("%s:%s", p.Config.IAM.GRPCHost, p.Config.IAM.GRPCPort)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("Closing IAM account-bootstrap gRPC client connection")
			return conn.Close()
		},
	})
	return &AccountBootstrapper{
		client: pbiamv1.NewIAMCommandServiceClient(conn),
	}, nil
}

func (b *AccountBootstrapper) EnsureRootOrganization(
	ctx context.Context,
	accessToken string,
	userID uint,
	username string,
) error {
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+accessToken)
	_, err := b.client.EnsureRootOrganization(ctx, &pbiamv1.EnsureRootOrganizationRequest{
		Name: username,
		Slug: fmt.Sprintf("account-%d", userID),
	})
	return err
}
