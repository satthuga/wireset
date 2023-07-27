package router

import (
	"github.com/aiocean/wireset/configsvc"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/garsue/watermillzap"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var DefaultWireset = wire.NewSet(
	NewRouter,
)

func NewRouter(
	logSvc *zap.Logger,
	cfg *configsvc.ConfigService,
) (*message.Router, func(), error) {
	logger := logSvc.With(zap.Strings("tags", []string{"Router"}))
	waterLogger := watermillzap.NewLogger(logger)

	router, err := message.NewRouter(message.RouterConfig{}, waterLogger)
	if err != nil {
		return nil, nil, err
	}

	router.AddMiddleware(
		middleware.Recoverer,
		middleware.CorrelationID,
		Retry{
			MaxRetries:      2,
			InitialInterval: time.Second * 1,
			Logger:          waterLogger,
			OnFailed: func(msg *message.Message, err error) ([]*message.Message, error) {
				// save event to collection
				logger.Error("Router: error handling message", zap.Error(err))
				return nil, nil
			},
		}.Middleware,
	)

	cleanup := func() {
		logger.Info("Router: Cleaning up")
		if err := router.Close(); err != nil {
			logger.Error("Router: error closing router", zap.Error(err))
			return
		}
		logger.Info("Router: router closed")

	}

	return router, cleanup, nil
}
