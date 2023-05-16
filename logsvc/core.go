package logsvc

import (
	"go.uber.org/zap/zapcore"
	"strings"
)

type ignoreHealthCheckCore struct {
	c             zapcore.Core
	isHealthCheck bool
}

func (ig ignoreHealthCheckCore) Enabled(lv zapcore.Level) bool {
	return ig.c.Enabled(lv)
}

func (ig ignoreHealthCheckCore) With(fs []zapcore.Field) zapcore.Core {
	for _, f := range fs {
		// GRPC health check
		if f.Key == "grpc.service" && f.String == "grpc.health.v1.Health" {
			ig.isHealthCheck = true
			break
		}

		// HTTP health check
		if strings.HasSuffix(f.String, "/healthz") {
			ig.isHealthCheck = true
			break
		}
	}
	return ignoreHealthCheckCore{
		c:             ig.c.With(fs),
		isHealthCheck: ig.isHealthCheck,
	}
}

func (ig ignoreHealthCheckCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if ig.isHealthCheck {
		return nil
	}

	return ig.c.Check(e, ce)
}

func (ig ignoreHealthCheckCore) Write(e zapcore.Entry, fs []zapcore.Field) error {
	return ig.c.Write(e, fs)
}

func (ig ignoreHealthCheckCore) Sync() error {
	return ig.c.Sync()
}
