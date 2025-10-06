package pdelasticsearch

import (
	"fmt"
	"testing"

	"go.uber.org/fx/fxtest"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdtestenv"
)

func TestElasticsearchModule_Integration(t *testing.T) {
	mock := pdtestenv.Setup(t, pdtestenv.Options{
		StartElasticsearch: true,
		Reuse:              true,
		Namespace:          "podzone",
	})
	esURL := mock.ElasticsearchURL

	cfg := fmt.Sprintf(`
logger:
  app_name: "test"
  provider: "slog"
  level: "debug"
  env: "dev"

elasticsearch:
  test:
    addresses: ["%s"]
`, esURL)
	pdtestenv.MakeConfigDir(t, cfg)

	app := fxtest.New(t,
		pdconfig.Module,
		pdlog.Module,
		ModuleFor("test"),
	)

	app.RequireStart()
	defer app.RequireStop()
}
