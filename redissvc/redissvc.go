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
	"strconv"
)

var DefaultWireset = wire.NewSet(
	NewRedisClient,
	NewSubscriber,
	NewPublisher,
	wire.NewSet(new(redisstream.Subscriber), new(message.Subscriber)),
	wire.NewSet(new(redisstream.Publisher), new(message.Publisher)),
)

func RedisConfigFromEnv() (*redis.Options, error) {
	addr, ok := os.LookupEnv("REDIS_ADDRESS")
	if !ok {
		return nil, errors.New("REDIS_ADDR is not set")
	}

	username, ok := os.LookupEnv("REDIS_USERNAME")
	if !ok {
		return nil, errors.New("REDIS_USERNAME is not set")
	}

	password, ok := os.LookupEnv("REDIS_PASSWORD")
	if !ok {
		return nil, errors.New("REDIS_PASSWORD is not set")
	}

	redisDb, ok := os.LookupEnv("REDIS_DB")
	if !ok {
		return nil, errors.New("REDIS_DB is not set")
	}
	redisDbNumber, err := strconv.Atoi(redisDb)
	if err != nil {
		return nil, errors.Wrap(err, "REDIS_DB is not a number")
	}

	return &redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       redisDbNumber,
	}, nil
}
func NewRedisClient(
	logSvc *zap.Logger,
	config *redis.Options,
) (*redis.Client, func(), error) {
	logger := logSvc.With(zap.Strings("tags", []string{"redis-client"}))
	redisClient := redis.NewClient(config)

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
