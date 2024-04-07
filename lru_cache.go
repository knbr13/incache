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
