package incache

import (
	"container/list"
	"sync"
	"time"
)

// Least Frequently Used Cache
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

// Set adds the key-value pair to the cache.
func (l *LFUCache[K, V]) Set(key K, value V) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.set(key, value, 0)
}

// SetWithTimeout adds the key-value pair to the cache with a specified expiration time.
func (l *LFUCache[K, V]) SetWithTimeout(key K, value V, exp time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.set(key, value, exp)
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
			key:      key,
			value:    value,
			expireAt: tm,
			freq:     1,
		}

		l.m[key] = l.evictionList.PushBack(&lfuItem)
		l.move(l.m[key])
	}
}

// Get retrieves the value associated with the given key from the cache.
// If the key is not found or has expired, it returns (zero value of V, false).
// Otherwise, it returns (value, true).
func (l *LFUCache[K, V]) Get(key K) (v V, b bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.m[key]
	if !ok {
		return
	}

	lfuItem := item.Value.(*lfuItem[K, V])
	if lfuItem.expireAt != nil && lfuItem.expireAt.Before(time.Now()) {
		l.delete(key, item)
		return
	}

	lfuItem.freq++
	l.move(item)

	return lfuItem.value, true
}

// NotFoundSet adds the key-value pair to the cache only if the key does not exist.
// It returns true if the key was added to the cache, otherwise false.
func (l *LFUCache[K, V]) NotFoundSet(k K, v V) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.m[k]
	if ok {
		return false
	}

	l.set(k, v, 0)
	return true
}

// NotFoundSetWithTimeout adds the key-value pair to the cache only if the key does not exist.
// It sets an expiration time for the key-value pair.
// It returns true if the key was added to the cache, otherwise false.
func (l *LFUCache[K, V]) NotFoundSetWithTimeout(k K, v V, t time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	_, ok := l.m[k]
	if ok {
		return false
	}

	l.set(k, v, t)
	return true
}

// GetAll retrieves all key-value pairs from the cache.
// It returns a map containing all the key-value pairs that are not expired.
func (l *LFUCache[K, V]) GetAll() map[K]V {
	l.mu.RLock()
	defer l.mu.RUnlock()

	m := make(map[K]V)
	for k, v := range l.m {
		if v.Value.(*lfuItem[K, V]).expireAt == nil || !v.Value.(*lfuItem[K, V]).expireAt.Before(time.Now()) {
			m[k] = v.Value.(*lfuItem[K, V]).value
		}
	}

	return m
}

// TransferTo transfers all non-expired key-value pairs from the source cache to the destination cache.
func (src *LFUCache[K, V]) TransferTo(dst *LFUCache[K, V]) {
	src.mu.Lock()
	defer src.mu.Unlock()

	for k, v := range src.m {
		if v.Value.(*lfuItem[K, V]).expireAt == nil || !v.Value.(*lfuItem[K, V]).expireAt.Before(time.Now()) {
			src.delete(k, v)
			dst.Set(k, v.Value.(*lfuItem[K, V]).value)
		}
	}
}

// CopyTo copies all non-expired key-value pairs from the source cache to the destination cache.
func (src *LFUCache[K, V]) CopyTo(dst *LFUCache[K, V]) {
	src.mu.RLock()
	defer src.mu.RUnlock()

	for k, v := range src.m {
		if v.Value.(*lfuItem[K, V]).expireAt == nil || !v.Value.(*lfuItem[K, V]).expireAt.Before(time.Now()) {
			dst.Set(k, v.Value.(*lfuItem[K, V]).value)
		}
	}
}

// Keys returns a slice of all keys currently stored in the cache.
// The returned slice does not include expired keys.
// The order of keys in the slice is not guaranteed.
func (l *LFUCache[K, V]) Keys() []K {
	l.mu.RLock()
	defer l.mu.RUnlock()

	keys := make([]K, 0, l.Count())

	for k, v := range l.m {
		if v.Value.(*lfuItem[K, V]).expireAt == nil || !v.Value.(*lfuItem[K, V]).expireAt.Before(time.Now()) {
			keys = append(keys, k)
		}
	}

	return keys
}

// Purge removes all key-value pairs from the cache.
func (l *LFUCache[K, V]) Purge() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.m = make(map[K]*list.Element)
	l.evictionList.Init()
}

// Count returns the number of non-expired key-value pairs currently stored in the cache.
func (l *LFUCache[K, V]) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var count int
	for _, v := range l.m {
		if v.Value.(*lfuItem[K, V]).expireAt == nil || !v.Value.(*lfuItem[K, V]).expireAt.Before(time.Now()) {
			count++
		}
	}

	return count
}

// Len returns the number of elements in the cache.
func (l *LFUCache[K, V]) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return len(l.m)
}

// Delete removes the key-value pair associated with the given key from the cache.
func (l *LFUCache[K, V]) Delete(k K) {
	l.mu.Lock()
	defer l.mu.Unlock()

	item, ok := l.m[k]
	if !ok {
		return
	}

	l.delete(k, item)
}

func (l *LFUCache[K, V]) delete(key K, elem *list.Element) {
	delete(l.m, key)
	l.evictionList.Remove(elem)
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
		if freq > curr.Value.(*lfuItem[K, V]).freq {
			break
		}
	}

	if curr == elem {
		return
	}

	if curr.Value.(*lfuItem[K, V]).freq == freq {
		l.evictionList.MoveToFront(elem)
		return
	}

	l.evictionList.MoveAfter(elem, curr)
}
