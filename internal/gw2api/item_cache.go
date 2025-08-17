package gw2api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// ItemCache provides in-memory caching of items loaded from a local JSON file
type ItemCache struct {
	items     map[int]*Item // ID -> Item mapping for fast lookups
	itemsList []*Item       // All items as slice for iteration
	loaded    bool
	mutex     sync.RWMutex
	stats     ItemCacheStats
}

// ItemCacheStats tracks cache performance
type ItemCacheStats struct {
	LoadedItems  int
	LoadTime     time.Duration
	CacheHits    int64
	CacheMisses  int64
	LastLoadTime time.Time
}

// NewItemCache creates a new item cache
func NewItemCache() *ItemCache {
	return &ItemCache{
		items:     make(map[int]*Item),
		itemsList: make([]*Item, 0),
		loaded:    false,
	}
}

// LoadFromFile loads all items from a JSONL file (one JSON object per line)
func (ic *ItemCache) LoadFromFile(filePath string) error {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	startTime := time.Now()
	
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open items file %s: %w", filePath, err)
	}
	defer file.Close()

	// Clear existing data
	ic.items = make(map[int]*Item)
	ic.itemsList = make([]*Item, 0)

	scanner := bufio.NewScanner(file)
	itemCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		var item Item
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			// Skip invalid lines but continue processing
			continue
		}

		// Store in both map and slice
		ic.items[item.ID] = &item
		ic.itemsList = append(ic.itemsList, &item)
		itemCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading items file: %w", err)
	}

	ic.loaded = true
	ic.stats.LoadedItems = itemCount
	ic.stats.LoadTime = time.Since(startTime)
	ic.stats.LastLoadTime = time.Now()

	return nil
}

// GetByID retrieves an item by its ID from the cache
func (ic *ItemCache) GetByID(id int) (*Item, bool) {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()

	if !ic.loaded {
		ic.stats.CacheMisses++
		return nil, false
	}

	item, found := ic.items[id]
	if found {
		ic.stats.CacheHits++
	} else {
		ic.stats.CacheMisses++
	}
	
	return item, found
}

// GetByIDs retrieves multiple items by their IDs
func (ic *ItemCache) GetByIDs(ids []int) []*Item {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()

	if !ic.loaded {
		ic.stats.CacheMisses += int64(len(ids))
		return nil
	}

	results := make([]*Item, 0, len(ids))
	for _, id := range ids {
		if item, found := ic.items[id]; found {
			results = append(results, item)
			ic.stats.CacheHits++
		} else {
			ic.stats.CacheMisses++
		}
	}

	return results
}

// SearchItems performs in-memory search on cached items
func (ic *ItemCache) SearchItems(options ItemSearchOptions) []*Item {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()

	if !ic.loaded {
		return nil
	}

	var results []*Item
	count := 0

	// Set default limit if not specified
	limit := options.Limit
	if limit == 0 {
		limit = 50
	}

	for _, item := range ic.itemsList {
		if matchesSearchCriteria(item, options) {
			results = append(results, item)
			count++

			// Check if we've reached the limit
			if count >= limit {
				break
			}
		}
	}

	ic.stats.CacheHits++
	return results
}

// GetAll returns all cached items (use with caution for large datasets)
func (ic *ItemCache) GetAll() []*Item {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()

	if !ic.loaded {
		return nil
	}

	// Return a copy to prevent external modification
	result := make([]*Item, len(ic.itemsList))
	copy(result, ic.itemsList)
	
	ic.stats.CacheHits++
	return result
}

// IsLoaded returns whether the cache has been loaded
func (ic *ItemCache) IsLoaded() bool {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()
	return ic.loaded
}

// Stats returns cache statistics
func (ic *ItemCache) Stats() ItemCacheStats {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()
	return ic.stats
}

// Clear clears the cache
func (ic *ItemCache) Clear() {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	ic.items = make(map[int]*Item)
	ic.itemsList = make([]*Item, 0)
	ic.loaded = false
	ic.stats = ItemCacheStats{}
}

// Size returns the number of items in the cache
func (ic *ItemCache) Size() int {
	ic.mutex.RLock()
	defer ic.mutex.RUnlock()
	return len(ic.itemsList)
}