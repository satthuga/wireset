package server

import (
	"api/pkg/configsvc"
	"api/pkg/feature/core"
	"api/pkg/pubsub"
	"api/pkg/tracersvc"
	"context"
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

var Wireset = wire.NewSet(
	wire.Struct(new(Server), "*"),
	core.DefaultWireset,
)

func (s *Server) Start(ctx context.Context) chan error {
	if s.ConfigSvc.IsProduction() {
		s.TracerSvc.Start()
	}

	s.CoreFeature.Register()
	for _, feature := range s.Features {
		feature.Register()
		s.LogSvc.Info("feature registered", zap.String("name", feature.GetName()))
	}

	errChan := make(chan error, 1)

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
