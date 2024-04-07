package incache

import (
	"container/list"
	"time"
)

type LRUCache[K comparable, V any] struct {
	baseCache
	m            map[K]lruItem[V] // where the key-value pairs are stored
	evictionList *list.List
}

type lruItem[V any] struct {
	value    *list.Element
	expireAt *time.Time
}
