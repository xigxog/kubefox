package utils

import (
	"fmt"
	"sync"
	"time"
)

type CacheLogger interface {
	Debugf(template string, args ...interface{})
}

type Cache[T any] struct {
	m     map[string]*item[T]
	mTTL  int64
	iTTL  int64
	log   CacheLogger
	tick  *time.Ticker
	mutex sync.RWMutex
}

type item[T any] struct {
	Key       string
	Value     T
	Immutable bool
	aTime     int64
	cTime     int64
}

func NewCache[T any](mTTL, iTTL int64, log CacheLogger) *Cache[T] {
	c := &Cache[T]{
		m:    make(map[string]*item[T]),
		mTTL: mTTL,
		log:  log,
		tick: time.NewTicker(time.Duration(mTTL * int64(time.Second))),
		iTTL: iTTL,
	}
	go c.sweep()

	return c
}

func (c *Cache[T]) Get(key string) T {
	var val T
	if itm := c.m[key]; itm != nil {
		itm.aTime = time.Now().Unix()
		val = itm.Value
	}

	return val
}

func (c *Cache[T]) Set(key string, value T, immutable bool) {
	c.mutex.Lock()
	now := time.Now().Unix()
	c.m[key] = &item[T]{
		Key:       key,
		Value:     value,
		Immutable: immutable,
		aTime:     now,
		cTime:     now,
	}
	c.mutex.Unlock()
}

func (it *item[T]) String() string {
	return fmt.Sprintf("key: %s, immutable: %t, ctime: %d, atime: %d", it.Key, it.Immutable, it.cTime, it.aTime)
}

// sweep deletes items from the cache that have expired.
func (c *Cache[_]) sweep() {
	for range c.tick.C {
		c.mutex.Lock()
		now := time.Now().Unix()
		for k, v := range c.m {
			if v.Immutable && (now-v.aTime) > c.iTTL ||
				!v.Immutable && (now-v.cTime) > c.mTTL {
				if c.log != nil {
					c.log.Debugf("deleting expired item; %s", v)
				}
				delete(c.m, k)
			}
		}
		c.mutex.Unlock()
	}
}
