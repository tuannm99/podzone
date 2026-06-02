package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/internal/grpcgateway"
	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdglobalmiddleware"
	"github.com/tuannm99/podzone/pkg/pdgrpcgateway"
	pdlog "github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdpprof"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"
	"google.golang.org/protobuf/types/known/emptypb"

	pbauthv1 "github.com/tuannm99/podzone/pkg/api/proto/auth/v1"
)

func TestAppContainerGraph(t *testing.T) {
	err := fx.ValidateApp(
		fx.NopLogger,
		pdconfig.Module,
		pdlog.Module,
		pdpprof.Module,
		pdglobalmiddleware.CommonHttpModule,
		pdgrpcgateway.Module,
		gatewayRegistrarModule,
		fx.Provide(
			fx.Annotate(
				NewRedirectResponseModifier,
				fx.ResultTags(`group:"gateway-options"`),
			),
		),
		fx.Invoke(grpcgateway.RegisterGWHandlers),
	)
	require.NoError(t, err)
}

func TestRedirectForwardFunc(t *testing.T) {
	var logger pdlog.Logger

	app := fxtest.New(t,
		fx.Provide(func() *koanf.Koanf {
			k := koanf.New(".")
			k.Set("logger.provider", "slog")
			k.Set("logger.level", "debug")
			k.Set("logger.env", "dev")
			// optional:
			// k.Set("logger.app_name", "test")
			return k
		}),
		fx.Provide(func(k *koanf.Koanf) (*pdlog.Config, error) {
			return pdlog.GetLogConfig(k)
		}),
		fx.Provide(func(cfg *pdlog.Config) (pdlog.Logger, error) {
			return pdlog.NewLogger(cfg)
		}),
		fx.Invoke(func(l pdlog.Logger) { logger = l }),
		fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
	)

	app.RequireStart()
	defer app.RequireStop()

	fn := RedirectForwardFunc(logger)
	ctx := context.Background()

	t.Run("login_redirect", func(t *testing.T) {
		rr := httptest.NewRecorder()
		msg := &pbauthv1.GoogleLoginResponse{RedirectUrl: "https://example.com/login"}
		if err := fn(ctx, rr, msg); err != nil {
			t.Fatalf("fn error: %v", err)
		}
		if rr.Code != http.StatusTemporaryRedirect {
			t.Fatalf("status = %d, want %d (body=%s)", rr.Code, http.StatusTemporaryRedirect, rr.Body.String())
		}
		if loc := rr.Header().Get("Location"); loc != "https://example.com/login" {
			t.Fatalf("Location = %q, want %q", loc, "https://example.com/login")
		}
	})

	t.Run("callback_redirect", func(t *testing.T) {
		rr := httptest.NewRecorder()
		msg := &pbauthv1.GoogleCallbackResponse{RedirectUrl: "https://example.com/app"}
		if err := fn(ctx, rr, msg); err != nil {
			t.Fatalf("fn error: %v", err)
		}
		if rr.Code != http.StatusTemporaryRedirect {
			t.Fatalf("status = %d, want %d", rr.Code, http.StatusTemporaryRedirect)
		}
		if loc := rr.Header().Get("Location"); loc != "https://example.com/app" {
			t.Fatalf("Location = %q, want %q", loc, "https://example.com/app")
		}
	})

	t.Run("no_redirect", func(t *testing.T) {
		rr := httptest.NewRecorder()
		msg := &emptypb.Empty{}
		if err := fn(ctx, rr, msg); err != nil {
			t.Fatalf("fn error: %v", err)
		}
		if loc := rr.Header().Get("Location"); loc != "" {
			t.Fatalf("Location = %q, want empty", loc)
		}
	})
}
