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
	NewGoChannel,
)

// for golang, we can use the same channel for both publisher and subscriber

func NewGoChannel(logger *zap.Logger) (*gochannel.GoChannel, func(), error) {
	channel := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer:            100,
			BlockPublishUntilSubscriberAck: false,
		},
		watermillzap.NewLogger(logger.Named("channel")),
	)
	cleanup := func() {
		if err := channel.Close(); err != nil {
			logger.Error("Router: error closing router", zap.Error(err))
			return
		}
		logger.Info("Router: router closed")
	}

	return channel, cleanup, nil
}

func NewGoroutineSubscriber(logger *zap.Logger, channel *gochannel.GoChannel) (message.Subscriber, func(), error) {
	cleanup := func() {
		logger.Info("Router: Cleaning up")
	}
	return channel, cleanup, nil
}

func NewGoroutinePublisher(logger *zap.Logger, channel *gochannel.GoChannel) (message.Publisher, func(), error) {

	cleanup := func() {
		logger.Info("Router: Cleaning up")
	}

	return channel, cleanup, nil
}
