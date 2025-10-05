package pdelasticsearch

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxtest"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdtestenv"
)

func TestElasticsearchModule_Integration(t *testing.T) {
	mockConn := pdtestenv.Setup(t, pdtestenv.Options{
		// StartPostgres:   true,
		// StartRedis:      true,
		// StartMongo:      true,
		StartElasticsearch: true,
		Reuse:              true,
		Namespace:          "podzone",
	})
	esURL := mockConn.OpenSearchURL

	config := fmt.Sprintf(`
logger:
  app_name: "test"
  provider: "slog"
  level: "debug"
  env: "dev"

elasticsearch:
  test:
    addresses: ["%s"]
    ping_timeout: 3
    connect_timeout: 5
`, esURL)

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(config), 0o644))
	t.Setenv("CONFIG_PATH", path)

	app := fxtest.New(t,
		pdconfig.Module,
		pdlog.Module,
		ModuleFor("test"),
	)

	app.RequireStart()
	defer app.RequireStop()
}
