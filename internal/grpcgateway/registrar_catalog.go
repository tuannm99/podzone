package grpcgateway

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pb "github.com/tuannm99/podzone/pkg/api/proto/catalog"
)

type CatalogRegistrar struct {
	AddrVal string
}

func (r *CatalogRegistrar) Register(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pb.RegisterCatalogServiceHandler(ctx, mux, conn)
}
func (r *CatalogRegistrar) Addr() string { return r.AddrVal }
func (r *CatalogRegistrar) Name() string { return "catalog" }
