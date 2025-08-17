package cache

import (
	"container/list"
	"sync"
	"time"
)

// Cache represents a generic cache interface
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
	Stats() Stats
}

// Stats represents cache statistics
type Stats struct {
	Hits        int64
	Misses      int64
	Evictions   int64
	Size        int
	MaxSize     int
	HitRate     float64
	MemoryUsage int64
}

// Item represents a cached item with TTL
type Item struct {
	Key        string
	Value      interface{}
	Expiration time.Time
	AccessTime time.Time
	element    *list.Element
}

// IsExpired checks if the item has expired
func (i *Item) IsExpired() bool {
	return time.Now().After(i.Expiration)
}

// LRUCache implements an LRU cache with TTL support
type LRUCache struct {
	maxSize    int
	items      map[string]*Item
	lruList    *list.List
	mutex      sync.RWMutex
	hits       int64
	misses     int64
	evictions  int64
}

// NewLRUCache creates a new LRU cache with the specified maximum size
func NewLRUCache(maxSize int) *LRUCache {
	return &LRUCache{
		maxSize: maxSize,
		items:   make(map[string]*Item),
		lruList: list.New(),
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	item, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}

	// Check if item has expired
	if item.IsExpired() {
		c.removeItem(item)
		c.misses++
		return nil, false
	}

	// Move to front (most recently used)
	c.lruList.MoveToFront(item.element)
	item.AccessTime = time.Now()
	c.hits++

	return item.Value, true
}

// Set stores a value in the cache with the specified TTL
func (c *LRUCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expiration := now.Add(ttl)

	// If item already exists, update it
	if existingItem, exists := c.items[key]; exists {
		existingItem.Value = value
		existingItem.Expiration = expiration
		existingItem.AccessTime = now
		c.lruList.MoveToFront(existingItem.element)
		return
	}

	// Create new item
	item := &Item{
		Key:        key,
		Value:      value,
		Expiration: expiration,
		AccessTime: now,
	}

	// Add to front of LRU list
	item.element = c.lruList.PushFront(item)
	c.items[key] = item

	// Check if we need to evict items
	if len(c.items) > c.maxSize {
		c.evictLRU()
	}
}

// Delete removes an item from the cache
func (c *LRUCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, exists := c.items[key]; exists {
		c.removeItem(item)
	}
}

// Clear removes all items from the cache
func (c *LRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*Item)
	c.lruList.Init()
}

// Stats returns cache statistics
func (c *LRUCache) Stats() Stats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}

	// Estimate memory usage (rough calculation)
	var memoryUsage int64
	for _, item := range c.items {
		memoryUsage += c.estimateItemSize(item)
	}

	return Stats{
		Hits:        c.hits,
		Misses:      c.misses,
		Evictions:   c.evictions,
		Size:        len(c.items),
		MaxSize:     c.maxSize,
		HitRate:     hitRate,
		MemoryUsage: memoryUsage,
	}
}

// CleanupExpired removes all expired items from the cache
func (c *LRUCache) CleanupExpired() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	var toRemove []*Item

	// Collect expired items
	for _, item := range c.items {
		if now.After(item.Expiration) {
			toRemove = append(toRemove, item)
		}
	}

	// Remove expired items
	for _, item := range toRemove {
		c.removeItem(item)
	}
}

// removeItem removes an item from both the map and LRU list
func (c *LRUCache) removeItem(item *Item) {
	delete(c.items, item.Key)
	c.lruList.Remove(item.element)
}

// evictLRU removes the least recently used item
func (c *LRUCache) evictLRU() {
	if c.lruList.Len() > 0 {
		oldest := c.lruList.Back()
		if oldest != nil {
			item := oldest.Value.(*Item)
			c.removeItem(item)
			c.evictions++
		}
	}
}

// estimateItemSize provides a rough estimate of item memory usage
func (c *LRUCache) estimateItemSize(item *Item) int64 {
	// Basic size estimation - this is simplified
	size := int64(len(item.Key) + 64) // Basic overhead
	
	// Add estimated value size based on type
	switch v := item.Value.(type) {
	case string:
		size += int64(len(v))
	case []byte:
		size += int64(len(v))
	default:
		size += 256 // Default estimate for complex objects
	}
	
	return size
}

// StartCleanupRoutine starts a background goroutine to clean up expired items
func (c *LRUCache) StartCleanupRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for range ticker.C {
			c.CleanupExpired()
		}
	}()
}