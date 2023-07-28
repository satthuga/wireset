package redissvc

import (
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/garsue/watermillzap"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var DefaultWireset = wire.NewSet(
	NewRedisClient,
	NewSubscriber,
	NewPublisher,
	wire.NewSet(new(redisstream.Subscriber), new(message.Subscriber)),
	wire.NewSet(new(redisstream.Publisher), new(message.Publisher)),
)

func NewRedisClient(
	logSvc *zap.Logger,
) (*redis.Client, func(), error) {

	logger := logSvc.With(zap.Strings("tags", []string{"redis-client"}))

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "containers-us-west-78.railway.app:5482",
		Username: "default",
		Password: "NtSp7hZLEfC7fYUp2D7f",
		DB:       0,
	})

	cleanup := func() {
		logger.Info("Router: Cleaning up")
		if err := redisClient.Close(); err != nil {
			logger.Error("Router: error closing router", zap.Error(err))
			return
		}
		logger.Info("Router: router closed")
	}

	return redisClient, cleanup, nil
}

func NewSubscriber(subClient *redis.Client, logger *zap.Logger) (*redisstream.Subscriber, func(), error) {
	subscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:        subClient,
			Unmarshaller:  redisstream.DefaultMarshallerUnmarshaller{},
			ConsumerGroup: "test_consumer_group",
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

func NewPublisher(pubClient *redis.Client, logger *zap.Logger) (*redisstream.Publisher, func(), error) {
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
