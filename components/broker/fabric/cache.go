package fabric

import (
	"fmt"
	"sync"
	"time"

	"github.com/xigxog/kubefox/libs/core/logger"
)

type cache[T any] struct {
	m     map[string]*item[T]
	mTTL  int64
	iTTL  int64
	log   *logger.Log
	tick  *time.Ticker
	mutex sync.RWMutex
}

type item[T any] struct {
	value     T
	key       string
	immutable bool
	aTime     int64
	cTime     int64
}

func NewCache[T any](mTTL, iTTL int64, log *logger.Log) *cache[T] {
	c := &cache[T]{
		m:    make(map[string]*item[T]),
		mTTL: mTTL,
		log:  log,
		tick: time.NewTicker(time.Duration(mTTL * int64(time.Second))),
		iTTL: iTTL,
	}
	go c.sweep()

	return c
}

func (c *cache[T]) GetItem(key string) *item[T] {
	return c.m[key]
}

func (c *cache[T]) SetItem(key string, it *item[T]) {
	c.mutex.Lock()
	it.key = key
	c.m[key] = it
	c.mutex.Unlock()
}

func (it *item[T]) String() string {
	return fmt.Sprintf("key: %s, immutable: %t, ctime: %d, atime: %d", it.key, it.immutable, it.cTime, it.aTime)
}

// sweep deletes items from the cache that have expired.
func (c *cache[_]) sweep() {
	for range c.tick.C {
		c.mutex.Lock()
		now := time.Now().Unix()
		for k, v := range c.m {
			if v.immutable && (now-v.aTime) > iTTL ||
				!v.immutable && (now-v.cTime) > mTTL {
				c.log.Debugf("deleting expired item; %s", v)
				delete(c.m, k)
			}
		}
		c.mutex.Unlock()
	}
}
