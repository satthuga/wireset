package pubsub

import (
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/aiocean/wireset/configsvc"
	"github.com/garsue/watermillzap"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var RedisWireset = wire.NewSet(
	NewRedisPublisher,
	NewRedisSubscriber,
	wire.Bind(new(message.Subscriber), new(*redisstream.Subscriber)),
	wire.Bind(new(message.Publisher), new(*redisstream.Publisher)),
)

func NewRedisSubscriber(subClient *redis.Client, logger *zap.Logger, globalConfig *configsvc.ConfigService) (*redisstream.Subscriber, func(), error) {
	subscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        subClient,
			Unmarshaller:  redisstream.DefaultMarshallerUnmarshaller{},
			ConsumerGroup: "consumer_group_" + globalConfig.ServiceName,
		},
		watermillzap.NewLogger(logger.Named("subscriber")),
	)

	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		logger.Info("Router: Cleaning up")
		if err := subscriber.Close(); err != nil {
			logger.Error("Router: error closing router", zap.Error(err))
			return
		}
		logger.Info("Router: router closed")
	}

	return subscriber, cleanup, nil
}

func NewRedisPublisher(pubClient *redis.Client, logger *zap.Logger) (*redisstream.Publisher, func(), error) {
	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client:     pubClient,
			Marshaller: redisstream.DefaultMarshallerUnmarshaller{},
		},
		watermillzap.NewLogger(logger.Named("publisher")),
	)

	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		logger.Info("Router: Cleaning up")
		if err := publisher.Close(); err != nil {
			logger.Error("Router: error closing router", zap.Error(err))
			return
		}
		logger.Info("Router: router closed")
	}

	return publisher, cleanup, nil
}
