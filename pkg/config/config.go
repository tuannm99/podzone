package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Config holds all configuration for the service
type Config struct {
	// Service info
	ServiceName string        `mapstructure:"service_name"`
	ServicePort int           `mapstructure:"service_port"`
	Environment string        `mapstructure:"env"`
	LogLevel    string        `mapstructure:"log_level"`
	Timeout     time.Duration `mapstructure:"timeout"`

	// Database
	Database DatabaseConfig `mapstructure:"database"`
	
	// Redis
	Redis RedisConfig `mapstructure:"redis"`
	
	// Telemetry
	Telemetry TelemetryConfig `mapstructure:"telemetry"`
	
	// Services endpoints (for service-to-service communication)
	Services ServicesConfig `mapstructure:"services"`
}

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	Type     string `mapstructure:"type"` // postgres, mongodb, etc.
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"ssl_mode"`
	MaxConns int    `mapstructure:"max_conns"`
	MinConns int    `mapstructure:"min_conns"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// TelemetryConfig holds telemetry and monitoring configuration
type TelemetryConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	JaegerAddress string `mapstructure:"jaeger_address"`
	SamplingRate  string `mapstructure:"sampling_rate"`
}

// ServicesConfig holds endpoints for other services
type ServicesConfig struct {
	Catalog string `mapstructure:"catalog"`
	Order   string `mapstructure:"order"`
	User    string `mapstructure:"user"`
	Cart    string `mapstructure:"cart"`
	Payment string `mapstructure:"payment"`
	Gateway string `mapstructure:"gateway"`
}

// ConnectionString returns a formatted connection string for the database
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

// Load loads the configuration from environment variables and config files
func Load(configDir string, serviceName string) (*Config, error) {
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Read from config directory if provided
	if configDir != "" {
		v.SetConfigName("config")
		v.AddConfigPath(configDir)
		v.AddConfigPath(filepath.Join(configDir, "config"))

		// Service-specific config
		if serviceName != "" {
			serviceConfigPath := filepath.Join(configDir, "config", serviceName)
			v.AddConfigPath(serviceConfigPath)
		}

		if err := v.ReadInConfig(); err != nil {
			return nil, errors.Wrap(err, "failed to read config file")
		}
	}

	// Override with environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Override from ENV file if available
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

	// Override service name if provided
	if serviceName != "" && cfg.ServiceName == "" {
		cfg.ServiceName = serviceName
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Service defaults
	v.SetDefault("service_port", 8000)
	v.SetDefault("env", "development")
	v.SetDefault("log_level", "info")
	v.SetDefault("timeout", "30s")

	// Database defaults
	v.SetDefault("database.type", "postgres")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.username", "postgres")
	v.SetDefault("database.database", "podzone")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_conns", 10)
	v.SetDefault("database.min_conns", 2)

	// Redis defaults
	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)

	// Telemetry defaults
	v.SetDefault("telemetry.enabled", true)
	v.SetDefault("telemetry.sampling_rate", "0.1")
}
