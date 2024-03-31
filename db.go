package inmemdb

import (
	"time"
)

type MCache[K comparable, V any] struct {
	baseCache[K, V]
	stopCh       chan struct{} // Channel to signal timeout goroutine to stop
	timeInterval time.Duration // Time interval to sleep the goroutine that checks for expired keys
}

type valueWithTimeout[V any] struct {
	value    V
	expireAt *time.Time
}

// New creates a new in-memory database instance with optional configuration provided by the specified options.
// The database starts a background goroutine to periodically check for expired keys based on the configured time interval.
func newManual[K comparable, V any](timeInterval time.Duration) *MCache[K, V] {
	db := &MCache[K, V]{
		baseCache: baseCache[K, V]{
			m: make(map[K]valueWithTimeout[V]),
		},
		stopCh:       make(chan struct{}),
		timeInterval: timeInterval,
	}
	if db.timeInterval > 0 {
		go db.expireKeys()
	}
	return db
}

// Set adds or updates a key-value pair in the database without setting an expiration time.
// If the key already exists, its value will be overwritten with the new value.
// This function is safe for concurrent use.
func (d *MCache[K, V]) Set(k K, v V) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.m[k] = valueWithTimeout[V]{
		value:    v,
		expireAt: nil,
	}
}

func (d *MCache[K, V]) setValueWithTimeout(k K, v valueWithTimeout[V]) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.m[k] = v
}

// NotFoundSet adds a key-value pair to the database if the key does not already exist and returns true. Otherwise, it does nothing and returns false.
func (d *MCache[K, V]) NotFoundSet(k K, v V) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.m[k]
	if !ok {
		d.m[k] = valueWithTimeout[V]{
			value:    v,
			expireAt: nil,
		}
	}
	return !ok
}

// SetWithTimeout adds or updates a key-value pair in the database with an expiration time.
// If the timeout duration is zero or negative, the key-value pair will not have an expiration time.
// This function is safe for concurrent use.
func (d *MCache[K, V]) SetWithTimeout(k K, v V, timeout time.Duration) {
	if timeout > 0 {
		d.mu.Lock()
		defer d.mu.Unlock()

		now := time.Now().Add(timeout)
		d.m[k] = valueWithTimeout[V]{
			value:    v,
			expireAt: &now,
		}
	} else {
		d.Set(k, v)
	}
}

// NotFoundSetWithTimeout adds a key-value pair to the database with an expiration time if the key does not already exist and returns true. Otherwise, it does nothing and returns false.
// If the timeout is zero or negative, the key-value pair will not have an expiration time.
// If expiry is disabled, it behaves like NotFoundSet.
func (d *MCache[K, V]) NotFoundSetWithTimeout(k K, v V, timeout time.Duration) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	var ok bool
	if timeout > 0 {
		now := time.Now().Add(timeout)
		_, ok = d.m[k]
		if !ok {
			d.m[k] = valueWithTimeout[V]{
				value:    v,
				expireAt: &now,
			}
		}
	} else {
		_, ok = d.m[k]
		if !ok {
			d.m[k] = valueWithTimeout[V]{
				value:    v,
				expireAt: nil,
			}
		}
	}
	return !ok
}

func (d *MCache[K, V]) Get(k K) (v V, b bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	val, ok := d.m[k]
	if !ok {
		return
	}
	if val.expireAt != nil && val.expireAt.Before(time.Now()) {
		delete(d.m, k)
		return
	}
	return val.value, ok
}

func (d *MCache[K, V]) Delete(k K) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.m, k)
}

// TransferTo transfers all key-value pairs from the source DB to the provided destination DB.
//
// The source DB is locked during the entire operation, and the destination DB is locked for the duration of the function call.
// The function is safe to call concurrently with other operations on any of the source DB or Destination DB.
func (src *MCache[K, V]) TransferTo(dst Cache[K, V]) {
	src.mu.Lock()
	defer src.mu.Unlock()
	for k, v := range src.m {
		dst.setValueWithTimeout(k, v)
	}
	src.m = make(map[K]valueWithTimeout[V])
}

// CopyTo copies all key-value pairs from the source DB to the provided destination DB.
//
// The source DB is locked during the entire operation, and the destination DB is locked for the duration of the function call.
// The function is safe to call concurrently with other operations on any of the source DB or Destination DB.
func (src *MCache[K, V]) CopyTo(dst Cache[K, V]) {
	src.mu.RLock()
	defer src.mu.RUnlock()
	for k, v := range src.m {
		dst.setValueWithTimeout(k, v)
	}
}

// Keys returns a slice containing the keys of the map in random order.
func (d *MCache[K, V]) Keys() []K {
	d.mu.RLock()
	defer d.mu.RUnlock()
	keys := make([]K, 0, len(d.m))
	for k := range d.m {
		// if v.expireAt != nil && v.expireAt.Before(time.Now()) {
		// 	delete(d.m, k)
		// 	continue
		// }
		keys = append(keys, k)
	}
	return keys
}

// expireKeys is a background goroutine that periodically checks for expired keys and removes them from the database.
// It runs until the Close method is called.
// This function is not intended to be called directly by users.
func (d *MCache[K, V]) expireKeys() {
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
func (d *MCache[K, V]) Close() {
	if d.timeInterval > 0 {
		d.stopCh <- struct{}{} // Signal the expiration goroutine to stop
		close(d.stopCh)
	}
	d.m = nil
}

// Count returns the number of key-value pairs in the database.
func (d *MCache[K, V]) Count() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.m)
}
