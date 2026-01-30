package pdredis

import (
	"context"
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/testkit"
)

func TestGetConfigFromKoanf(t *testing.T) {
	k := koanf.New(".")

	k.Set("redis.test.uri", "redis://localhost:6379/0")

	cfg, err := GetConfigFromKoanf("test", k)
	require.NoError(t, err)
	assert.Equal(t, "redis://localhost:6379/0", cfg.URI)
}

func TestNewClientFromConfig_ValidURI(t *testing.T) {
	addr := testkit.RedisAddr(t)
	cfg := &Config{URI: "redis://" + addr + "/2"}
	client, err := NewClientFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	opts := client.Options()
	assert.Equal(t, addr, opts.Addr)
	assert.Equal(t, 2, opts.DB)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, client.Ping(ctx).Err())
}

func TestNewClientFromConfig_NilConfig(t *testing.T) {
	client, err := NewClientFromConfig(nil)
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "nil redis config")
}

func TestNewClientFromConfig_InvalidURI(t *testing.T) {
	t.Parallel()

	cfg := &Config{URI: "redis://localhost:abc/0"}
	client, err := NewClientFromConfig(cfg)
	require.Error(t, err)
	require.Nil(t, client)
}
