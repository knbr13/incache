package incache

import (
	"testing"
	"time"
)

func TestSet_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.Set("key1", "value1")
	if c.m["key1"].Value.(*lruItem[string, string]).value != "value1" {
		t.Errorf("Set failed")
	}
}

func TestGet_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.Set("key1", "value1")
	if v, ok := c.Get("key1"); !ok || v != "value1" {
		t.Errorf("Get failed")
	}
}

func TestGetAll_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")
	c.Set("key4", "value4")
	c.Set("key5", "value5")
	c.Set("key6", "value6")
	c.Set("key7", "value7")
	c.Set("key8", "value8")
	c.Set("key9", "value9")
	c.Set("key10", "value10")
	c.Set("key11", "value11")
	c.Set("key12", "value12")

	if l := len(c.GetAll()); l != 10 {
		t.Errorf("GetAll failed: expected %d, got %d\n", 10, l)
	}
}

func TestSetWithTimeout_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.SetWithTimeout("key1", "value1", time.Millisecond)

	time.Sleep(time.Millisecond)

	if c.m["key1"].Value.(*lruItem[string, string]).value != "value1" {
		t.Errorf("SetWithTimeout failed")
	}
}

func TestNotFoundSet_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	if !c.NotFoundSet("key1", "value1") {
		t.Errorf("NotFoundSet failed")
	}

	if c.NotFoundSet("key1", "value2") {
		t.Errorf("NotFoundSet failed")
	}
}

func TestNotFoundSetWithTimeout_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	if !c.NotFoundSetWithTimeout("key1", "value1", 0) {
		t.Errorf("NotFoundSetWithTimeout failed")
	}

	if c.NotFoundSetWithTimeout("key1", "value2", 0) {
		t.Errorf("NotFoundSetWithTimeout failed")
	}
}

func TestDelete_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.Set("key1", "value1")

	c.Delete("key1")

	if _, ok := c.Get("key1"); ok {
		t.Errorf("Delete failed")
	}
}

func TestTransferTo_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")
	c.Set("key4", "value4")
	c.Set("key5", "value5")

	c2 := NewLRU[string, string](10)

	c.TransferTo(c2)

	if _, ok := c2.Get("key1"); !ok {
		t.Errorf("TransferTo failed")
	}

	if c.Len() != c.Count() || c.Len() != 0 || c2.Len() != c2.Count() || c2.Len() != 5 {
		t.Errorf("TransferTo failed")
	}
}

func TestCopyTo_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")
	c.Set("key4", "value4")
	c.Set("key5", "value5")

	c2 := NewLRU[string, string](10)

	c.CopyTo(c2)

	if _, ok := c2.Get("key1"); !ok {
		t.Errorf("CopyTo failed")
	}

	if c.Len() != c.Count() || c.Len() != 5 || c2.Len() != c2.Count() || c2.Len() != 5 {
		t.Errorf("TransferTo failed")
	}
}

func TestKeys_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")
	c.Set("key4", "value4")
	c.Set("key5", "value5")
	c.SetWithTimeout("key6", "value6", 1)
	c.SetWithTimeout("key7", "value7", 1)
	c.SetWithTimeout("key8", "value8", 1)
	c.SetWithTimeout("key9", "value9", 1)
	c.Set("key10", "value10")

	keys := c.Keys()

	if len(keys) != 6 {
		t.Errorf("Keys failed")
	}
}

func TestPurge_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")

	c.Purge()

	if _, ok := c.Get("key1"); ok {
		t.Errorf("Purge failed")
	}

	if _, ok := c.Get("key2"); ok {
		t.Errorf("Purge failed")
	}

	if _, ok := c.Get("key3"); ok {
		t.Errorf("Purge failed")
	}
}

func TestCount_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")
	c.Set("key4", "value4")
	c.Set("key5", "value5")

	if c.Count() != 5 {
		t.Errorf("Count failed")
	}

	c.SetWithTimeout("key6", "value6", time.Microsecond)
	time.Sleep(time.Microsecond)

	if c.Count() != 5 {
		t.Errorf("Count failed")
	}
}

func TestLen_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")
	c.Set("key4", "value4")
	c.Set("key5", "value5")

	if c.Len() != 5 {
		t.Errorf("Len failed")
	}

	c.SetWithTimeout("key6", "value6", time.Microsecond)
	time.Sleep(time.Microsecond)

	if c.Len() != 6 {
		t.Errorf("Len failed")
	}
}

func TestEvict_LRU(t *testing.T) {
	c := NewLRU[string, string](10)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Set("key3", "value3")
	c.Set("key4", "value4")
	c.Set("key5", "value5")

	c.evict(1)

	if c.Len() != c.Count() || c.Len() != 4 {
		t.Errorf("Evict failed")
	}
}
