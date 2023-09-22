package redissvc

import (
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/garsue/watermillzap"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"os"
	"time"
)

var DefaultWireset = wire.NewSet(
	NewRedisClient,
	NewSubscriber,
	NewPublisher,
	wire.NewSet(new(redisstream.Subscriber), new(message.Subscriber)),
	wire.NewSet(new(redisstream.Publisher), new(message.Publisher)),
)

func RedisConfigFromEnv() (*redis.Options, error) {
	uri, ok := os.LookupEnv("REDIS_URI")
	if !ok {
		return nil, errors.New("REDIS_URI is not set")
	}

	opt, err := redis.ParseURL(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse REDIS_URI")
	}

	return opt, nil
}
func NewRedisClient(
	logSvc *zap.Logger,
	config *redis.Options,
) (*redis.Client, func(), error) {
	logger := logSvc.With(zap.Strings("tags", []string{"redis-client"}))
	redisClient := redis.NewClient(config).WithTimeout(20 * time.Second)

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
