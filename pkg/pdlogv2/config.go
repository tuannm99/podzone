package pdlogv2

// Config is usually loaded by app (e.g. via Viper) and injected.
type Config struct {
	Provider string `mapstructure:"provider"` // "zap" | "slog" | "noop"
	Level    string `mapstructure:"level"`    // "debug" | "info" | "warn" | "error"
	Env      string `mapstructure:"env"`      // "dev" | "prod"
	AppName  string `mapstructure:"-"`        // set by caller
}

type Loader func() Config

func Defaults(app string) Loader {
	return func() Config {
		return Config{Provider: "zap", Level: "info", Env: "prod", AppName: app}
	}
}
