package cachesvc

import (
	"github.com/dgraph-io/ristretto"
	"github.com/google/wire"
)

var DefaultWireset = wire.NewSet(NewCacheService)

func NewCacheService() (*ristretto.Cache, func(), error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})

	cleanup := func() {
		cache.Close()
	}

	return cache, cleanup, err
}
