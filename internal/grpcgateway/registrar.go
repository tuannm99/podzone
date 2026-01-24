package grpcgateway

import (
	"context"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
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
	Lifecycle  fx.Lifecycle
	Mux        *runtime.ServeMux
	Logger     pdlog.Logger
	Registrars []GatewayRegistrar `group:"gateway-registrars"`
}

func RegisterGWHandlers(p GWParams) {
	p.Logger.Info("Launching gRPC Gateway registration")
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	ctx, cancel := context.WithCancel(context.Background())
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cancel()
			return nil
		},
	})

	var muxMu sync.Mutex

	for _, registrar := range p.Registrars {
		go func(reg GatewayRegistrar) {
			name := reg.Name()
			addr := reg.Addr()

			var conn *grpc.ClientConn
			var err error

			logCtx := p.Logger.With("service", name)

			// Retry gRPC dial until success or shutdown.
			for {
				if ctx.Err() != nil {
					return
				}
				conn, err = grpc.NewClient(addr, opts...)
				if err == nil {
					break
				}
				logCtx.Warn("Failed to dial gRPC for registration", "err", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(3 * time.Second):
				}
			}

			muxMu.Lock()
			err = reg.Register(ctx, p.Mux, conn)
			muxMu.Unlock()
			if err != nil {
				logCtx.Warn("Handler registration failed", "err", err)
				return
			}

			logCtx.Info("Service registered", "addr", addr)
			go func() {
				<-ctx.Done()
				_ = conn.Close()
			}()

			// Launch health check in background
			go monitorHealth(ctx, p.Logger, reg, conn)
		}(registrar)
	}
}

func monitorHealth(ctx context.Context, logger pdlog.Logger, registrar GatewayRegistrar, conn *grpc.ClientConn) {
	name := registrar.Name()
	client := grpc_health_v1.NewHealthClient(conn)

	logCtx := logger.With("service", name)
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}

		checkCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		resp, err := client.Check(checkCtx, &grpc_health_v1.HealthCheckRequest{})
		cancel()

		if err != nil || resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
			logCtx.Debug("Health check failed", "err", err)
		} else {
			logCtx.Debug("Health check OK")
		}
	}
}
