package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/pkg/pdtestenv"
)

func TestMain(t *testing.T) {
	mockConn := pdtestenv.Setup(t, pdtestenv.Options{
		StartPostgres: true,
		StartRedis:    true,
		// StartMongo:      true,
		// StartOpenSearch: true,
		Reuse:     true,
		Namespace: "podzone",
	})
	pgDSN := mockConn.PostgresDSN
	redisURI := mockConn.RedisURI

	config := fmt.Sprintf(`
logger:
  app_name: "test"
  provider: "slog"
  level: "debug"
  env: "dev"

redis:
  auth:
    uri: %q

sql:
  auth:
    uri: %q
    provider: postgres
    should_run_migration: false

grpc:
  port: 0
`, redisURI, pgDSN)

	// Write config to tmp file
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(config), 0o644))
	t.Setenv("CONFIG_PATH", path)

	// Start main in goroutine
	done := make(chan struct{})
	go func() {
		main()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
}
