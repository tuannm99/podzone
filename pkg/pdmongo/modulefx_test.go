package pdmongo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tuannm99/podzone/pkg/pdtestenv"
)

func TestMongoModule_Integration(t *testing.T) {
	env := pdtestenv.Setup(t)
	cfg := &Config{
		URI:            env.MongoURI,
		Database:       "testdb",
		ConnectTimeout: 5 * time.Second,
		PingTimeout:    3 * time.Second,
	}
	client, err := NewMongoDbFromConfig(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.NoError(t, client.Ping(ctx, nil))
}
