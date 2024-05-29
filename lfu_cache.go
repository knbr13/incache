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

func (l *LFUCache[K, V]) Set(key K, value V) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.set(key, value, 0)
}

func (l *LFUCache[K, V]) set(key K, value V, exp time.Duration) {
	item, ok := l.m[key]
	var tm *time.Time
	if exp > 0 {
		t := time.Now().Add(exp)
		tm = &t
	}
	if ok {
		lfuItem := item.Value.(*lfuItem[K, V])

		lfuItem.value = value
		lfuItem.expireAt = tm
		lfuItem.freq++

		l.move(item)
	} else {
		if len(l.m) == int(l.size) {
			l.evict(1)
		}

		lfuItem := lfuItem[K, V]{
			key:   key,
			value: value,
			freq:  1,
		}

		l.m[key] = l.evictionList.PushBack(&lfuItem)
	}
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
