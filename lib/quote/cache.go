package quote

import (
	"context"
	"errors"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

// Cache defines the interface for caching security ID lookups.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
}

type lruCache struct {
	lc *lru.Cache[string, string]
}

func (c *lruCache) Get(_ context.Context, key string) (string, error) {
	val, ok := c.lc.Get(key)
	if !ok {
		return "", ErrCacheMiss
	}
	return val, nil
}

func (c *lruCache) Set(_ context.Context, key string, value string) error {
	c.lc.Add(key, value)
	return nil
}

type redisCache struct {
	rdb *redis.Client
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	s, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		err = ErrCacheMiss
	}
	return s, err
}

func (c *redisCache) Set(ctx context.Context, key string, value string) error {
	return c.rdb.Set(ctx, key, value, 48*time.Hour).Err()
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
