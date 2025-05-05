package main

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	pbAuth "github.com/tuannm99/podzone/pkg/api/proto/auth"
	pbOrder "github.com/tuannm99/podzone/pkg/api/proto/order"
	"github.com/tuannm99/podzone/pkg/globalmiddlewarefx"
	"github.com/tuannm99/podzone/pkg/grpcgatewayfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		logfx.Module,
		globalmiddlewarefx.CommonHttpModule,
		grpcgatewayfx.Module,

		fx.Provide(
			fx.Annotate(
				func() GatewayRegistrar {
					return &AuthRegistrar{AddrVal: toolkit.FallbackEnv("AUTH_GRPC_ADDR", "localhost:50051")}
				},
				fx.ResultTags(`group:"gateway-registrars"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() GatewayRegistrar {
					return &OrderRegistrar{AddrVal: toolkit.FallbackEnv("ORDER_GRPC_ADDR", "localhost:50052")}
				},
				fx.ResultTags(`group:"gateway-registrars"`),
			),
		),

		fx.Invoke(RegisterGWHandlers),
	)

	app.Run()
}

type GatewayRegistrar interface {
	Register(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
	Addr() string
	Name() string
}

// auth service
type AuthRegistrar struct {
	AddrVal string
}

func (r *AuthRegistrar) Register(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pbAuth.RegisterAuthServiceHandler(ctx, mux, conn)
}
func (r *AuthRegistrar) Addr() string { return r.AddrVal }
func (r *AuthRegistrar) Name() string { return "auth" }

// order service
type OrderRegistrar struct {
	AddrVal string
}

func (r *OrderRegistrar) Register(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return pbOrder.RegisterOrderServiceHandler(ctx, mux, conn)
}
func (r *OrderRegistrar) Addr() string { return r.AddrVal }
func (r *OrderRegistrar) Name() string { return "order" }

type GWParams struct {
	fx.In
	Mux        *runtime.ServeMux
	Logger     *zap.Logger
	Registrars []GatewayRegistrar `group:"gateway-registrars"`
}

func RegisterGWHandlers(p GWParams) {
	p.Logger.Info("Launching gRPC Gateway dynamic registration")
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	for _, registrar := range p.Registrars {
		go monitorService(p.Logger, registrar, p.Mux, opts)
	}
}

func monitorService(logger *zap.Logger, registrar GatewayRegistrar, mux *runtime.ServeMux, opts []grpc.DialOption) {
	name := registrar.Name()
	addr := registrar.Addr()
	var conn *grpc.ClientConn
	var err error

	// Retry until success
	for {
		conn, err = grpc.NewClient(addr, opts...)
		if err == nil {
			break
		}
		logger.Warn("Failed to dial gRPC for registration", zap.String("service", name), zap.Error(err))
		time.Sleep(3 * time.Second)
	}

	if err := registrar.Register(context.Background(), mux, conn); err != nil {
		logger.Error("Initial handler registration failed", zap.String("service", name), zap.Error(err))
		return
	}

	logger.Info("Service registered", zap.String("service", name), zap.String("addr", addr))

	// Background health checks
	go func() {
		client := grpc_health_v1.NewHealthClient(conn)
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
			cancel()

			if err != nil || resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
				logger.Warn("Health check failed", zap.String("service", name), zap.Error(err))
			} else {
				logger.Debug("Health check OK", zap.String("service", name))
			}

			time.Sleep(10 * time.Second)
		}
	}()
}
