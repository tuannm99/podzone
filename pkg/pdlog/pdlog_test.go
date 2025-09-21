package pdlog

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

type fakeLogger struct{ synced bool }

func (f *fakeLogger) With(...any) Logger { return f }
func (f *fakeLogger) Debug(string) Entry { return &fakeEntry{} }
func (f *fakeLogger) Info(string) Entry  { return &fakeEntry{} }
func (f *fakeLogger) Warn(string) Entry  { return &fakeEntry{} }
func (f *fakeLogger) Error(string) Entry { return &fakeEntry{} }
func (f *fakeLogger) Sync() error        { f.synced = true; return nil }

type fakeEntry struct{}

func (e *fakeEntry) With(...any) Entry { return e }
func (e *fakeEntry) Err(error) Entry   { return e }
func (e *fakeEntry) Send()             {}

func TestModuleFor_LoadsFromFile_AndLifecycleSync(t *testing.T) {
	Registry.Register("testfake", func(ctx context.Context, cfg Config) (Logger, error) {
		return &fakeLogger{}, nil
	})

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(cfgPath, []byte(`
logger:
  provider: "testfake"
  level: "debug"
  env: "dev"
`), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var got *fakeLogger
	app := fxtest.New(t,
		fx.Provide(func() *viper.Viper {
			v := viper.New()
			v.SetConfigFile(cfgPath)
			if err := v.ReadInConfig(); err != nil {
				t.Fatalf("read config: %v", err)
			}
			v.AutomaticEnv()
			v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
			return v
		}),
		ModuleFor("unit_test_app"),
		fx.Invoke(func(l Logger) {
			fl, ok := l.(*fakeLogger)
			if !ok {
				t.Fatalf("expected *fakeLogger, got %T", l)
			}
			got = fl
			l.Info("hello").With("k", "v").Send()
		}),
	)

	app.RequireStart()
	if got == nil {
		t.Fatalf("logger should be constructed")
	}
	if got.synced {
		t.Fatalf("should not be synced before Stop")
	}
	app.RequireStop()
	if !got.synced {
		t.Fatalf("expected Sync() to be called on Stop")
	}
}

func TestModuleFor_Defaults_WhenNoConfig(t *testing.T) {
	app := fxtest.New(t,
		fx.Provide(func() *viper.Viper {
			v := viper.New()
			v.AutomaticEnv()
			v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
			return v
		}),
		ModuleFor("unit_test_app"),
		fx.Invoke(func(l Logger) {
			if l == nil {
				t.Fatalf("logger is nil")
			}
			l.Debug("debug").Send()
			l.Info("info").Send()
			l.Warn("warn").Err(nil).Send()
			l.Error("err").With("x", 1).Send()
		}),
	)
	app.RequireStart()
	app.RequireStop()
}

func TestFactories_Noops_Slog_Zap(t *testing.T) {
	ln, err := noopFactory(context.Background(), Config{AppName: "t"})
	if err != nil || ln == nil {
		t.Fatalf("noopFactory failed: %v", err)
	}
	ln.Info("ok").Send()

	ls, err := slogFactory(context.Background(), Config{AppName: "t", Level: "debug", Env: "dev"})
	if err != nil || ls == nil {
		t.Fatalf("slogFactory failed: %v", err)
	}
	ls.Debug("ok").With("k", "v").Send()

	lz, err := zapFactory(context.Background(), Config{AppName: "t", Level: "info", Env: "dev"})
	if err != nil || lz == nil {
		t.Fatalf("zapFactory failed: %v", err)
	}
	lz.Info("ok").Send()
}

func TestRegistry_Get_Lookup(t *testing.T) {
	if _, ok := Registry.Lookup("noop"); !ok {
		t.Fatalf("expected noop factory in registry")
	}
	if Registry.Get() == nil {
		t.Fatalf("Registry.Get() returned nil")
	}
}
