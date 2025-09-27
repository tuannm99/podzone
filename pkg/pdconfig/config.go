package pdconfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func NewAppConfig() (*viper.Viper, error) {
	v := viper.New()

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		fmt.Println("CONFIG_PATH is empty: running ENV-only mode")
	} else {
		v.SetConfigFile(path)
		if err := v.ReadInConfig(); err != nil {
			fmt.Println("No config file found, fallback to ENV only:", err)
		} else {
			fmt.Println("Loaded config from:", v.ConfigFileUsed())
		}
	}

	// ENV override: postgres.auth.uri -> POSTGRES_AUTH_URI
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return v, nil
}
