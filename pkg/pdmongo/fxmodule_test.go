package pdmongo

import (
	"fmt"
	"testing"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdtestenv"
	"go.uber.org/fx/fxtest"
)

func TestMongoModule_Integration(t *testing.T) {
	mockConn := pdtestenv.Setup(t, pdtestenv.Options{
		// StartPostgres:   true,
		// StartRedis:      true,
		StartMongo:      true,
		// StartOpenSearch: true,
		Reuse:           true,
		Namespace:       "podzone",
	})
	mongoURI := mockConn.MongoURI
	config := fmt.Sprintf(`
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

`, mongoURI)
	pdtestenv.MakeConfigDir(t, config)

	appTest := fxtest.New(
		t,
		pdconfig.Module,
		pdlog.Module,
		ModuleFor("test"),
	)

	defer appTest.RequireStop()
	appTest.RequireStart()
}
