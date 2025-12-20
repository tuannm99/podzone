package pdelasticsearch

import (
	"fmt"
	"testing"

	"go.uber.org/fx/fxtest"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit"
)

func TestElasticsearchModule_Integration(t *testing.T) {
	cfg := fmt.Sprintf(`
logger:
  app_name: "test"
  provider: "slog"
  level: "debug"
  env: "dev"

elasticsearch:
  test:
    addresses: ["%s"]
        `, "http://localhost:9200")
	toolkit.MakeConfigDir(t, cfg)

	app := fxtest.New(t,
		pdconfig.Module,
		pdlog.Module,
		ModuleFor("test"),
	)

	app.RequireStart()
	defer app.RequireStop()
}
