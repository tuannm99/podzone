package testkit

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

var (
	postgresMu  sync.Mutex
	postgresDSN string
	errPostgres error
	postgresC   testcontainers.Container

	redisMu   sync.Mutex
	redisAddr string
	errRedis  error
	redisC    testcontainers.Container

	mongoMu  sync.Mutex
	mongoURI string
	errMongo error
	mongoC   testcontainers.Container

	esMu  sync.Mutex
	esURL string
	errES error
	esC   testcontainers.Container
)

func reuseEnabled() bool {
	if v := os.Getenv("PODZONE_TC_REUSE"); v != "" {
		return v == "1" || strings.EqualFold(v, "true")
	}
	return strings.EqualFold(os.Getenv("TESTCONTAINERS_REUSE_ENABLE"), "true")
}

func registerCleanup(t *testing.T, c testcontainers.Container) {
	t.Helper()
	// Containers in this package are process-wide singletons, so terminating them
	// in per-test cleanup causes flakiness when later tests reuse the cached DSN/client.
	// We intentionally keep them alive for the whole test process.
	_ = c
}

func isDockerUnavailable(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "cannot connect to the docker daemon") ||
		strings.Contains(msg, "permission denied while trying to connect to the docker daemon") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "is the docker daemon running") ||
		strings.Contains(msg, "operation not permitted") ||
		strings.Contains(msg, "xdg_runtime_dir")
}

func isContainerRunning(c testcontainers.Container) bool {
	if c == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	state, err := c.State(ctx)
	return err == nil && state.Running
}

func captureDockerPanic(target *error) {
	if recovered := recover(); recovered != nil {
		*target = fmt.Errorf("docker unavailable: %v", recovered)
	}
}
