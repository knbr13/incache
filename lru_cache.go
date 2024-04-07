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

	c.set(k, v, 0)
}

func (c *LRUCache[K, V]) SetWithTimeout(k K, v V, t time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.set(k, v, t)
}

func (c *LRUCache[K, V]) NotFoundSet(k K, v V) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.m[k]
	if ok {
		return false
	}

	c.set(k, v, 0)
	return true
}

func (c *LRUCache[K, V]) NotFoundSetWithTimeout(k K, v V, t time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.m[k]
	if ok {
		return false
	}

	c.set(k, v, t)
	return true
}

func (c *LRUCache[K, V]) Delete(k K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.delete(k)
}

func (c *LRUCache[K, V]) delete(k K) {
	item, ok := c.m[k]
	if !ok {
		return
	}

	delete(c.m, k)
	c.evictionList.Remove(item)
}

func (src *LRUCache[K, V]) TransferTo(dst Cache[K, V]) {
	src.mu.Lock()
	defer src.mu.Unlock()

	for k, v := range src.m {
		if v.Value.(*lruItem[K, V]).expireAt != nil && v.Value.(*lruItem[K, V]).expireAt.Before(time.Now()) {
			src.delete(k)
			dst.Set(k, v.Value.(*lruItem[K, V]).value)
		}
	}
}

func (src *LRUCache[K, V]) CopyTo(dst Cache[K, V]) {
	src.mu.Lock()
	defer src.mu.Unlock()

	for k, v := range src.m {
		if v.Value.(*lruItem[K, V]).expireAt != nil && v.Value.(*lruItem[K, V]).expireAt.Before(time.Now()) {
			dst.Set(k, v.Value.(*lruItem[K, V]).value)
		}
	}
}

func (c *LRUCache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.m))
	for k := range c.m {
		keys = append(keys, k)
	}

	return keys
}

func (c *LRUCache[K, V]) set(k K, v V, exp time.Duration) {
	item, ok := c.m[k]
	var tm *time.Time
	if exp > 0 {
		t := time.Now().Add(exp)
		tm = &t
	}
	if ok {
		item.Value.(*lruItem[K, V]).value = v
		item.Value.(*lruItem[K, V]).expireAt = tm
		c.evictionList.MoveToFront(item)
	} else {
		if len(c.m) == int(c.size) {
			c.evict(1)
		}

		c.m[k] = &list.Element{
			Value: &lruItem[K, V]{
				key:      k,
				value:    v,
				expireAt: tm,
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
