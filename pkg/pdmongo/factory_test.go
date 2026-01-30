package pdmongo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tuannm99/podzone/pkg/testkit"
)

func TestNewMongoDbFromConfig(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		uri     string
		wantErr bool
	}{
		{
			name:    "invalid_uri_returns_error",
			uri:     "mongodb://localhost:abc", // invalid port -> ERROR
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, err := NewMongoDbFromConfig(&Config{
				URI:            tc.uri,
				Database:       "db",
				ConnectTimeout: 50 * time.Millisecond,
				PingTimeout:    50 * time.Millisecond,
			})

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, c)

			// cleanup best-effort
			_ = c.Disconnect(context.Background())
		})
	}
}

func TestNewMongoDbFromConfig_WithServer(t *testing.T) {
	uri := testkit.MongoURI(t)

	c, err := NewMongoDbFromConfig(&Config{
		URI:            uri,
		Database:       "db",
		ConnectTimeout: 2 * time.Second,
		PingTimeout:    2 * time.Second,
	})
	require.NoError(t, err)
	require.NotNil(t, c)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.NoError(t, c.Ping(ctx, nil))

	_ = c.Disconnect(context.Background())
}
