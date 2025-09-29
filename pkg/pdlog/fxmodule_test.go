package pdlog_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdlog/mocks"
)

var configAppTest = `
logger:
  app_name: "test"
  provider: "slog"
  level: "debug"
  env: "dev"
`

func newConfigFile(t *testing.T) string {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yml")
	require.NoError(t, os.WriteFile(path, []byte(configAppTest), 0o644))
	t.Setenv("CONFIG_PATH", path)
	return path
}

func TestRegisterLoggerLifecycle(t *testing.T) {
	tests := []struct {
		name       string
		syncErr    error
		expectWarn bool
	}{
		{"sync success", nil, false},
		{"stderr sync error (ignored)", errors.New("sync /dev/stderr: invalid argument"), false},
		{"other error (warn)", errors.New("boom"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newConfigFile(t)

			m := &mocks.MockLogger{}
			m.On("Sync").Return(tt.syncErr).Once()

			if tt.expectWarn {
				m.On("Warn", mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "logger sync failed")
				}), "error", tt.syncErr).Once()
			}

			app := fx.New(
				pdconfig.Module,
				pdlog.Module,
				fx.Decorate(
					func(_ pdlog.Logger) pdlog.Logger { return m },
				),
			)

			require.NoError(t, app.Start(context.Background()))
			require.NoError(t, app.Stop(context.Background()))

			m.AssertExpectations(t)
		})
	}
}
