package grpcgatewayfx

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(newGatewayMux),
	fx.Invoke(startHTTPGateway),
)

func newGatewayMux() *runtime.ServeMux {
	return runtime.NewServeMux()
}
