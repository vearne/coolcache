package coolcache

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	slog "github.com/vearne/simplelog"
	"hash/fnv"
	"math/rand"
	"time"
)

type Cache struct {
	Kind string
	Cap  uint64
	// shard number
	ShardNumber         uint64
	shards              []*ShardCache
	ExpireCleanInterval time.Duration
	ExitChan            chan struct{}
	CallBackFunc        CallBackFunc
	// prometheus
	PromReqTotal *prometheus.CounterVec
	PromSize     prometheus.GaugeFunc
}

func NewCache(opts ...Option) *Cache {
	c := Cache{
		Kind:                "default",
		Cap:                 1000000,
		ShardNumber:         100,
		ExpireCleanInterval: 100 * time.Millisecond,
		ExitChan:            make(chan struct{}),
	}

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		opt(&c)
	}

	// init shard slice
	c.shards = make([]*ShardCache, c.ShardNumber)
	for i := 0; i < int(c.ShardNumber); i++ {
		c.shards[i] = NewShardCache(c.Cap/c.ShardNumber, c.CallBackFunc)
	}

	// prometheus metrics
	c.PromReqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        PromMetricCacheRequestTotalName,
			Help:        PromMetricCacheRequestTotalHelp,
			ConstLabels: map[string]string{"kind": c.Kind},
		},
		[]string{"state"},
	)

	c.PromSize = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name:        PromMetricCacheSizeName,
			Help:        PromMetricCacheSizeHelp,
			ConstLabels: map[string]string{"kind": c.Kind},
		},
		func() float64 { return float64(c.Size()) },
	)
	prometheus.MustRegister(c.PromReqTotal, c.PromSize)

	// cleaner
	// periodic execution
	go c.Clean()

	return &c
}

func (c *Cache) Get(key any) any {
	value := c.get(key)
	c.PromReqTotal.With(prometheus.Labels{
		"state": All,
	}).Inc()
	if value != nil {
		c.PromReqTotal.With(prometheus.Labels{
			"state": Hit,
		}).Inc()
	}
	return value
}

func (c *Cache) get(key any) any {
	hashCode := c.Sum64(key)
	return c.shards[hashCode%c.ShardNumber].Get(key)
}

func (c *Cache) Remove(key any) any {
	hashCode := c.Sum64(key)
	return c.shards[hashCode%c.ShardNumber].Remove(key)
}

// If <duration> <=0 means it does not expire.
func (c *Cache) Set(key any, value any, duration time.Duration) {
	hashCode := c.Sum64(key)
	c.shards[hashCode%c.ShardNumber].Set(key, value, duration)
}

func (c *Cache) Close() {
	close(c.ExitChan)
	for _, sc := range c.shards {
		sc.Close()
	}
}

func (c *Cache) Sum64(key any) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(fmt.Sprintf("%v", key)))
	return hash.Sum64()
}

func (c *Cache) Size() int {
	total := 0
	for _, sc := range c.shards {
		total += sc.Size()
	}
	return total
}

func (c *Cache) Clean() {
	ticker := time.NewTicker(c.ExpireCleanInterval)
	for {
		select {
		case <-ticker.C:
			idx := rand.Uint64() % c.ShardNumber
			count := c.shards[idx].Clean()
			slog.Debug("Clean: shards[%d], count:%v", idx, count)
		case <-c.ExitChan:
			// do some clean task
			slog.Debug("cache:%v exit...", c.Kind)
			return
		}
	}
}
