package inmemdb

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

func (d *DB[K, V]) Delete(k K) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.m, k)
}

func (d *DB[K, V]) MoveTo(db *DB[K, V]) {
	db.mu.Lock()
	defer db.mu.Unlock()
	for k, v := range d.m {
		db.m[k] = v
	}
	d.m = make(map[K]V)
}

func (d *DB[K, V]) CopyTo(db *DB[K, V]) {
	db.mu.Lock()
	defer db.mu.Unlock()
	for k, v := range d.m {
		db.m[k] = v
	}
}
