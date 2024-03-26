package main

import "sync"

type DB[K comparable, V any] struct {
	m  map[K]V
	mu sync.RWMutex
}

func New[K comparable, V any]() *DB[K, V] {
	return &DB[K, V]{
		m:  make(map[K]V),
		mu: sync.RWMutex{},
	}
}

func (d *DB[K, V]) Set(k K, v V) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.m[k] = v
}

func (d *DB[K, V]) Get(k K) (V, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	v, ok := d.m[k]
	return v, ok
}
