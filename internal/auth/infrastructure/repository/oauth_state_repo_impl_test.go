package repository

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	tcred "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupRedis(t *testing.T) (*redis.Client, func()) {
	t.Helper()

	ctx := context.Background()
	rc, err := tcred.RunContainer(ctx)
	require.NoError(t, err)

	endpoint, err := rc.Endpoint(ctx, "")
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: endpoint, // "host:port"
		DB:   0,
	})
	require.NoError(t, client.Ping(ctx).Err())

	cleanup := func() {
		_ = client.Close()
		_ = rc.Terminate(context.Background())
	}
	return client, cleanup
}

func TestOauthStateRepository_Redis_Integration(t *testing.T) {
	rdb, cleanup := setupRedis(t)
	defer cleanup()

	repo := NewOauthStateRepositoryImpl(OauthStateRepoParams{
		RedisClient: rdb,
		Logger:      nopLogger{},
	})

	key := "oauth:google:test"
	err := repo.Set(key, 2*time.Minute)
	require.NoError(t, err)

	val, err := repo.Get(key)
	require.NoError(t, err)
	require.NotEmpty(t, val)

	require.NoError(t, repo.Del(key))

	_, err = repo.Get(key)
	require.Error(t, err)
}
