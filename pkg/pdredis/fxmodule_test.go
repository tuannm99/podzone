package pdredis

import (
	"fmt"
	"testing"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdtestenv"
	"go.uber.org/fx/fxtest"
)

func TestModuleFor(t *testing.T) {
	mockConn := pdtestenv.Setup(t, pdtestenv.Options{
		// StartPostgres:   true,
		StartRedis: true,
		// StartMongo:      true,
		// StartOpenSearch: true,
		Reuse:     true,
		Namespace: "podzone",
	})
	config := fmt.Sprintf(`
logger:
  app_name: "test"
  provider: "slog"
  level: "debug"
  env: "dev"

redis:
  test:
    uri: %q
`, mockConn.RedisURI)
	pdtestenv.MakeConfigDir(t, config)

	appTest := fxtest.New(t,
		pdconfig.Module,
		pdlog.Module,
		ModuleFor("test"),
	)

	defer appTest.RequireStop()
	appTest.RequireStart()
}
