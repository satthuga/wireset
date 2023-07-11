package server

import (
	"context"
	"github.com/aiocean/wireset/configsvc"
	"github.com/aiocean/wireset/fiberapp"
	"github.com/aiocean/wireset/pubsub"
	"github.com/aiocean/wireset/tracersvc"
	"github.com/pkg/errors"
	"os"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gofiber/fiber/v2"
	"github.com/google/wire"
	"go.uber.org/zap"
)

type Server struct {
	MsgRouter           *message.Router
	TracerSvc           *tracersvc.TracerSvc
	ConfigSvc           *configsvc.ConfigService
	LogSvc              *zap.Logger
	FiberSvc            *fiber.App
	HttpHandlerRegistry *fiberapp.Registry
	EventBus            *pubsub.Pubsub
	Features            []Feature
}

var DefaultWireset = wire.NewSet(
	wire.Struct(new(Server), "*"),
)

func (s *Server) Start(ctx context.Context) chan error {
	if s.ConfigSvc.DataDogAgentAddress != "" {
		s.TracerSvc.Start()
	}

	errChan := make(chan error, 1)

	// in it features
	for _, feature := range s.Features {
		if err := feature.Init(); err != nil {
			errChan <- errors.WithMessage(err, "failed to init feature")
			return errChan
		}
	}

	// Register event bus
	if err := s.EventBus.Register(); err != nil {
		errChan <- errors.WithMessage(err, "failed to register event bus")
		return errChan
	}

	// start message router
	go func() {
		err := s.MsgRouter.Run(context.Background())
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

		s.HttpHandlerRegistry.RegisterHandlers(s.FiberSvc)

		if err := s.FiberSvc.Listen(":" + port); err != nil {
			errChan <- errors.WithMessage(err, "failed to listen fiber")
			return
		}

		errChan <- errors.New("fiber stopped")
	}()

	return errChan
}
