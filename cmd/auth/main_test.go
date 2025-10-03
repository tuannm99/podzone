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
  provider: "slog"
  level: "debug"
  env: "dev"

redis:
  auth:
    uri: redis://localhost:6379/0

sql:
  auth:
    uri: postgres://postgres:postgres@localhost:5432/auth
    provider: postgres

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
