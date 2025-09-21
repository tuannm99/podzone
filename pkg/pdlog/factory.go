package pdlog

import (
	"github.com/tuannm99/podzone/pkg/toolkit"
)

type LoggerFactory = toolkit.Factory[Logger, Config]

// config.yml -> `logger.*`
type Config struct {
	Provider string `mapstructure:"provider"` // "zap" | "slog" | "noop"
	Level    string `mapstructure:"level"`    // "debug" | "info" | "warn" | "error"
	Env      string `mapstructure:"env"`      // "dev" | "prod"
	AppName  string `mapstructure:"-"`        // set trong ModuleFor(appName)
}

var Registry = toolkit.NewRegistry[Logger, Config]("zap")

func init() {
	Registry.Register("zap", zapFactory)
	Registry.Register("slog", slogFactory)
	Registry.Register("noop", noopFactory)
}
