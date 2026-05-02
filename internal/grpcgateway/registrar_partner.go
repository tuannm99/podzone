package grpcgateway

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pb "github.com/tuannm99/podzone/pkg/api/proto/partner/v1"
)

type PartnerRegistrar struct {
	AddrVal string
}

func (r *PartnerRegistrar) Register(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pb.RegisterPartnerServiceHandler(ctx, mux, conn)
}

func (r *PartnerRegistrar) Addr() string { return r.AddrVal }
func (r *PartnerRegistrar) Name() string { return "partner" }
