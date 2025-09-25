package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	pdlog "github.com/tuannm99/podzone/pkg/pdlogv2"
	"github.com/tuannm99/podzone/pkg/pdlogv2/provider"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"google.golang.org/protobuf/types/known/emptypb"

	pbAuth "github.com/tuannm99/podzone/pkg/api/proto/auth"
)

var configAppTest = `
logger:
  app_name: "test"
  provider: "mock"
  level: "debug"
  env: "dev"

redis:
  auth:
    uri: redis://localhost:6379/0
    provider: mock
`

func TestMain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(configAppTest), 0o644))
	t.Setenv("CONFIG_PATH", path)

	done := make(chan struct{})
	go func() {
		main()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Log("main() still running, test will stop here")
	}
}

func TestRedirectForwardFunc(t *testing.T) {
	var logger pdlog.Logger

	app := fxtest.New(t,
		fx.Provide(func() *viper.Viper {
			v := viper.New()
			v.Set("logger.provider", "mock")
			v.Set("logger.level", "debug")
			v.Set("logger.env", "dev")
			return v
		}),
		pdlog.Module(
			pdlog.ViperLoaderFor("logger"),
			pdlog.WithProvider("mock", provider.MockFactory),
		),
		fx.Invoke(func(l pdlog.Logger) { logger = l }),
	)
	app.RequireStart()
	defer app.RequireStop()

	fn := RedirectForwardFunc(logger)
	ctx := context.Background()

	t.Run("login_redirect", func(t *testing.T) {
		rr := httptest.NewRecorder()
		msg := &pbAuth.GoogleLoginResponse{RedirectUrl: "https://example.com/login"}
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
		msg := &pbAuth.GoogleCallbackResponse{RedirectUrl: "https://example.com/app"}
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
