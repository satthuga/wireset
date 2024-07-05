package logsvc

import (
	"os"
	"time"

	"github.com/google/wire"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var DefaultWireset = wire.NewSet(NewLogger, DefaultConfig)

func DefaultConfig() (zap.Config, error) {

	loc := time.FixedZone("UTC+7", 7*60*60)

	environment := os.Getenv("ENVIRONMENT")

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, pae zapcore.PrimitiveArrayEncoder) {
			zapcore.RFC3339TimeEncoder(t.In(loc), pae)
		},
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	logLevel := zap.ErrorLevel
	if environment == "development" {
		logLevel = zap.DebugLevel
	}

	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(logLevel),
		Development:       environment == "development",
		EncoderConfig:     encoderConfig,
		DisableStacktrace: true,
		DisableCaller:     true,
		Encoding:          "json",
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	return zapConfig, nil
}

func NewLogger(zapConfig zap.Config) (*zap.Logger, error) {

	logger, err := zapConfig.Build(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return ignoreHealthCheckCore{c: c}
	}))
	if err != nil {
		return nil, err
	}

	return logger, nil

}
