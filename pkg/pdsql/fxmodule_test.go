package pdsql

import (
	"fmt"
	"testing"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/toolkit"
	"go.uber.org/fx/fxtest"
)

func TestModuleFor(t *testing.T) {
	config := fmt.Sprintf(`
logger:
  app_name: "test"
  provider: "slog"
  level: "debug"
  env: "dev"

sql:
  test:
    uri: %q
    provider: postgres
    should_run_migration: false

`, "postgres://localhost:5432")
	toolkit.MakeConfigDir(t, config)

	appTest := fxtest.New(
		t,
		pdconfig.Module,
		pdlog.Module,
		ModuleFor("test"),
	)

	defer appTest.RequireStop()
	appTest.RequireStart()
}
