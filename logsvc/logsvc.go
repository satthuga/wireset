package logsvc

import (
	"time"

	"github.com/google/wire"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var LogSvcWireset = wire.NewSet(NewLogger)

func NewLogger() (*zap.Logger, error) {
	loc := time.FixedZone("UTC+7", 7*60*60)

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

	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:       true,
		EncoderConfig:     encoderConfig,
		DisableStacktrace: true,
		DisableCaller:     true,
		Encoding:          "json",
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	logger, err := zapConfig.Build(zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return ignoreHealthCheckCore{c: c}
	}))
	if err != nil {
		return nil, err
	}

	return logger, nil

}
