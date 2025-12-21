package main

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"google.golang.org/protobuf/proto"

	"github.com/tuannm99/podzone/internal/grpcgateway"
	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpcgateway"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdpprof"
	"github.com/tuannm99/podzone/pkg/toolkit"

	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

func main() {
	_ = godotenv.Load()
	newAppContainer().Run()
}

func newAppContainer() *fx.App {
	return fx.New(
		pdconfig.Module,
		pdlog.Module,

		pdpprof.Module,
		pdglobalmiddleware.CommonHttpModule,
		pdgrpcgateway.Module,

		fx.Provide(
			fx.Annotate(
				func() grpcgateway.GatewayRegistrar {
					return &grpcgateway.AuthRegistrar{
						AddrVal: toolkit.GetEnv("AUTH_GRPC_ADDR", "localhost:50051"),
					}
				},
				fx.ResultTags(`group:"gateway-registrars"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() grpcgateway.GatewayRegistrar {
					return &grpcgateway.CatalogRegistrar{
						AddrVal: toolkit.GetEnv("CATALOG_GRPC_ADDR", "localhost:50052"),
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
}

func NewRedirectResponseModifier(logger pdlog.Logger) runtime.ServeMuxOption {
	return runtime.WithForwardResponseOption(RedirectForwardFunc(logger))
}

func RedirectForwardFunc(
	logger pdlog.Logger,
) func(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	return func(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
		if loginResp, ok := resp.(*pbauthv1.GoogleLoginResponse); ok && loginResp.RedirectUrl != "" {
			logger.Info("Redirecting to OAuth provider", "url", loginResp.RedirectUrl)
			w.Header().Set("Location", loginResp.RedirectUrl)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return nil
		}
		if callbackResp, ok := resp.(*pbauthv1.GoogleCallbackResponse); ok && callbackResp.RedirectUrl != "" {
			logger.Info("Redirecting to app after OAuth callback", "url", callbackResp.RedirectUrl)
			w.Header().Set("Location", callbackResp.RedirectUrl)
			w.WriteHeader(http.StatusTemporaryRedirect)
			return nil
		}
		return nil
	}
}
