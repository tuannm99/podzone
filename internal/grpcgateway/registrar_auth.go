package grpcgateway

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	pb "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

type AuthRegistrar struct {
	AddrVal string
}

func (r *AuthRegistrar) Register(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pb.RegisterAuthServiceHandler(ctx, mux, conn)
}
func (r *AuthRegistrar) Addr() string { return r.AddrVal }
func (r *AuthRegistrar) Name() string { return "auth" }
