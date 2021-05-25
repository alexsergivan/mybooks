package cache

import (
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
)

var cache *Ristretto

var onceCache sync.Once

const RistrettoCacheTTL = 10 * time.Minute

type Ristretto struct {
	cache *ristretto.Cache
}

func NewRistrettoCache() *Ristretto {
	onceCache.Do(func() {
		ristrettoCache, _ := ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e7,     // Num keys to track frequency of (10M).
			MaxCost:     1 << 30, // Maximum cost of cache (1GB).
			BufferItems: 64,      // Number of keys per Get buffer.
		})
		cache = &Ristretto{
			cache: ristrettoCache,
		}
	})

	return cache
}

func (r *Ristretto) Get(key string) (interface{}, bool) {
	return r.cache.Get(key)
}

// Set sets data to cache with specific ttl. If ttl == -1, default cache ttl value will be used.
func (r *Ristretto) Set(key string, value interface{}, ttl time.Duration) {
	if ttl == -1 {
		ttl = RistrettoCacheTTL
	}
	r.cache.SetWithTTL(key, value, 1, ttl)
}
