package partnerdirectory

import (
	"context"
	"fmt"

	boconfig "github.com/tuannm99/podzone/internal/backoffice/config"
	routingctx "github.com/tuannm99/podzone/internal/backoffice/domain/routing"
	pbpartnerv1 "github.com/tuannm99/podzone/pkg/api/proto/partner/v1"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ routingctx.PartnerDirectory = (*Adapter)(nil)

type Adapter struct {
	client pbpartnerv1.PartnerServiceClient
}

type params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    pdlog.Logger
	Config    boconfig.Config
}

func New(
	p params,
) (routingctx.PartnerDirectory, error) {
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
	return &Adapter{client: pbpartnerv1.NewPartnerServiceClient(conn)}, nil
}

func (d *Adapter) ListActivePartners(
	ctx context.Context,
	tenantID string,
) ([]routingctx.PartnerRoutingProfile, error) {
	resp, err := d.client.ListPartners(ctx, &pbpartnerv1.ListPartnersRequest{
		TenantId: tenantID,
		Status:   "active",
	})
	if err != nil {
		return nil, err
	}
	items := resp.GetPartners()
	out := make([]routingctx.PartnerRoutingProfile, 0, len(items))
	for _, item := range items {
		out = append(out, routingctx.PartnerRoutingProfile{
			ID:                    item.GetId(),
			Code:                  item.GetCode(),
			Name:                  item.GetName(),
			PartnerType:           item.GetPartnerType(),
			Status:                item.GetStatus(),
			SupportedProductTypes: append([]string(nil), item.GetSupportedProductTypes()...),
			SupportedRegions:      append([]string(nil), item.GetSupportedRegions()...),
			SLADays:               item.GetSlaDays(),
			RoutingPriority:       item.GetRoutingPriority(),
			BaseFulfillmentCost:   item.GetBaseFulfillmentCost(),
			ShippingCostRules:     toPartnerShippingCostRules(item.GetShippingCostRules()),
		})
	}
	return out, nil
}

func toPartnerShippingCostRules(items []*pbpartnerv1.ShippingCostRule) []routingctx.PartnerShippingCostRule {
	out := make([]routingctx.PartnerShippingCostRule, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		out = append(out, routingctx.PartnerShippingCostRule{
			Region: item.GetRegion(),
			Cost:   item.GetCost(),
		})
	}
	return out
}
