package cache

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type Cache[T any] interface {
	Get(key string) (T, bool)
	Set(key string, value T)
	Delete(key string)
}

type cache[T any] struct {
	m     map[string]*item[T]
	ttl   time.Duration
	tick  *time.Ticker
	mutex sync.RWMutex
}

type item[T any] struct {
	Key   string
	Value T
	aTime int64
	cTime int64
}

func New[T any](ttl time.Duration) Cache[T] {
	// Ensure sweep does not run constantly if TTL is short.
	tickD := time.Duration(math.Max(float64(ttl), float64(time.Second)))
	c := &cache[T]{
		m:    make(map[string]*item[T]),
		ttl:  ttl,
		tick: time.NewTicker(tickD),
	}
	go c.sweep()

	return c
}

func (c *cache[T]) Get(key string) (T, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var val T
	itm, found := c.m[key]
	if found && itm != nil {
		itm.aTime = time.Now().Unix()
		val = itm.Value
	}

	return val, found
}

func (c *cache[T]) Set(key string, value T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now().Unix()
	c.m[key] = &item[T]{
		Key:   key,
		Value: value,
		aTime: now,
		cTime: now,
	}
}

func (c *cache[T]) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.m, key)
}

func (it *item[T]) String() string {
	return fmt.Sprintf("key: %s, ctime: %d, atime: %d", it.Key, it.cTime, it.aTime)
}

// sweep deletes items from the cache that have expired.
func (c *cache[_]) sweep() {
	for range c.tick.C {
		c.mutex.Lock()
		now := time.Now().Unix()
		for k, v := range c.m {
			if (now - v.aTime) >= int64(c.ttl.Seconds()) {
				delete(c.m, k)
			}
		}
		c.mutex.Unlock()
	}
}
