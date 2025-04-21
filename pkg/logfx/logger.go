package logfx

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/tuannm99/podzone/pkg/common"
)

type Config struct {
	level string
	env   string
}

func fromEnv() *Config {
	conf := &Config{
		level: common.FallbackEnv("DEFAULT_LOG_LEVEL", "debug"),
		env:   common.FallbackEnv("APP_ENV", "dev"), // dev | prod
	}
	return conf
}

func LoggerProvider() (*zap.Logger, error) {
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
