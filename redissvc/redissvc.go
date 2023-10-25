package redissvc

import (
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
