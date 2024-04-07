package incache

import (
	"container/list"
	"time"
)

type lruItem[V any] struct {
	value    *list.Element
	expireAt *time.Time
}

type LRUCache[K comparable, V any] struct {
	baseCache
	m            map[K]lruItem[V] // where the key-value pairs are stored
	evictionList *list.List
}

func newLRU[K comparable, V any](cacheBuilder *CacheBuilder[K, V]) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		baseCache: baseCache{
			size: cacheBuilder.size,
		},
		m:            make(map[K]lruItem[V]),
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

	if item.expireAt != nil && item.expireAt.Before(time.Now()) {
		delete(c.m, k)
		c.evictionList.Remove(item.value)
		return
	}

	return item.value.Value.(V), true
}

func (c *LRUCache[K, V]) GetAll() map[K]V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	m := make(map[K]V)
	for k, v := range c.m {
		if v.expireAt == nil || !v.expireAt.Before(time.Now()) {
			m[k] = v.value.Value.(V)
		}
	}

	return m
}
