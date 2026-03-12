package quote

import (
	"context"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/redis/go-redis/v9"
)

// Cache defines the interface for caching security ID lookups.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
}

type lruCache struct {
	lc *lru.Cache[string, string]
}

func (c *lruCache) Get(ctx context.Context, key string) (string, error) {
	val, ok := c.lc.Get(key)
	if !ok {
		return "", redis.Nil
	}
	return val, nil
}

func (c *lruCache) Set(ctx context.Context, key string, value string) error {
	c.lc.Add(key, value)
	return nil
}

type redisCache struct {
	rdb *redis.Client
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *redisCache) Set(ctx context.Context, key string, value string) error {
	return c.rdb.Set(ctx, key, value, 0).Err()
}

// newCache creates a Redis-backed cache. If rdb is nil, it creates an in-memory
// LRU cache instead.
func newCache(rdb *redis.Client) Cache {
	if rdb == nil {
		c, _ := lru.New[string, string](256)
		return &lruCache{c}
	}
	return &redisCache{rdb: rdb}
}
