package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestOauthStateRepository_Set_OK(t *testing.T) {
	rdb, mock := redismock.NewClientMock()

	repo := NewOauthStateRepositoryImpl(OauthStateRepoParams{
		RedisClient: rdb,
		Logger:      nopLogger{},
	})

	key := "oauth:google:test"
	ttl := 2 * time.Minute

	mock.Regexp().ExpectSet(key, `.+`, ttl).SetVal("OK")

	err := repo.Set(key, ttl)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOauthStateRepository_Get_OK(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewOauthStateRepositoryImpl(OauthStateRepoParams{
		RedisClient: rdb,
		Logger:      nopLogger{},
	})

	key := "oauth:google:test"
	mock.ExpectGet(key).SetVal("some-state")

	val, err := repo.Get(key)
	require.NoError(t, err)
	require.Equal(t, "some-state", val)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOauthStateRepository_Get_NotFound(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewOauthStateRepositoryImpl(OauthStateRepoParams{
		RedisClient: rdb,
		Logger:      nopLogger{},
	})

	key := "oauth:missing"
	mock.ExpectGet(key).RedisNil()

	val, err := repo.Get(key)
	require.Error(t, err)
	require.Empty(t, val)
	require.Contains(t, err.Error(), "invalid state")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOauthStateRepository_Get_RedisError(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewOauthStateRepositoryImpl(OauthStateRepoParams{
		RedisClient: rdb,
		Logger:      nopLogger{},
	})

	key := "oauth:error"
	mock.ExpectGet(key).SetErr(errors.New("boom"))

	_, err := repo.Get(key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get state")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOauthStateRepository_Del_OK(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewOauthStateRepositoryImpl(OauthStateRepoParams{
		RedisClient: rdb,
		Logger:      nopLogger{},
	})

	key := "oauth:del"
	mock.ExpectDel(key).SetVal(1)

	err := repo.Del(key)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOauthStateRepository_Del_Error(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewOauthStateRepositoryImpl(OauthStateRepoParams{
		RedisClient: rdb,
		Logger:      nopLogger{},
	})

	key := "oauth:del:err"
	mock.ExpectDel(key).SetErr(errors.New("del failed"))

	err := repo.Del(key)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// Optional: test redis.Nil branch explicitly (tương tự RedisNil() ở trên)
func TestOauthStateRepository_Get_RedisNil_UsingErr(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewOauthStateRepositoryImpl(OauthStateRepoParams{
		RedisClient: rdb,
		Logger:      nopLogger{},
	})

	key := "oauth:nil"
	mock.ExpectGet(key).SetErr(redis.Nil)

	_, err := repo.Get(key)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid state")
	require.NoError(t, mock.ExpectationsWereMet())
}
