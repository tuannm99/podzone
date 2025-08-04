package logfx

import (
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/tuannm99/podzone/pkg/toolkit"
)

var Module = fx.Provide(
	NewLogger,
)

type loggerConfig struct {
	level string
	env   string
}

func fromEnv() *loggerConfig {
	conf := &loggerConfig{
		level: toolkit.GetEnv("DEFAULT_LOG_LEVEL", "debug"),
		env:   toolkit.GetEnv("APP_ENV", "dev"), // dev | prod
	}
	return conf
}

func NewLogger() (*zap.Logger, error) {
	conf := fromEnv()
	var config zap.Config

	if conf.env == "prod" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(conf.level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %s", conf.level)
	}
	config.Level = zap.NewAtomicLevelAt(logLevel)

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
