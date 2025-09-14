package grpcgateway

import (
	"context"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GatewayRegistrar interface {
	Register(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
	Addr() string
	Name() string
}

type GWParams struct {
	fx.In
	Mux        *runtime.ServeMux
	Logger     pdlog.Logger
	Registrars []GatewayRegistrar `group:"gateway-registrars"`
}

func RegisterGWHandlers(p GWParams) {
	p.Logger.Info("Launching gRPC Gateway dynamic registration")
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	var muxMu sync.Mutex

	for _, registrar := range p.Registrars {
		go func(reg GatewayRegistrar) {
			name := reg.Name()
			addr := reg.Addr()

			var conn *grpc.ClientConn
			var err error

			// Retry gRPC dial until success
			for {
				conn, err = grpc.NewClient(addr, opts...)
				if err == nil {
					break
				}
				p.Logger.Warn("Failed to dial gRPC for registration").With("service", name).Err(err)
				time.Sleep(3 * time.Second)
			}

			muxMu.Lock()
			err = reg.Register(context.Background(), p.Mux, conn)
			muxMu.Unlock()
			if err != nil {
				p.Logger.Warn("Handler registration failed").With("service", name).Err(err)
				return
			}

			p.Logger.Info("Service registered").With("service", name).With("addr", addr)

			// Launch health check in background
			go monitorHealth(p.Logger, reg, conn)
		}(registrar)
	}
}

func monitorHealth(logger pdlog.Logger, registrar GatewayRegistrar, conn *grpc.ClientConn) {
	name := registrar.Name()
	client := grpc_health_v1.NewHealthClient(conn)

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		cancel()

		if err != nil || resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
			logger.Debug("Health check failed").With("service", name).Err(err).Send()
		} else {
			logger.Debug("Health check OK").With("service", name).Send()
		}

		time.Sleep(3 * time.Second)
	}
}
