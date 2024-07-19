// Package poolsvc provides a wrapper around the github.com/alitto/pond worker pool
// with additional retry functionality and integration with Wire for dependency injection.
//
// The package offers a simple interface to create and manage a worker pool,
// allowing tasks to be submitted with a retry mechanism in case of submission failures.
package poolsvc

import (
	"time"

	"github.com/alitto/pond"
	"github.com/google/wire"
	"go.uber.org/zap"
)

// DefaultWireset is a Wire provider set that can be used to inject a PoolSvc
// and its configuration into a larger application.
var DefaultWireset = wire.NewSet(
	NewPool,
	DefaultPoolConfig,
)

// PoolSvc represents a worker pool service that wraps a pond.WorkerPool
// and provides additional functionality.
type PoolSvc struct {
	pool   *pond.WorkerPool
	logger *zap.Logger
	config *PoolConfig
}

// PoolConfig holds the configuration options for the PoolSvc.
type PoolConfig struct {
	// MaxWorkers is the maximum number of worker goroutines in the pool.
	MaxWorkers int

	// MaxCapacity is the maximum number of tasks that can be queued.
	MaxCapacity int

	// MaxRetries is the maximum number of retry attempts for task submission.
	MaxRetries int

	// RetryDelay is the duration to wait between retry attempts.
	RetryDelay time.Duration
}

// DefaultPoolConfig returns a PoolConfig with predefined default values.
// These values can be overridden by the user if needed.
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxWorkers:  50,
		MaxCapacity: 100,
		MaxRetries:  3,
		RetryDelay:  time.Second,
	}
}

// NewPool creates a new PoolSvc with the given logger and configuration.
// It returns the PoolSvc, a cleanup function, and an error (always nil in the current implementation).
// The cleanup function should be called when the PoolSvc is no longer needed to properly release resources.
func NewPool(
	logSvc *zap.Logger,
	config *PoolConfig,
) (*PoolSvc, func(), error) {
	logger := logSvc.With(zap.String("component", "PoolSvc"))
	pool := pond.New(config.MaxWorkers, config.MaxCapacity)

	poolSvc := &PoolSvc{
		pool:   pool,
		logger: logger,
		config: config,
	}

	cleanup := func() {
		logger.Info("stopping pool")
		pool.StopAndWait()
		logger.Info("pool stopped")
	}

	return poolSvc, cleanup, nil
}

// TrySubmitWithRetry attempts to submit a task to the worker pool.
// If the submission fails, it will retry up to the configured maximum number of retries.
// It returns true if the task was successfully submitted, false otherwise.
func (s *PoolSvc) TrySubmitWithRetry(task func()) bool {
	for i := 0; i < s.config.MaxRetries; i++ {
		if s.pool.TrySubmit(task) {
			return true
		}
		if i < s.config.MaxRetries-1 {
			s.logger.Debug("Task submission failed, retrying", zap.Int("attempt", i+1))
			time.Sleep(s.config.RetryDelay)
		}
	}
	s.logger.Warn("Task submission failed after max retries", zap.Int("maxRetries", s.config.MaxRetries))
	return false
}
