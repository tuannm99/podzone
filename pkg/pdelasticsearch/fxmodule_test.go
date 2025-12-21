package pdelasticsearch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

// Mock transport that go-elasticsearch will call via elastic-transport.
// Important: NEVER return (nil, nil) because the library will dereference response internally.
type mockTransport struct {
	status int
	err    error
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.Perform(req)
}

// Perform is what elastic-transport expects.
func (m *mockTransport) Perform(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}

	body := io.NopCloser(strings.NewReader(""))
	h := make(http.Header)
	h.Set("X-Elastic-Product", "Elasticsearch") // IMPORTANT for ES v8 product check

	return &http.Response{
		StatusCode: m.status,
		Status:     fmt.Sprintf("%d %s", m.status, http.StatusText(m.status)),
		Body:       body,
		Header:     h,
		Request:    req,
	}, nil
}

func newMockESClient(t *testing.T, tr *mockTransport) *elasticsearch.Client {
	t.Helper()

	c, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://mock:9200"},
		Transport: tr,
	})
	require.NoError(t, err)
	return c
}

func TestElasticsearchLifecycle_OnStart_Table(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		tr        *mockTransport
		pingTO    time.Duration
		wantError bool
	}{
		{
			name:      "ping_ok_200",
			tr:        &mockTransport{status: 200},
			pingTO:    10 * time.Millisecond,
			wantError: false,
		},
		{
			name:      "ping_transport_error",
			tr:        &mockTransport{err: errors.New("dial failed")},
			pingTO:    10 * time.Millisecond,
			wantError: true,
		},
		{
			name:      "ping_status_error_500",
			tr:        &mockTransport{status: 500},
			pingTO:    10 * time.Millisecond,
			wantError: true,
		},
		{
			name:      "ping_timeout_default_when_cfg_zero",
			tr:        &mockTransport{status: 200},
			pingTO:    0, // will fallback to 3s in your code
			wantError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := newMockESClient(t, tc.tr)
			cfg := &Config{
				Addresses:   []string{"http://mock:9200"},
				PingTimeout: tc.pingTO,
			}

			app := fx.New(
				fx.Supply(client),
				fx.Supply(cfg),
				fx.Supply(fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger)))),
				fx.Invoke(registerLifecycle),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			err := app.Start(ctx)
			if tc.wantError {
				require.Error(t, err)
				_ = app.Stop(context.Background())
				return
			}

			require.NoError(t, err)
			require.NoError(t, app.Stop(context.Background()))
		})
	}
}
