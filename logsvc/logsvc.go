package logsvc

import (
	"os"
	"time"

	"github.com/google/wire"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// DefaultWireset provides the default wire set for the logging service
var DefaultWireset = wire.NewSet(NewLogger, DefaultConfig)

// Config represents the configuration for the logging service
type Config struct {
	Environment string
	TimeZone    *time.Location
	LogLevel    zapcore.Level
}

// DefaultConfig returns a default configuration for the logging service
func DefaultConfig() (*Config, error) {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "" {
		environment = "production"
	}

	logLevel := zap.ErrorLevel
	if environment == "development" {
		logLevel = zap.DebugLevel
	}

	loc := time.FixedZone("UTC+7", 7*60*60)

	return &Config{
		Environment: environment,
		TimeZone:    loc,
		LogLevel:    logLevel,
	}, nil
}

// NewLogger creates a new zap logger based on the provided configuration
func NewLogger(config *Config) (*zap.Logger, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339),
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(config.LogLevel),
		Development:       config.Environment == "development",
		EncoderConfig:     encoderConfig,
		DisableStacktrace: config.Environment != "development",
		DisableCaller:     config.Environment != "development",
		Encoding:          "json",
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	logger, err := zapConfig.Build(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stderr),
			config.LogLevel,
		)
	}))

	if err != nil {
		return nil, err
	}

	return logger.WithOptions(zap.WithCaller(config.Environment == "development")), nil
}
