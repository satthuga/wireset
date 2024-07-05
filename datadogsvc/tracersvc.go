package datadogsvc

import (
	"github.com/aiocean/wireset/configsvc"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DatadogSvc struct {
	confSvc *configsvc.ConfigService
	logger  *zap.Logger
}

var DefaultWireset = wire.NewSet(NewDatadogSvc)

func NewDatadogSvc(
	confSvc *configsvc.ConfigService,
	logger *zap.Logger,
) (*DatadogSvc, func(), error) {
	tracerSvc := &DatadogSvc{
		confSvc: confSvc,
		logger:  logger.Named("tracesvc"),
	}

	return tracerSvc, tracerSvc.Stop, nil
}

func (s *DatadogSvc) Start() {

	tracer.Start(
		tracer.WithService(s.confSvc.ServiceName),
		tracer.WithEnv(s.confSvc.Environment),
		tracer.WithLogStartup(false),
		tracer.WithAnalytics(true),
	)
}

func (s *DatadogSvc) Stop() {
	tracer.Stop()
}
