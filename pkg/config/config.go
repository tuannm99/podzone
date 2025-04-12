package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func LoadEnv() {
	_ = godotenv.Load()
}

type Config struct {
	ServiceName string        `mapstructure:"service_name"`
	ServicePort int           `mapstructure:"service_port"`
	Environment string        `mapstructure:"env"`
	LogLevel    string        `mapstructure:"log_level"`
	Timeout     time.Duration `mapstructure:"timeout"`

	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Services ServicesConfig `mapstructure:"services"`
}

type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
	MaxConns int    `mapstructure:"max_conns"`
	MinConns int    `mapstructure:"min_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type ServicesConfig struct {
	Catalog string `mapstructure:"catalog"`
	Order   string `mapstructure:"order"`
	User    string `mapstructure:"user"`
	Cart    string `mapstructure:"cart"`
	Payment string `mapstructure:"payment"`
	Gateway string `mapstructure:"gateway"`
}

func (c *DatabaseConfig) ConnectionString() string {
	switch c.Type {
	case "postgres":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
	case "mongodb":
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	default:
		return ""
	}
}

func Load(configDir string, serviceName string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	if configDir != "" {
		v.SetConfigName("config")
		v.AddConfigPath(configDir)
		v.AddConfigPath(filepath.Join(configDir, "config"))

		if serviceName != "" {
			serviceConfigPath := filepath.Join(configDir, "config", serviceName)
			v.AddConfigPath(serviceConfigPath)
		}

		if err := v.ReadInConfig(); err != nil {
			return nil, errors.Wrap(err, "failed to read config file")
		}
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	envFile := os.Getenv("ENV_FILE")
	if envFile != "" {
		v.SetConfigFile(envFile)
		if err := v.MergeInConfig(); err != nil {
			return nil, errors.Wrap(err, "failed to merge ENV file")
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	if serviceName != "" && cfg.ServiceName == "" {
		cfg.ServiceName = serviceName
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("service_port", 8000)
	v.SetDefault("env", "development")
	v.SetDefault("log_level", "info")
	v.SetDefault("timeout", "30s")

	v.SetDefault("database.type", "postgres")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.username", "postgres")
	v.SetDefault("database.database", "podzone")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_conns", 10)
	v.SetDefault("database.min_conns", 2)

	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)
}
