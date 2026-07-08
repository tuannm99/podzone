package testkit

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func RedisAddr(t *testing.T) string {
	t.Helper()
	startRedis(t)
	if errRedis != nil {
		if isDockerUnavailable(errRedis) {
			t.Skipf("docker unavailable: %v", errRedis)
		}
		t.Fatalf("start redis container: %v", errRedis)
	}
	registerCleanup(t, redisC)
	return redisAddr
}

func RedisURI(t *testing.T, db int) string {
	t.Helper()
	return fmt.Sprintf("redis://%s/%d", RedisAddr(t), db)
}

func RedisClient(t *testing.T) *redis.Client {
	t.Helper()
	addr := RedisAddr(t)
	client := redis.NewClient(&redis.Options{Addr: addr})
	var lastErr error
	for attempt := 0; attempt < 5; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		lastErr = client.Ping(ctx).Err()
		cancel()
		if lastErr == nil {
			t.Cleanup(func() { _ = client.Close() })
			return client
		}
		if isDockerUnavailable(lastErr) {
			t.Skipf("docker unavailable: %v", lastErr)
		}
		if strings.Contains(strings.ToLower(lastErr.Error()), "connection reset by peer") {
			time.Sleep(300 * time.Millisecond)
			continue
		}
		break
	}
	t.Fatalf("ping redis: %v", lastErr)
	return nil
}

func startRedis(t *testing.T) {
	t.Helper()
	redisMu.Lock()
	defer redisMu.Unlock()
	defer captureDockerPanic(&errRedis)

	if isContainerRunning(redisC) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(2 * time.Minute),
	}
	reuse := reuseEnabled()
	if reuse {
		req.Name = "podzone-it-redis"
	}

	redisC, errRedis = testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            reuse,
	})
	if errRedis != nil {
		return
	}

	host, err := redisC.Host(ctx)
	if err != nil {
		errRedis = err
		return
	}
	port, err := redisC.MappedPort(ctx, "6379/tcp")
	if err != nil {
		errRedis = err
		return
	}
	redisAddr = fmt.Sprintf("%s:%s", host, port.Port())
}
