package main

import (
	"testing"
	"time"

	"github.com/tuannm99/podzone/pkg/pdtestenv"
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

mongo:
  catalog:
    uri: mongodb://localhost:27017
    database: catalog
    ping_timeout: 3s
    connect_timeout: 5s

grpc:
  port: 0
`

func TestMain(t *testing.T) {
	pdtestenv.MakeConfigDir(t, configAppTest)

	done := make(chan struct{})
	go func() { main(); close(done) }()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
	}
}
