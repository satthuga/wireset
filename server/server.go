package server

import (
	"context"
	"github.com/aiocean/wireset/configsvc"
	"github.com/aiocean/wireset/feature/core"
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
	MsgRouter   *message.Router
	TracerSvc   *tracersvc.TracerSvc
	ConfigSvc   *configsvc.ConfigService
	LogSvc      *zap.Logger
	FiberSvc    *fiber.App
	EventBus    *pubsub.Pubsub
	Features    []Feature
	CoreFeature *core.FeatureCore
}

var DefaultWireset = wire.NewSet(
	wire.Struct(new(Server), "*"),
	core.DefaultWireset,
)

func (s *Server) Start(ctx context.Context) chan error {
	if s.ConfigSvc.IsProduction() {
		s.TracerSvc.Start()
	}

	errChan := make(chan error, 1)

	if err := s.CoreFeature.Init(); err != nil {
		errChan <- errors.WithMessage(err, "failed to init core feature")
		return errChan
	}

	for _, feature := range s.Features {
		if err := feature.Init(); err != nil {
			errChan <- errors.WithMessage(err, "failed to init feature")
			return errChan
		}
	}

	if err := s.EventBus.Register(); err != nil {
		errChan <- errors.WithMessage(err, "failed to register event bus")
		return errChan
	}

	go func() {
		err := s.MsgRouter.Run(context.Background())
		if err != nil {
			errChan <- errors.WithMessage(err, "failed to run message router")
			return
		}

		errChan <- errors.New("message router stopped")
	}()

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		if err := s.FiberSvc.Listen(":" + port); err != nil {
			errChan <- errors.WithMessage(err, "failed to listen fiber")
			return
		}

		errChan <- errors.New("fiber stopped")
	}()

	return errChan
}
