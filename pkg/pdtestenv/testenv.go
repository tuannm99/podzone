// initilizing connection to some external storage (infrastrucure layer),
// used when writing intergration test
package pdtestenv

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainers struct {
	PostgresDSN string
	RedisURI    string
	MongoURI    string
	cleanup     func()
}

var (
	env      *TestContainers
	initOnce sync.Once
)

// Global init for all tests
func Setup(t *testing.T) *TestContainers {
	t.Helper()
	initOnce.Do(func() {
		ctx := context.Background()
		cleanups := []func(){}

		// Start Postgres
		pgReq := tc.ContainerRequest{
			Image:        "postgres:15-alpine",
			Env:          map[string]string{"POSTGRES_PASSWORD": "postgres"},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
		}
		pg, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{ContainerRequest: pgReq, Started: true})
		if err != nil {
			t.Fatalf("start postgres: %v", err)
		}
		cleanups = append(cleanups, func() { _ = pg.Terminate(ctx) })

		pgHost, _ := pg.Host(ctx)
		pgPort, _ := pg.MappedPort(ctx, "5432/tcp")
		pgDSN := fmt.Sprintf("postgres://postgres:postgres@%s:%s/testdb?sslmode=disable", pgHost, pgPort.Port())

		// Start Redis
		rReq := tc.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(30 * time.Second),
		}
		r, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{ContainerRequest: rReq, Started: true})
		if err != nil {
			t.Fatalf("start redis: %v", err)
		}
		cleanups = append(cleanups, func() { _ = r.Terminate(ctx) })
		rHost, _ := r.Host(ctx)
		rPort, _ := r.MappedPort(ctx, "6379/tcp")
		rURI := fmt.Sprintf("redis://%s:%s/0", rHost, rPort.Port())

		// Start Mongo
		mReq := tc.ContainerRequest{
			Image:        "mongo:6.0",
			ExposedPorts: []string{"27017/tcp"},
			WaitingFor:   wait.ForListeningPort("27017/tcp").WithStartupTimeout(30 * time.Second),
		}
		m, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{ContainerRequest: mReq, Started: true})
		if err != nil {
			t.Fatalf("start mongo: %v", err)
		}
		cleanups = append(cleanups, func() { _ = m.Terminate(ctx) })
		mHost, _ := m.Host(ctx)
		mPort, _ := m.MappedPort(ctx, "27017/tcp")
		mURI := fmt.Sprintf("mongodb://%s:%s", mHost, mPort.Port())

		env = &TestContainers{
			PostgresDSN: pgDSN,
			RedisURI:    rURI,
			MongoURI:    mURI,
			cleanup: func() {
				for _, c := range cleanups {
					c()
				}
			},
		}
	})
	return env
}

// Cleanup at the end of all tests
func Teardown() {
	if env != nil && env.cleanup != nil {
		env.cleanup()
	}
}
