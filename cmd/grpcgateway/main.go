package main

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/tuannm99/podzone/pkg/globalmiddlewarefx"
	"github.com/tuannm99/podzone/pkg/grpcgatewayfx"
	"github.com/tuannm99/podzone/pkg/logfx"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"github.com/tuannm99/podzone/services/grpcgateway"

	pbAuth "github.com/tuannm99/podzone/pkg/api/proto/auth"
)

func main() {
	_ = godotenv.Load()

	app := fx.New(
		logfx.Module,
		globalmiddlewarefx.CommonHttpModule,
		grpcgatewayfx.Module,

		fx.Provide(
			fx.Annotate(
				func() grpcgateway.GatewayRegistrar {
					return &grpcgateway.AuthRegistrar{
						AddrVal: toolkit.FallbackEnv("AUTH_GRPC_ADDR", "localhost:50051"),
					}
				},
				fx.ResultTags(`group:"gateway-registrars"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() grpcgateway.GatewayRegistrar {
					return &grpcgateway.CatalogRegistrar{
						AddrVal: toolkit.FallbackEnv("CATALOG_GRPC_ADDR", "localhost:50052"),
					}
				},
				fx.ResultTags(`group:"gateway-registrars"`),
			),
		),

		fx.Provide(
			fx.Annotate(
				NewRedirectResponseModifier,
				fx.ResultTags(`group:"gateway-options"`),
			),
		),

		fx.Invoke(grpcgateway.RegisterGWHandlers),
	)

	app.Run()
}

// Auth Callback redirect
func NewRedirectResponseModifier(logger *zap.Logger) runtime.ServeMuxOption {
	return runtime.WithForwardResponseOption(
		func(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
			if loginResp, ok := resp.(*pbAuth.GoogleLoginResponse); ok && loginResp.RedirectUrl != "" {
				logger.Info("Redirecting to OAuth provider", zap.String("url", loginResp.RedirectUrl))
				w.Header().Set("Location", loginResp.RedirectUrl)
				w.WriteHeader(http.StatusTemporaryRedirect)
				return nil
			}
			if callbackResp, ok := resp.(*pbAuth.GoogleCallbackResponse); ok && callbackResp.RedirectUrl != "" {
				logger.Info("Redirecting to app after OAuth callback", zap.String("url", callbackResp.RedirectUrl))
				w.Header().Set("Location", callbackResp.RedirectUrl)
				w.WriteHeader(http.StatusTemporaryRedirect)
				return nil
			}
			return nil
		},
	)
}
