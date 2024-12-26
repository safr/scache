package scache

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

const (
	testKey   = "testKey"
	testValue = "testValue"
)

func TestCacheInitialization(t *testing.T) {
	cache := New(10)
	if cache == nil {
		t.Errorf("New() = %v, want non-nil", cache)
	}
}

func TestCacheSetAndGet(t *testing.T) {
	cache := New(10)

	if err := cache.Set(testKey, testValue, 1*time.Hour); err != nil {
		t.Errorf("Set() = %v, want %v", err, nil)
	}

	value, err := cache.Get(testKey)
	if err != nil || value != testValue {
		t.Errorf("Get() = %v, %v, want %v, %v", value, err, testValue, "key not found")
	}
}

func TestCacheContainsKey(t *testing.T) {
	cache := New(10)

	if err := cache.Set(testKey, testValue, 1*time.Hour); err != nil {
		t.Errorf("Set() = %v, want %v", err, nil)
	}

	if !cache.Contains(testKey) {
		t.Errorf("contains failed: the key %s should be exist", testKey)
	}
}

func TestCacheFlush(t *testing.T) {
	cache := New(10)

	if err := cache.Set(testKey, testValue, 1*time.Hour); err != nil {
		t.Errorf("Set() = %v, want %v", err, nil)
	}

	if err := cache.Flush(); err != nil {
		t.Errorf("flush failed: expected nil, got %v", err)
	}

	if cache.Contains(testKey) {
		t.Errorf("contains failed: the key %s should not be exist", testKey)
	}
}

func TestCacheGetNonExistentKey(t *testing.T) {
	cache := New(10)

	_, err := cache.Get("nonExistentKey")
	if err == nil {
		t.Errorf("Get() = %v, want %v", err, "key not found")
	}
}

func TestCacheSetOverwritesValue(t *testing.T) {
	cache := New(10)

	if err := cache.Set(testKey, testValue, 1*time.Hour); err != nil {
		t.Errorf("Set() = %v, want %v", err, nil)
	}

	if err := cache.Set(testKey, "value2", 1*time.Hour); err != nil {
		t.Errorf("Set() = %v, want %v", err, nil)
	}

	value, _ := cache.Get(testKey)
	if value != "value2" {
		t.Errorf("Get() = %v, want %v", value, "value2")
	}
}

func TestCacheSetUpdatesExpiryTime(t *testing.T) {
	cache := New(2)
	if err := cache.Set(testKey, testValue, 1*time.Second); err != nil {
		t.Errorf("Set() = %v, want %v", err, nil)
	}

	time.Sleep(2 * time.Second)
	_, err := cache.Get(testKey)
	if err == nil {
		t.Errorf("Get() = %v, want %v", err, "key not found")
	}
	if err := cache.Set(testKey, testValue, 1*time.Hour); err != nil {
		t.Errorf("Set() = %v, want %v", err, nil)
	}
	_, err = cache.Get(testKey)
	if err != nil {
		t.Errorf("Get() = %v, want %v", err, nil)
	}
}

func TestCacheEvictsLRU(t *testing.T) {
	cache := New(2)
	if err := cache.Set(testKey, testValue, 1*time.Hour); err != nil {
		t.Errorf("Set() = %v, want %v", err, nil)
	}
	if err := cache.Set("key2", "value2", 1*time.Hour); err != nil {
		t.Errorf("Set() = %v, want %v", err, nil)
	}
	if err := cache.Set("key3", "value3", 1*time.Hour); err != nil {
		t.Errorf("Set() = %v, want %v", err, nil)
	}

	_, err := cache.Get(testKey)
	if err == nil {
		t.Errorf("Get() = %v, want %v", err, "key not found")
	}
}

func TestCacheEvictsExpiredItems(t *testing.T) {
	cache := New(2)
	if err := cache.Set(testKey, testValue, 1*time.Second); err != nil {
		t.Errorf("Set() = %v, want %v", err, nil)
	}
	time.Sleep(2 * time.Second)
	cache.evictExpiredItems()
	_, err := cache.Get(testKey)
	if err == nil {
		t.Errorf("Get() = %v, want %v", err, "key not found")
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := New(10) // Set a small capacity to induce eviction
	var wg sync.WaitGroup

	// Number of concurrent goroutines
	numGoroutines := 10

	// Number of operations per goroutine
	opsPerGoroutine := 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < opsPerGoroutine; j++ {
				key := strconv.Itoa(j) // Use simple keys for testing
				value := "value" + key

				// Set with a short TTL to test eviction
				if err := cache.Set(key, value, 1*time.Millisecond); err != nil {
					t.Errorf("Set() = %v, want %v", err, nil)
				}
				time.Sleep(1 * time.Millisecond) // Add slight delay for TTL to expire

				// Get should either return the value or "", false (if expired/evicted)
				val, err := cache.Get(key)
				if err == nil && val != value {
					t.Errorf("Unexpected value for key %s: got %s, want %s", key, val, value)
				}
			}
		}()
	}

	wg.Wait()
}
