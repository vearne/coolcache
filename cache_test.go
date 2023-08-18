package coolcache

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	cache := NewCache(WithName("test1"))
	for i := 0; i < 100; i++ {
		key := strconv.Itoa(i)
		value := key
		cache.Set(key, value, 0)
	}
	for i := 0; i < 100; i++ {
		key := strconv.Itoa(i)
		value := key
		assert.Equal(t, value, cache.get(key))
	}
}
func TestSetRemoveGet(t *testing.T) {
	cache := NewCache(WithName("test2"))
	for i := 0; i < 100; i++ {
		key := strconv.Itoa(i)
		value := key
		cache.Set(key, value, 0)
	}
	cache.Remove("50")
	for i := 0; i < 100; i++ {
		key := strconv.Itoa(i)
		value := key
		if i == 50 {
			continue
		}
		assert.Equal(t, value, cache.get(key))
	}
	assert.Equal(t, nil, cache.get("50"))
	assert.Equal(t, 99, cache.Size())
}

func TestExire(t *testing.T) {
	cache := NewCache(WithName("test3"))
	for i := 0; i < 100; i++ {
		key := strconv.Itoa(i)
		value := key
		cache.Set(key, value, time.Millisecond*10)
	}

	time.Sleep(20 * time.Millisecond)
	for i := 0; i < 100; i++ {
		key := strconv.Itoa(i)
		assert.Equal(t, nil, cache.get(key))
	}
	assert.Equal(t, 0, cache.Size())
}
