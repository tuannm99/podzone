package pdmongo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/fx/fxtest"

	"github.com/tuannm99/podzone/pkg/pdlog"
)

type mockMongo struct {
	pingErr       error
	disconnectErr error
	pingCalled    int
	stopCalled    int
}

func (m *mockMongo) Ping(ctx context.Context) error {
	m.pingCalled++
	return m.pingErr
}

func (m *mockMongo) Disconnect(ctx context.Context) error {
	m.stopCalled++
	return m.disconnectErr
}

func TestMongoLifecycle_Table(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		pingErr       error
		disconnectErr error
		wantStartErr  bool
		wantStopErr   bool
	}{
		{name: "start_ok_stop_ok"},
		{name: "start_fail_ping", pingErr: errors.New("ping failed"), wantStartErr: true},
		{name: "stop_fail_disconnect", disconnectErr: errors.New("disconnect failed"), wantStopErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := &mockMongo{
				pingErr:       tc.pingErr,
				disconnectErr: tc.disconnectErr,
			}
			cfg := &Config{URI: "mongodb://unit-test"}

			app := fxtest.New(
				t,
				fx.Supply(
					fx.Annotate(m, fx.As(new(MongoClient))),
					cfg,
					fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger))),
				),
				fx.Invoke(registerLifecycle),
				fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()

			startErr := app.Start(ctx)
			if tc.wantStartErr {
				require.Error(t, startErr)
				require.GreaterOrEqual(t, m.pingCalled, 1)
				_ = app.Stop(context.Background())
				return
			}
			require.NoError(t, startErr)
			require.Equal(t, 1, m.pingCalled)

			stopErr := app.Stop(ctx)
			if tc.wantStopErr {
				require.Error(t, stopErr)
			} else {
				require.NoError(t, stopErr)
			}
			require.Equal(t, 1, m.stopCalled)
		})
	}
}

func TestModuleFor_DefaultName_Wiring(t *testing.T) {
	t.Parallel()

	k := koanf.New(".")

	k.Set("mongo.default.uri", "mongodb://127.0.0.1:27017")
	k.Set("mongo.default.database", "db")
	k.Set("mongo.default.ping_timeout", "50ms")
	k.Set("mongo.default.connect_timeout", "50ms")

	app := fx.New(
		ModuleFor(""), // <- cover default name branch
		fx.Supply(k),
		fx.Supply(fx.Annotate(pdlog.NopLogger{}, fx.As(new(pdlog.Logger)))),
	)

	require.NoError(t, app.Err())
}
