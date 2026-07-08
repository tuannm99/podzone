package testkit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func ElasticsearchURL(t *testing.T) string {
	t.Helper()
	startElasticsearch(t)
	if errES != nil {
		if isDockerUnavailable(errES) {
			t.Skipf("docker unavailable: %v", errES)
		}
		t.Fatalf("start elasticsearch container: %v", errES)
	}
	registerCleanup(t, esC)
	return esURL
}

func startElasticsearch(t *testing.T) {
	t.Helper()
	esMu.Lock()
	defer esMu.Unlock()
	defer captureDockerPanic(&errES)

	if isContainerRunning(esC) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "elasticsearch:8.12.0",
		ExposedPorts: []string{"9200/tcp"},
		Env: map[string]string{
			"discovery.type":         "single-node",
			"xpack.security.enabled": "false",
			"ES_JAVA_OPTS":           "-Xms512m -Xmx512m",
		},
		WaitingFor: wait.ForListeningPort("9200/tcp").WithStartupTimeout(3 * time.Minute),
	}
	reuse := reuseEnabled()
	if reuse {
		req.Name = "podzone-it-es"
	}

	esC, errES = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            reuse,
	})
	if errES != nil {
		return
	}

	host, err := esC.Host(ctx)
	if err != nil {
		errES = err
		return
	}
	port, err := esC.MappedPort(ctx, "9200/tcp")
	if err != nil {
		errES = err
		return
	}
	esURL = fmt.Sprintf("http://%s:%s", host, port.Port())
}
