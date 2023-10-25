package pubsub

import (
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/garsue/watermillzap"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var GoroutineWireset = wire.NewSet(
	NewGoroutinePublisher,
	NewGoroutineSubscriber,
)

func NewGoroutineSubscriber(logger *zap.Logger) (message.Subscriber, func(), error) {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermillzap.NewLogger(logger.Named("subscriber")),
	)

	cleanup := func() {
		logger.Info("Router: Cleaning up")
		if err := pubSub.Close(); err != nil {
			logger.Error("Router: error closing router", zap.Error(err))
			return
		}
		logger.Info("Router: router closed")
	}

	return pubSub, cleanup, nil
}

func NewGoroutinePublisher(logger *zap.Logger) (message.Publisher, func(), error) {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermillzap.NewLogger(logger.Named("publisher")),
	)

	cleanup := func() {
		logger.Info("Router: Cleaning up")
		if err := pubSub.Close(); err != nil {
			logger.Error("Router: error closing router", zap.Error(err))
			return
		}
		logger.Info("Router: router closed")
	}

	return pubSub, cleanup, nil
}
