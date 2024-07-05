package server

import (
	"context"
	"github.com/aiocean/wireset/configsvc"
	"github.com/aiocean/wireset/fiberapp"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"os"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/wire"
	"go.uber.org/zap"
)

type ApiServer struct {
	MsgRouter           *message.Router
	ConfigSvc           *configsvc.ConfigService
	LogSvc              *zap.Logger
	FiberSvc            *fiber.App
	HttpHandlerRegistry *fiberapp.Registry
	Features            []Feature
}

var DefaultWireset = wire.NewSet(
	wire.Struct(new(ApiServer), "*"),
	wire.Bind(new(Server), new(*ApiServer)),
)

func (s *ApiServer) Start(ctx context.Context) chan error {
	errChan := make(chan error, 1)

	// in it features
	for _, feature := range s.Features {
		s.LogSvc.Info("Initializing feature", zap.String("feature", feature.Name()))
		if err := feature.Init(); err != nil {
			errChan <- errors.WithMessage(err, "failed to init feature")
			return errChan
		}
	}

	// start message router
	go func() {
		err := s.MsgRouter.Run(ctx)
		if err != nil {
			errChan <- errors.WithMessage(err, "failed to run message router")
			return
		}

		errChan <- errors.New("message router stopped")
	}()

	// start fiber
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}

		s.HttpHandlerRegistry.RegisterMiddlewares(s.FiberSvc)
		s.HttpHandlerRegistry.RegisterHandlers(s.FiberSvc)

		if err := s.FiberSvc.Listen(":" + port); err != nil {
			errChan <- errors.WithMessage(err, "failed to listen fiber")
			return
		}

		errChan <- errors.New("fiber stopped")
	}()

	return errChan
}
