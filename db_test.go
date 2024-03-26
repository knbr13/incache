package inmemdb

import (
	"fmt"
	"sync"
	"testing"
)

func TestSet(t *testing.T) {
	db := New[string, string]()

	db.Set("key1", "value1")
	if db.m["key1"] != "value1" {
		t.Errorf("Set failed")
	}
}

func TestGet(t *testing.T) {
	db := New[string, string]()

	db.Set("key1", "value1")

	value, ok := db.Get("key1")
	if !ok || value != "value1" {
		t.Errorf("Get returned unexpected value for key1")
	}
}

func TestGetNonExistentKey(t *testing.T) {
	db := New[string, string]()

	_, ok := db.Get("nonexistent")
	if ok {
		t.Errorf("Get returned true for a non-existent key")
	}
}

func TestSetConcurrently(t *testing.T) {
	db := New[string, string]()
	var wg sync.WaitGroup
	numRoutines := 10_000

	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(i int) {
			defer wg.Done()
			db.Set(fmt.Sprintf("%d", i), fmt.Sprintf("%d", i))
		}(i)
	}
	wg.Wait()

	for i := 0; i < numRoutines; i++ {
		value, ok := db.Get(fmt.Sprintf("%d", i))
		if !ok || value != fmt.Sprintf("%d", i) {
			t.Errorf("Get returned unexpected value for key %d", i)
		}
	}
}

func TestDelete(t *testing.T) {
	db := New[string, string]()

	db.Set("key1", "value1")
	db.Delete("key1")

	_, ok := db.Get("key1")
	if ok {
		t.Errorf("Get returned true for a deleted key")
	}
}

func TestTransferTo(t *testing.T) {
	src := New[string, string]()
	dst := New[string, string]()

	src.Set("key1", "value1")
	src.TransferTo(dst)

	value, ok := dst.Get("key1")
	if !ok || value != "value1" {
		t.Errorf("TransferTo did not transfer data correctly")
	}

	_, ok = src.Get("key1")
	if ok {
		t.Errorf("TransferTo did not clear source database")
	}
}

func TestCopyTo(t *testing.T) {
	src := New[string, string]()
	dst := New[string, string]()

	src.Set("key1", "value1")
	src.CopyTo(dst)

	value, ok := dst.Get("key1")
	if !ok || value != "value1" {
		t.Errorf("CopyTo did not copy data correctly")
	}

	value, ok = src.Get("key1")
	if !ok || value != "value1" {
		t.Errorf("CopyTo modified the source database")
	}
}

func TestKeys(t *testing.T) {
	db := New[string, string]()

	db.Set("key1", "value1")
	db.Set("key2", "value2")

	keys := db.Keys()

	if len(keys) != 2 {
		t.Errorf("Unexpected number of keys returned")
	}

	expectedKeys := map[string]bool{"key1": true, "key2": true}
	for _, key := range keys {
		if !expectedKeys[key] {
			t.Errorf("Unexpected key %s returned", key)
		}
	}
}
