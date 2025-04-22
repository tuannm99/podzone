package grpcgatewayfx

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(newGatewayMux),
	fx.Invoke(startHTTPGateway),
)

type GatewayMuxParams struct {
	fx.In

	Opts []runtime.ServeMuxOption `group:"gateway-options"`
}

func newGatewayMux(p GatewayMuxParams) *runtime.ServeMux {
	return runtime.NewServeMux(p.Opts...)
}
