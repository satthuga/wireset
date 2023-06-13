package cachesvc

import (
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(NewCacheService)

type CacheService struct {
	cache *ristretto.Cache
}

func NewCacheService() (*CacheService, func(), error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})

	cleanup := func() {
		cache.Close()
	}

	cacheSvc := &CacheService{
		cache: cache,
	}

	return cacheSvc, cleanup, err
}

func (s *CacheService) Get(key string) (interface{}, bool) {
	return s.cache.Get(key)
}

func (s *CacheService) Set(key string, value interface{}) bool {
	return s.cache.Set(key, value, 0)
}

func (s *CacheService) SetWithTTL(key string, value interface{}, ttl time.Duration) bool {
	return s.cache.SetWithTTL(key, value, 0, ttl)
}
