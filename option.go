package coolcache

import (
	"time"
)

type CallBackFunc func(key, value any)
type Option func(cache *Cache)

func WithKind(kind string) Option {
	return func(c *Cache) {
		c.Kind = kind
	}
}

func WithCapacity(cap uint64) Option {
	return func(c *Cache) {
		c.Cap = cap
	}
}

func WithShardNumber(num uint64) Option {
	return func(c *Cache) {
		c.ShardNumber = num
	}
}

func WithExpireCleanInterval(interval time.Duration) Option {
	return func(c *Cache) {
		c.ExpireCleanInterval = interval
	}
}

func WithCallBackFunc(f CallBackFunc) Option {
	return func(c *Cache) {
		c.CallBackFunc = f
	}
}
