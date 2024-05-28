package incache

import (
	"container/list"
	"sync"
	"time"
)

type LFUCache[K comparable, V any] struct {
	mu           sync.RWMutex
	size         uint
	m            map[K]*list.Element
	evictionList *list.List
}

func NewLFU[K comparable, V any](size uint) *LFUCache[K, V] {
	return &LFUCache[K, V]{
		size:         size,
		m:            make(map[K]*list.Element),
		evictionList: list.New(),
	}
}

type lfuItem[K comparable, V any] struct {
	key      K
	value    V
	freq     uint
	expireAt *time.Time
}
