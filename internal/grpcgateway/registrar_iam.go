package grpcgateway

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

type IAMRegistrar struct {
	AddrVal string
}

func (r *IAMRegistrar) Register(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pb.RegisterIAMServiceHandler(ctx, mux, conn)
}

func (r *IAMRegistrar) Addr() string { return r.AddrVal }
func (r *IAMRegistrar) Name() string { return "iam" }
