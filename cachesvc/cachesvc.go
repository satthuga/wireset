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

// NewCacheService creates a new CacheService, a cleanup function, and returns an error if any.
// The cleanup function should be deferred to ensure the cache is properly closed when no longer needed.
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

// Get retrieves a value from the cache using a key. It returns the value and a boolean indicating if the key was found.
func (s *CacheService) Get(key string) (interface{}, bool) {
	return s.cache.Get(key)
}

// Set adds a value to the cache with a specified key. It returns a boolean indicating if the operation was successful.
func (s *CacheService) Set(key string, value interface{}) bool {
	return s.cache.Set(key, value, 0)
}

// SetWithTTL adds a value to the cache with a specified key and a time-to-live duration.
// It returns a boolean indicating if the operation was successful.
func (s *CacheService) SetWithTTL(key string, value interface{}, ttl time.Duration) bool {
	return s.cache.SetWithTTL(key, value, 0, ttl)
}
