package pdconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

func NewAppConfig() (*koanf.Koanf, error) {
	k := koanf.New(".")

	// 1) YAML
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		fmt.Println("CONFIG_PATH is empty: running ENV-only mode")
	} else {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config file %q failed: %w", path, err)
		}

		expanded := os.ExpandEnv(string(b)) // expand ${VAR} in YAML (optional)
		if err := k.Load(rawbytes.Provider([]byte(expanded)), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("parse yaml config %q failed: %w", path, err)
		}
		fmt.Println("Loaded config from:", path)
	}

	// 2) ENV override
	prefix := strings.ToUpper(strings.TrimSpace(os.Getenv("ENV_PREFIX")))
	if prefix != "" && !strings.HasSuffix(prefix, "_") {
		prefix += "_"
	}

	// cb: ENV var name -> config key (string)
	cb := func(s string) string {
		if prefix != "" {
			s = strings.TrimPrefix(s, prefix) // strip prefix from key
		}
		s = strings.ReplaceAll(s, "__", "-") // optional trick if you need literal underscores
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "_", ".")
		return s
	}

	// delim "." means "a.b.c" becomes nested map {a:{b:{c:...}}}
	if err := k.Load(env.Provider(prefix, ".", cb), nil); err != nil {
		return nil, fmt.Errorf("load env config failed: %w", err)
	}

	return k, nil
}

