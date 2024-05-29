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

func (l *LFUCache[K, V]) evict(n int) {
	for i := 0; i < n; i++ {
		if b := l.evictionList.Back(); b != nil {
			delete(l.m, b.Value.(*lfuItem[K, V]).key)
			l.evictionList.Remove(b)
		} else {
			return
		}
	}
}

func (l *LFUCache[K, V]) move(elem *list.Element) {
	item := elem.Value.(*lfuItem[K, V])
	freq := item.freq

	curr := elem
	for ; curr.Prev() != nil; curr = curr.Prev() {
		if freq != curr.Value.(*lfuItem[K, V]).freq {
			break
		}
	}

	if curr == elem {
		item.freq++
		return
	}

	if curr.Value.(*lfuItem[K, V]).freq == freq {
		l.evictionList.MoveToFront(elem)
		item.freq++
		return
	}

	l.evictionList.MoveBefore(elem, curr)
	item.freq++
}
