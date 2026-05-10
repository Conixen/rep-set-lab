package cache

import (
	"context"
	"sync"
	"time"
)

// GetWithCache is a generic cache-aside helper ready for Redis integration.
// Wire get/set to a Redis client when you add Redis; pass nil to skip caching.
//
// Carrying this pattern from FastTrack: one generic function eliminates
// per-entity cache duplication and makes hit/miss behaviour consistent.
func GetWithCache[T any](
	ctx context.Context,
	mu *sync.Mutex,
	get func(ctx context.Context, key string) (T, bool),
	set func(ctx context.Context, key string, val T, ttl time.Duration),
	key string,
	ttl time.Duration,
	getFromStore func(ctx context.Context) (T, error),
) (T, error) {
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}
	if get != nil {
		if val, ok := get(ctx, key); ok {
			return val, nil
		}
	}
	val, err := getFromStore(ctx)
	if err != nil {
		return val, err
	}
	if set != nil {
		set(ctx, key, val, ttl)
	}
	return val, nil
}
