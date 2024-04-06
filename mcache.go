package inmemdb

import (
	"time"
)

type MCache[K comparable, V any] struct {
	baseCache
	m            map[K]valueWithTimeout[V]
	stopCh       chan struct{} // Channel to signal timeout goroutine to stop
	timeInterval time.Duration // Time interval to sleep the goroutine that checks for expired keys
}

type valueWithTimeout[V any] struct {
	value    V
	expireAt *time.Time
}

// New creates a new in-memory database instance with optional configuration provided by the specified options.
// The database starts a background goroutine to periodically check for expired keys based on the configured time interval.
func newManual[K comparable, V any](cacheBuilder *CacheBuilder[K, V]) *MCache[K, V] {
	db := &MCache[K, V]{
		m:            make(map[K]valueWithTimeout[V]),
		stopCh:       make(chan struct{}),
		timeInterval: cacheBuilder.tmIvl,
		baseCache: baseCache{
			size: cacheBuilder.size,
		},
	}
	if db.timeInterval > 0 {
		go db.expireKeys()
	}
	return db
}

// Set adds or updates a key-value pair in the database without setting an expiration time.
// If the key already exists, its value will be overwritten with the new value.
// This function is safe for concurrent use.
func (c *MCache[K, V]) Set(k K, v V) {
	if c.size == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.m) == int(c.size) {
		c.evict(1)
	}

	c.m[k] = valueWithTimeout[V]{
		value:    v,
		expireAt: nil,
	}
}

func (c *MCache[K, V]) setValueWithTimeout(k K, v valueWithTimeout[V]) {
	if c.size == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.m) == int(c.size) {
		c.evict(1)
	}

	c.m[k] = v
}

// NotFoundSet adds a key-value pair to the database if the key does not already exist and returns true. Otherwise, it does nothing and returns false.
func (c *MCache[K, V]) NotFoundSet(k K, v V) bool {
	if c.size == 0 {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.m[k]
	if !ok {
		if len(c.m) == int(c.size) {
			c.evict(1)
		}

		c.m[k] = valueWithTimeout[V]{
			value:    v,
			expireAt: nil,
		}
	}
	return !ok
}

// SetWithTimeout adds or updates a key-value pair in the database with an expiration time.
// If the timeout duration is zero or negative, the key-value pair will not have an expiration time.
// This function is safe for concurrent use.
func (c *MCache[K, V]) SetWithTimeout(k K, v V, timeout time.Duration) {
	if c.size == 0 {
		return
	}
	if timeout > 0 {
		c.mu.Lock()
		defer c.mu.Unlock()

		if len(c.m) == int(c.size) {
			c.evict(1)
		}

		now := time.Now().Add(timeout)
		c.m[k] = valueWithTimeout[V]{
			value:    v,
			expireAt: &now,
		}
	} else {
		c.Set(k, v)
	}
}

// NotFoundSetWithTimeout adds a key-value pair to the database with an expiration time if the key does not already exist and returns true. Otherwise, it does nothing and returns false.
// If the timeout is zero or negative, the key-value pair will not have an expiration time.
// If expiry is disabled, it behaves like NotFoundSet.
func (c *MCache[K, V]) NotFoundSetWithTimeout(k K, v V, timeout time.Duration) bool {
	if c.size == 0 {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	var ok bool
	if timeout > 0 {
		now := time.Now().Add(timeout)
		_, ok = c.m[k]
		if !ok {
			if len(c.m) == int(c.size) {
				c.evict(1)
			}

			c.m[k] = valueWithTimeout[V]{
				value:    v,
				expireAt: &now,
			}
		}
	} else {
		_, ok = c.m[k]
		if !ok {
			if len(c.m) == int(c.size) {
				c.evict(1)
			}

			c.m[k] = valueWithTimeout[V]{
				value:    v,
				expireAt: nil,
			}
		}
	}
	return !ok
}

func (c *MCache[K, V]) Get(k K) (v V, b bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.m[k]
	if !ok {
		return
	}
	if val.expireAt != nil && val.expireAt.Before(time.Now()) {
		delete(c.m, k)
		return
	}
	return val.value, ok
}

func (c *MCache[K, V]) Delete(k K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.m, k)
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
func (c *MCache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, len(c.m))
	var i uint

	for k := range c.m {
		keys[i] = k
		i++
	}
	return keys
}

// expireKeys is a background goroutine that periodically checks for expired keys and removes them from the database.
// It runs until the Close method is callec.
// This function is not intended to be called directly by users.
func (c *MCache[K, V]) expireKeys() {
	ticker := time.NewTicker(c.timeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for k, v := range c.m {
				if v.expireAt != nil && v.expireAt.Before(time.Now()) {
					delete(c.m, k)
				}
			}
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}

// Purge clears the cache completely.
func (c *MCache[K, V]) Purge() {
	if c.timeInterval > 0 {
		c.stopCh <- struct{}{} // Signal the expiration goroutine to stop
		close(c.stopCh)
	}
	c.m = nil
}

// Count returns the number of key-value pairs in the database.
func (c *MCache[K, V]) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.m)
}

func (c *MCache[K, V]) evict(i int) {
	var counter int
	for k, v := range c.m {
		if counter == i {
			break
		}
		if v.expireAt != nil && !v.expireAt.After(time.Now()) {
			delete(c.m, k)
			counter++
		}
	}
	if i > len(c.m) {
		i = len(c.m)
	}
	for counter < i {
		for k := range c.m {
			delete(c.m, k)
			counter++
		}
	}
}
