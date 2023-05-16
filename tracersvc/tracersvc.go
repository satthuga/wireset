package tracersvc

import (
	"github.com/aiocean/wireset/configsvc"
	"github.com/google/wire"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type TracerSvc struct {
	confSvc *configsvc.ConfigService
}

var TracerSvcWireset = wire.NewSet(NewTracerSvc)

func NewTracerSvc(
	confSvc *configsvc.ConfigService,
) (*TracerSvc, func(), error) {
	tracerSvc := &TracerSvc{
		confSvc: confSvc,
	}

	cleanup := func() {
		tracerSvc.Stop()
	}

	return tracerSvc, cleanup, nil
}

func (s *TracerSvc) Start() {
	tracer.Start(
		tracer.WithService(s.confSvc.ServiceName),
	)
}

func (s *TracerSvc) Stop() {
	tracer.Stop()
}
