package inmemdb

import (
	"sync"
	"time"
)

// Option is a functional option type for configuring the behavior of the in-memory database.
type Option[K comparable, V any] func(*DB[K, V])

// WithTimeInterval sets the time interval for the expiration goroutine to sleep before checking for expired keys again.
// It accepts a time.Duration value representing the interval duration.
// By default, the time interval is set to 10 seconds.
func WithTimeInterval[K comparable, V any](t time.Duration) Option[K, V] {
	return func(db *DB[K, V]) {
		db.timeInterval = t
	}
}

type DB[K comparable, V any] struct {
	m            map[K]valueWithTimeout[V]
	mu           sync.RWMutex
	stopCh       chan struct{} // Channel to signal timeout goroutine to stop
	timeInterval time.Duration // Time interval to sleep the goroutine that checks for expired keys
	expiryEnable bool          // Whether the database can contain keys that have expiry time or not
}

type valueWithTimeout[V any] struct {
	value    V
	expireAt *time.Time
}

// New creates a new in-memory database instance with optional configuration provided by the specified options.
// The database starts a background goroutine to periodically check for expired keys based on the configured time interval.
func New[K comparable, V any](opts ...Option[K, V]) *DB[K, V] {
	db := &DB[K, V]{
		m:            make(map[K]valueWithTimeout[V]),
		stopCh:       make(chan struct{}),
		timeInterval: time.Second * 10,
		expiryEnable: true,
	}
	for _, opt := range opts {
		opt(db)
	}
	if db.timeInterval > 0 {
		go db.expireKeys()
	} else {
		db.expiryEnable = false
	}
	return db
}

// Set adds or updates a key-value pair in the database without setting an expiration time.
// If the key already exists, its value will be overwritten with the new value.
// This function is safe for concurrent use.
func (d *DB[K, V]) Set(k K, v V) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.m[k] = valueWithTimeout[V]{
		value:    v,
		expireAt: nil,
	}
}

// SetWithTimeout adds or updates a key-value pair in the database with an expiration time.
// If the timeout duration is zero or negative, the key-value pair will not have an expiration time.
// This function is safe for concurrent use.
func (d *DB[K, V]) SetWithTimeout(k K, v V, timeout time.Duration) {
	if !d.expiryEnable {
		d.Set(k, v)
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	if timeout > 0 {
		now := time.Now().Add(timeout)
		d.m[k] = valueWithTimeout[V]{
			value:    v,
			expireAt: &now,
		}
	} else {
		d.m[k] = valueWithTimeout[V]{
			value:    v,
			expireAt: nil,
		}
	}
}

func (d *DB[K, V]) Get(k K) (V, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	v, ok := d.m[k]
	return v.value, ok
}

func (d *DB[K, V]) Delete(k K) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.m, k)
}

// TransferTo transfers all key-value pairs from the source DB to the provided destination DB.
//
// The source DB is locked during the entire operation, and the destination DB is locked for the duration of the function call.
// The function is safe to call concurrently with other operations on any of the source DB or Destination DB.
func (src *DB[K, V]) TransferTo(dst *DB[K, V]) {
	dst.mu.Lock()
	src.mu.Lock()
	defer dst.mu.Unlock()
	defer src.mu.Unlock()
	for k, v := range src.m {
		dst.m[k] = v
	}
	src.m = make(map[K]valueWithTimeout[V])
}

// CopyTo copies all key-value pairs from the source DB to the provided destination DB.
//
// The source DB is locked during the entire operation, and the destination DB is locked for the duration of the function call.
// The function is safe to call concurrently with other operations on any of the source DB or Destination DB.
func (src *DB[K, V]) CopyTo(dst *DB[K, V]) {
	dst.mu.Lock()
	src.mu.RLock()
	defer dst.mu.Unlock()
	defer src.mu.RUnlock()
	for k, v := range src.m {
		dst.m[k] = v
	}
}

// Keys returns a slice containing the keys of the map in random order.
func (d *DB[K, V]) Keys() []K {
	d.mu.RLock()
	defer d.mu.RUnlock()
	keys := make([]K, 0, len(d.m))
	for k := range d.m {
		keys = append(keys, k)
	}
	return keys
}

// expireKeys is a background goroutine that periodically checks for expired keys and removes them from the database.
// It runs until the Close method is called.
// This function is not intended to be called directly by users.
func (d *DB[K, V]) expireKeys() {
	ticker := time.NewTicker(d.timeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			d.mu.Lock()
			for k, v := range d.m {
				if v.expireAt != nil && v.expireAt.Before(time.Now()) {
					delete(d.m, k)
				}
			}
			d.mu.Unlock()
		case <-d.stopCh:
			return
		}
	}
}

// Close signals the expiration goroutine to stop.
// It should be called when the database is no longer needed.
func (d *DB[K, V]) Close() {
	if d.expiryEnable {
		d.stopCh <- struct{}{} // Signal the expiration goroutine to stop
		close(d.stopCh)
	}
	d.m = nil
}
