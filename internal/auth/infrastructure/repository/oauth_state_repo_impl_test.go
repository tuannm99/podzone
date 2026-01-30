package repository

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/testkit"
)

func setupOauthStateRepo(t *testing.T) (*OauthStateRepositoryImpl, *redis.Client) {
	t.Helper()
	client := testkit.RedisClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, client.FlushDB(ctx).Err())

	repo := NewOauthStateRepositoryImpl(OauthStateRepoParams{
		RedisClient: client,
		Logger:      pdlog.NopLogger{},
	})
	return repo, client
}

func TestOauthStateRepository_Set_OK(t *testing.T) {
	repo, client := setupOauthStateRepo(t)

	key := "oauth:google:test"
	ttl := 2 * time.Minute

	err := repo.Set(key, ttl)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	val, err := client.Get(ctx, key).Result()
	require.NoError(t, err)
	require.NotEmpty(t, val)
}

func TestOauthStateRepository_Get_OK(t *testing.T) {
	repo, client := setupOauthStateRepo(t)

	key := "oauth:google:test"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, client.Set(ctx, key, "some-state", time.Minute).Err())

	val, err := repo.Get(key)
	require.NoError(t, err)
	require.Equal(t, "some-state", val)
}

func TestOauthStateRepository_Get_NotFound(t *testing.T) {
	repo, _ := setupOauthStateRepo(t)

	val, err := repo.Get("oauth:missing")
	require.Error(t, err)
	require.Empty(t, val)
	require.Contains(t, err.Error(), "invalid state")
}

func TestOauthStateRepository_Del_OK(t *testing.T) {
	repo, client := setupOauthStateRepo(t)

	key := "oauth:del"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, client.Set(ctx, key, "v", time.Minute).Err())

	err := repo.Del(key)
	require.NoError(t, err)

	exists, err := client.Exists(ctx, key).Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), exists)
}

func TestOauthStateRepository_Get_RedisClosed(t *testing.T) {
	repo, client := setupOauthStateRepo(t)

	require.NoError(t, client.Close())

	_, err := repo.Get("oauth:error")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get state")
}
