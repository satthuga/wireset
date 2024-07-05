package redissvc

import (
	"context"
	"github.com/google/wire"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"os"
	"time"
)

var DefaultWireset = wire.NewSet(
	NewRedisClient,
)

var EnvWireset = wire.NewSet(
	NewRedisClient,
	RedisConfigFromEnv,
)

func getEnvVar(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", errors.Errorf("%s is not set", key)
	}
	return value, nil
}

func RedisConfigFromEnv() (*redis.Options, error) {
	uri, err := getEnvVar("REDIS_URI")
	if err != nil {
		return nil, err
	}

	opt, err := redis.ParseURL(uri)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse REDIS_URI")
	}

	return opt, nil
}

func NewRedisClient(
	ctx context.Context,
	logSvc *zap.Logger,
	config *redis.Options,
) (*redis.Client, func(), error) {
	logger := logSvc.With(zap.Strings("tags", []string{"redis-client"}))
	redisClient := redis.NewClient(config).WithTimeout(20 * time.Second)

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to connect to Redis")
	}

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
