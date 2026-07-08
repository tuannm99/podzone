package testkit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoURI(t *testing.T) string {
	t.Helper()
	startMongo(t)
	if errMongo != nil {
		if isDockerUnavailable(errMongo) {
			t.Skipf("docker unavailable: %v", errMongo)
		}
		t.Fatalf("start mongo container: %v", errMongo)
	}
	registerCleanup(t, mongoC)
	return mongoURI
}

func MongoClient(t *testing.T) *mongo.Client {
	t.Helper()
	uri := MongoURI(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect mongo: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("ping mongo: %v", err)
	}
	t.Cleanup(func() { _ = client.Disconnect(context.Background()) })
	return client
}

func startMongo(t *testing.T) {
	t.Helper()
	mongoMu.Lock()
	defer mongoMu.Unlock()
	defer captureDockerPanic(&errMongo)

	if isContainerRunning(mongoC) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:8.0",
		ExposedPorts: []string{"27017/tcp"},
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "podzone",
			"MONGO_INITDB_ROOT_PASSWORD": "podzone123",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("27017/tcp"),
			wait.ForLog("Waiting for connections"),
		).WithStartupTimeout(2 * time.Minute),
	}
	reuse := reuseEnabled()
	if reuse {
		req.Name = "podzone-it-mongo"
	}

	mongoC, errMongo = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            reuse,
	})
	if errMongo != nil {
		return
	}

	host, err := mongoC.Host(ctx)
	if err != nil {
		errMongo = err
		return
	}
	port, err := mongoC.MappedPort(ctx, "27017/tcp")
	if err != nil {
		errMongo = err
		return
	}
	mongoURI = fmt.Sprintf(
		"mongodb://podzone:podzone123@%s:%s/?authSource=admin&directConnection=true",
		host,
		port.Port(),
	)
	if err := waitForMongoReady(ctx, mongoURI); err != nil {
		errMongo = err
	}
}

func waitForMongoReady(ctx context.Context, uri string) error {
	deadline := time.Now().Add(30 * time.Second)
	var lastErr error

	for time.Now().Before(deadline) {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		client, err := mongo.Connect(pingCtx, options.Client().ApplyURI(uri))
		if err == nil {
			err = client.Ping(pingCtx, nil)
		}
		if client != nil {
			_ = client.Disconnect(context.Background())
		}
		cancel()

		if err == nil {
			return nil
		}
		lastErr = err
		time.Sleep(500 * time.Millisecond)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("mongo not ready before timeout")
	}
	return lastErr
}
