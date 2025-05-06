package grpcgateway

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pbOrder "github.com/tuannm99/podzone/pkg/api/proto/order"
)

type OrderRegistrar struct {
	AddrVal string
}

func (r *OrderRegistrar) Register(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pbOrder.RegisterOrderServiceHandler(ctx, mux, conn)
}
func (r *OrderRegistrar) Addr() string { return r.AddrVal }
func (r *OrderRegistrar) Name() string { return "order" }
