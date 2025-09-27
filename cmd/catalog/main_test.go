package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var configAppTest = `
logger:
  app_name: "test"
  provider: "slog" # "zap" | "slog"
  level: "debug"
  env: "dev"

redis:
  catalog:
    uri: redis://localhost:6379/0
    provider: mock

mongo:
  catalog:
    uri: mongodb://localhost:27017/catalog
    provider: mock

grpc:
  port: 0
`

func TestMain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(configAppTest), 0o644))
	t.Setenv("CONFIG_PATH", path)

	done := make(chan struct{})
	go func() { main(); close(done) }()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
	}
}
