package trace

import (
	"github.com/aiocean/wireset/datadogsvc"
	"go.uber.org/zap"

	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	fibertrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gofiber/fiber.v2"
)

var DefaultWireset = wire.NewSet(
	wire.Struct(new(FeatureTrace), "*"),
)

type FeatureTrace struct {
	TraderSvc *datadogsvc.DatadogSvc
	FiberApp  *fiber.App
	Logger    *zap.Logger
}

func (f *FeatureTrace) Init() error {
	f.TraderSvc.Start()
	f.FiberApp.Use(fibertrace.Middleware())
	return nil
}
