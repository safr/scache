// Package scache providers a cache functionality that stores key/value pairs.
package scache

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

// CacheItem stores the value and the expiry time of a cache entry.
type CacheItem struct {
	Value      string
	ExpiryTime time.Time
}

// entry is a helper struct that stores a cache item along with its key.
type entry struct {
	key   string
	value CacheItem
}

// Cache represents a thread-safe in-memory cache with TTL and LRU eviction policies.
type Cache struct {
	mu       sync.RWMutex
	items    map[string]*list.Element // Map of keys to list elements
	eviction *list.List               // Doubly-linked list for eviction
	capacity int                      // Maximum number of items in the cache
}

// New initializes and returns a new Cache with the given capacity.
func New(capacity int) *Cache {
	return &Cache{
		items:    make(map[string]*list.Element),
		eviction: list.New(),
		capacity: capacity,
	}
}

// Set adds or updates a cache entry with the specified key, value, and TTL.
func (c *Cache) Set(key, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove the old value if it exists
	if elem, found := c.items[key]; found {
		c.eviction.Remove(elem)
		delete(c.items, key)
	}

	// Evict the least recently used item if the cache is at capacity
	if c.eviction.Len() >= c.capacity {
		c.evictLRU()
	}

	item := CacheItem{
		Value:      value,
		ExpiryTime: time.Now().Add(ttl),
	}
	elem := c.eviction.PushFront(&entry{key, item})
	c.items[key] = elem

	return nil
}

// Get retrieves a cache entry by its key. It returns the value and a boolean indicating whether the key was found.
func (c *Cache) Get(key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	elem, found := c.items[key]
	if !found || time.Now().After(elem.Value.(*entry).value.ExpiryTime) {
		// If the item is not found or has expired, return false
		if found {
			c.eviction.Remove(elem)
			delete(c.items, key)
		}
		return "", errors.New("key not found")
	}
	// Move the accessed element to the front of the eviction list
	c.eviction.MoveToFront(elem)
	return elem.Value.(*entry).value.Value, nil
}

// Contains checks if cached key exists in the cache.
func (c *Cache) Contains(key string) bool {
	_, err := c.Get(key)
	return err == nil
}

// Flush removes all cached keys of the cache.
func (c *Cache) Flush() error {
	c.items = make(map[string]*list.Element)
	c.eviction = list.New()
	return nil
}

// evictLRU removes the least recently used item from the cache.
func (c *Cache) evictLRU() {
	elem := c.eviction.Back()
	if elem != nil {
		c.eviction.Remove(elem)
		kv := elem.Value.(*entry)
		delete(c.items, kv.key)
	}
}

// StartEvictionTicker starts a background goroutine that periodically evicts expired items.
func (c *Cache) StartEvictionTicker(d time.Duration) {
	ticker := time.NewTicker(d)
	go func() {
		for range ticker.C {
			c.evictExpiredItems()
		}
	}()
}

// evictExpiredItems removes all expired items from the cache.
func (c *Cache) evictExpiredItems() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for key, elem := range c.items {
		if now.After(elem.Value.(*entry).value.ExpiryTime) {
			c.eviction.Remove(elem)
			delete(c.items, key)
		}
	}
}
