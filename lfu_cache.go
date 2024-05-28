package incache

import (
	"container/list"
	"sync"
)

type LFUCache[K comparable, V any] struct {
	mu           sync.RWMutex
	size         uint
	m            map[K]*list.Element
	evictionList *list.List
}
