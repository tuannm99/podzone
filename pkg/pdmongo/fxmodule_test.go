package pdmongo

import (
	"fmt"
	"testing"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.uber.org/fx/fxtest"
)

func TestMongoModule_Integration(t *testing.T) {
	cfg := fmt.Sprintf(`
logger:
  app_name: "test"
  provider: "slog"
  level: "debug"
  env: "dev"

mongo:
  test:
    uri: %q
    database: catalog
    ping_timeout: 3s
    connect_timeout: 5s

`, "mongodb://localhost:27017")
	toolkit.MakeConfigDir(t, cfg)

	appTest := fxtest.New(
		t,
		pdconfig.Module,
		pdlog.Module,
		ModuleFor("test"),
	)

	defer appTest.RequireStop()
	appTest.RequireStart()
}
