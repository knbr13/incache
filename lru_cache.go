package incache

import (
	"container/list"
	"time"
)

type lruItem[K comparable, V any] struct {
	key      K
	value    V
	expireAt *time.Time
}

type LRUCache[K comparable, V any] struct {
	baseCache
	m            map[K]*list.Element // where the key-value pairs are stored
	evictionList *list.List
}

func newLRU[K comparable, V any](cacheBuilder *CacheBuilder[K, V]) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		baseCache: baseCache{
			size: cacheBuilder.size,
		},
		m:            make(map[K]*list.Element),
		evictionList: list.New(),
	}
}

func (c *LRUCache[K, V]) Get(k K) (v V, b bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.m[k]
	if !ok {
		return
	}

	if item.Value.(*lruItem[K, V]).expireAt != nil && item.Value.(*lruItem[K, V]).expireAt.Before(time.Now()) {
		delete(c.m, k)
		c.evictionList.Remove(item)
		return
	}

	return item.Value.(*lruItem[K, V]).value, true
}

func (c *LRUCache[K, V]) GetAll() map[K]V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	m := make(map[K]V)
	for k, v := range c.m {
		if v.Value.(*lruItem[K, V]).expireAt != nil && v.Value.(*lruItem[K, V]).expireAt.Before(time.Now()) {
			m[k] = v.Value.(*lruItem[K, V]).value
		}
	}

	return m
}

func (c *LRUCache[K, V]) Set(k K, v V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.set(k, v, nil)
}

func (c *LRUCache[K, V]) SetWithTimeout(k K, v V, t time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.set(k, v, nil)
}

func (c *LRUCache[K, V]) set(k K, v V, exp *time.Duration) {
	item, ok := c.m[k]
	tm := time.Now().Add(*exp)
	if ok {
		item.Value.(*lruItem[K, V]).value = v
		item.Value.(*lruItem[K, V]).expireAt = &tm
		c.evictionList.MoveToFront(item)
	} else {
		if len(c.m) == int(c.size) {
			c.evict(1)
		}

		c.m[k] = &list.Element{
			Value: &lruItem[K, V]{
				key:      k,
				value:    v,
				expireAt: &tm,
			},
		}
		c.evictionList.PushFront(c.m[k])
	}
}

func (c *LRUCache[K, V]) evict(i int) {
	for j := 0; j < i; j++ {
		if b := c.evictionList.Back(); b != nil {
			c.evictionList.Remove(b)
			delete(c.m, b.Value.(lruItem[K, V]).key)
		} else {
			return
		}
	}
}
