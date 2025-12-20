package pdredis

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

redis:
  test:
    uri: %q
`, "localhost:6379")
	toolkit.MakeConfigDir(t, config)

	appTest := fxtest.New(t,
		pdconfig.Module,
		pdlog.Module,
		ModuleFor("test"),
	)

	defer appTest.RequireStop()
	appTest.RequireStart()
}
