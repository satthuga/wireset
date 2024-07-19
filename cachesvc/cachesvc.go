package cachesvc

import (
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/google/wire"
)

// CacheConfig holds the configuration for the cache service
type CacheConfig struct {
	NumCounters int64
	MaxCost     int64
	BufferItems int64
	DefaultTTL  time.Duration
}

// ProvideCacheConfig returns a default cache configuration
func ProvideCacheConfig() *CacheConfig {
	return &CacheConfig{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
		DefaultTTL:  24 * time.Hour, // Example default TTL
	}
}

// CacheService wraps the ristretto cache and includes configuration
type CacheService struct {
	cache  *ristretto.Cache
	config *CacheConfig
}

// NewCacheService creates a new CacheService with the given configuration
func NewCacheService(config *CacheConfig) (*CacheService, func(), error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: config.NumCounters,
		MaxCost:     config.MaxCost,
		BufferItems: config.BufferItems,
	})

	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		cache.Close()
	}

	cacheSvc := &CacheService{
		cache:  cache,
		config: config,
	}

	return cacheSvc, cleanup, nil
}

// Get retrieves a value from the cache using a key
func (s *CacheService) Get(key string) (interface{}, bool) {
	return s.cache.Get(key)
}

// Set adds a value to the cache with a specified key, using the default TTL
func (s *CacheService) Set(key string, value interface{}) bool {
	return s.cache.SetWithTTL(key, value, 0, s.config.DefaultTTL)
}

// SetWithTTL adds a value to the cache with a specified key and TTL
func (s *CacheService) SetWithTTL(key string, value interface{}, ttl time.Duration) bool {
	return s.cache.SetWithTTL(key, value, 0, ttl)
}

// DefaultWireSet provides the set of providers for wire
var DefaultWireset = wire.NewSet(
	ProvideCacheConfig,
	NewCacheService,
)
