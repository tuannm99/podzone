package pdmongo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
		{
			name:    "valid_uri_returns_client_without_need_server",
			uri:     "mongodb://127.0.0.1:27017",
			wantErr: false,
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
