package backoffice

import (
	"context"
	"fmt"

	"github.com/tuannm99/podzone/internal/backoffice/domain/entity"
	"github.com/tuannm99/podzone/internal/backoffice/domain/outputport"
	pbpartnerv1 "github.com/tuannm99/podzone/pkg/api/proto/partner/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type partnerDirectory struct {
	client pbpartnerv1.PartnerServiceClient
}

func NewPartnerDirectory(
	p authClientParams,
) (outputport.PartnerDirectory, error) {
	addr := p.Config.Partner.GRPCHost + ":" + p.Config.Partner.GRPCPort
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect partner grpc %s: %w", addr, err)
	}

	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return conn.Close()
		},
	})

	p.Logger.Info("backoffice partner gRPC client connected", "addr", addr)
	return &partnerDirectory{client: pbpartnerv1.NewPartnerServiceClient(conn)}, nil
}

func (d *partnerDirectory) ListActivePartners(
	ctx context.Context,
	tenantID string,
) ([]entity.PartnerRoutingProfile, error) {
	resp, err := d.client.ListPartners(ctx, &pbpartnerv1.ListPartnersRequest{
		TenantId: tenantID,
		Status:   "active",
	})
	if err != nil {
		return nil, err
	}
	items := resp.GetPartners()
	out := make([]entity.PartnerRoutingProfile, 0, len(items))
	for _, item := range items {
		out = append(out, entity.PartnerRoutingProfile{
			ID:                    item.GetId(),
			Code:                  item.GetCode(),
			Name:                  item.GetName(),
			PartnerType:           item.GetPartnerType(),
			Status:                item.GetStatus(),
			SupportedProductTypes: append([]string(nil), item.GetSupportedProductTypes()...),
			SupportedRegions:      append([]string(nil), item.GetSupportedRegions()...),
			SLADays:               item.GetSlaDays(),
			RoutingPriority:       item.GetRoutingPriority(),
		})
	}
	return out, nil
}
