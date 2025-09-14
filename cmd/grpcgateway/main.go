package main

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
	"google.golang.org/protobuf/proto"

	"github.com/tuannm99/podzone/internal/grpcgateway"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpcgateway"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit"

	pbAuth "github.com/tuannm99/podzone/pkg/api/proto/auth"
)

func main() {
	_ = godotenv.Load()

	logger, err := pdlog.NewFrom(
		toolkit.GetEnv("LOG_BACKEND", "zap"),
		context.Background(),
		pdlog.WithLevel(toolkit.GetEnv("DEFAULT_LOG_LEVEL", "debug")),
		pdlog.WithEnv(toolkit.GetEnv("APP_ENV", "dev")),
		pdlog.WithAppName(toolkit.GetEnv("APP_NAME", "podzone_admin_grpcgateway")),
	)
	if err != nil {
		log.Fatal("error init logger %w", err)
	}

	app := fx.New(
		fx.Provide(func() pdlog.Logger { return logger }),
		fx.Invoke(func(lc fx.Lifecycle, log pdlog.Logger) {
			lc.Append(fx.Hook{
				OnStop: func(context.Context) error { return log.Sync() },
			})
		}),

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

	app.Run()
}

// Auth Callback redirect
func NewRedirectResponseModifier(logger pdlog.Logger) runtime.ServeMuxOption {
	return runtime.WithForwardResponseOption(
		func(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
			if loginResp, ok := resp.(*pbAuth.GoogleLoginResponse); ok && loginResp.RedirectUrl != "" {
				logger.Info("Redirecting to OAuth provider").With("url", loginResp.RedirectUrl).Send()
				w.Header().Set("Location", loginResp.RedirectUrl)
				w.WriteHeader(http.StatusTemporaryRedirect)
				return nil
			}
			if callbackResp, ok := resp.(*pbAuth.GoogleCallbackResponse); ok && callbackResp.RedirectUrl != "" {
				logger.Info("Redirecting to app after OAuth callback").With("url", callbackResp.RedirectUrl).Send()
				w.Header().Set("Location", callbackResp.RedirectUrl)
				w.WriteHeader(http.StatusTemporaryRedirect)
				return nil
			}
			return nil
		},
	)
}
