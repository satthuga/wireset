package poolsvc

import (
	"github.com/alitto/pond"
	"github.com/google/wire"
	"go.uber.org/zap"
)

var DefaultWireset = wire.NewSet(
	NewPool,
)

type PoolSvc struct {
	pool *pond.WorkerPool
}

func NewPool(
	logSvc *zap.Logger,
) (*PoolSvc, func(), error) {
	logger := logSvc.With(zap.Strings("tags", []string{"PoolSvc"}))
	pool := pond.New(10, 20)
	cleanup := func() {
		logger.Info("stopping pool")
		pool.StopAndWait()
		logger.Info("pool stopped")
	}

	poolSvc := &PoolSvc{
		pool: pool,
	}

	return poolSvc, cleanup, nil
}

func (s *PoolSvc) Submit(task func()) {
	s.pool.Submit(task)
}
