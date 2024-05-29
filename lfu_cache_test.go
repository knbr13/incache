package incache

import (
	"testing"
	"time"
)

func TestLFUCache_SetGet(t *testing.T) {
	cache := NewLFU[int, string](10)

	cache.Set(1, "one")
	cache.Set(2, "two")

	if value, ok := cache.Get(1); !ok || value != "one" {
		t.Errorf("Expected to get 'one', got '%v'", value)
	}

	if value, ok := cache.Get(2); !ok || value != "two" {
		t.Errorf("Expected to get 'two', got '%v'", value)
	}
}

func TestLFUCache_Eviction(t *testing.T) {
	cache := NewLFU[int, string](2)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")

	if _, ok := cache.Get(1); ok {
		t.Logf("all: %v\n", cache.GetAll())
		t.Errorf("Expected 1 to be evicted")
	}

	if value, ok := cache.Get(2); !ok || value != "two" {
		t.Errorf("Expected to get 'two', got '%v'", value)
	}

	if value, ok := cache.Get(3); !ok || value != "three" {
		t.Errorf("Expected to get 'three', got '%v'", value)
	}
}

func TestLFUCache_SetWithTimeout(t *testing.T) {
	cache := NewLFU[int, string](10)

	cache.SetWithTimeout(1, "one", 2*time.Millisecond)
	time.Sleep(1 * time.Millisecond)

	if value, ok := cache.Get(1); !ok || value != "one" {
		t.Errorf("Expected to get 'one', got '%v'", value)
	}

	time.Sleep(2 * time.Millisecond)

	if v, ok := cache.Get(1); ok {
		t.Logf("v: %v | ok: %v\n", v, ok)
		t.Errorf("Expected 1 to be expired")
	}
}

func TestLFUCache_NotFoundSet(t *testing.T) {
	cache := NewLFU[int, string](10)

	if !cache.NotFoundSet(1, "one") {
		t.Errorf("Expected to set key 1")
	}

	if cache.NotFoundSet(1, "one") {
		t.Errorf("Expected not to set key 1 again")
	}

	if value, ok := cache.Get(1); !ok || value != "one" {
		t.Errorf("Expected to get 'one', got '%v'", value)
	}
}

func TestLFUCache_TransferTo(t *testing.T) {
	srcCache := NewLFU[int, string](10)
	dstCache := NewLFU[int, string](10)

	srcCache.Set(1, "one")
	srcCache.Set(2, "two")
	srcCache.TransferTo(dstCache)

	if _, ok := srcCache.Get(1); ok {
		t.Errorf("Expected 1 to be transferred")
	}

	if value, ok := dstCache.Get(1); !ok || value != "one" {
		t.Errorf("Expected to get 'one' from dstCache, got '%v'", value)
	}

	if value, ok := dstCache.Get(2); !ok || value != "two" {
		t.Errorf("Expected to get 'two' from dstCache, got '%v'", value)
	}
}

func TestLFUCache_CopyTo(t *testing.T) {
	srcCache := NewLFU[int, string](10)
	dstCache := NewLFU[int, string](10)

	srcCache.Set(1, "one")
	srcCache.Set(2, "two")
	srcCache.CopyTo(dstCache)

	if value, ok := srcCache.Get(1); !ok || value != "one" {
		t.Errorf("Expected to get 'one' from srcCache, got '%v'", value)
	}

	if value, ok := dstCache.Get(1); !ok || value != "one" {
		t.Errorf("Expected to get 'one' from dstCache, got '%v'", value)
	}

	if value, ok := dstCache.Get(2); !ok || value != "two" {
		t.Errorf("Expected to get 'two' from dstCache, got '%v'", value)
	}
}

func TestLFUCache_Keys(t *testing.T) {
	cache := NewLFU[int, string](10)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Set(3, "three")

	keys := cache.Keys()

	expectedKeys := map[int]bool{
		1: true,
		2: true,
		3: true,
	}

	for _, key := range keys {
		if !expectedKeys[key] {
			t.Errorf("Unexpected key %v", key)
		}
	}
}

func TestLFUCache_Purge(t *testing.T) {
	cache := NewLFU[int, string](10)

	cache.Set(1, "one")
	cache.Set(2, "two")
	cache.Purge()

	if value, ok := cache.Get(1); ok {
		t.Errorf("Expected cache to be purged, got '%v'", value)
	}

	if value, ok := cache.Get(2); ok {
		t.Errorf("Expected cache to be purged, got '%v'", value)
	}
}

func TestLFUCache_Delete(t *testing.T) {
	cache := NewLFU[int, string](10)

	cache.Set(1, "one")
	cache.Set(2, "two")

	cache.Delete(1)

	if value, ok := cache.Get(1); ok {
		t.Errorf("Expected key 1 to be deleted, got '%v'", value)
	}

	if value, ok := cache.Get(2); !ok || value != "two" {
		t.Errorf("Expected to get 'two', got '%v'", value)
	}
}
