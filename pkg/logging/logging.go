package logging

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log  *zap.Logger
	once sync.Once
)

func GetLogger() *zap.Logger {
	once.Do(func() {
		level := "info"
		env := "prod"

		logger, err := NewLogger(level, env)
		if err != nil {
			logger, _ = zap.NewProduction()
		}
		log = logger
	})
	return log
}

func GetLoggerWithConfig(level string, env string) *zap.Logger {
	once.Do(func() {
		if level == "" {
			level = "info"
		}
		if env == "" {
			env = "development"
		}

		logger, err := NewLogger(level, env)
		if err != nil {
			logger, _ = zap.NewProduction()
		}
		log = logger
	})
	return log
}

// NewLogger creates a new zap logger instance with the specified configuration
func NewLogger(level string, env string) (*zap.Logger, error) {
	var config zap.Config

	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("invalid log level: %s", level)
	}

	if env == "prod" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.Level = zap.NewAtomicLevelAt(logLevel)

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}

func WithContext(logger *zap.Logger, fields ...zapcore.Field) *zap.Logger {
	return logger.With(fields...)
}
