package tile

import (
	"sync"
	"time"
)

// cacheEntry is one cached tile payload plus accounting metadata.
type cacheEntry struct {
	data      []byte
	createdAt time.Time
	hits      uint64
}

// TileCache is a bounded in-memory cache keyed by tile coordinate string.
type TileCache struct {
	mu       sync.RWMutex
	entries  map[string]*cacheEntry
	capacity int
	reads    uint64
	hits     uint64
}

// NewTileCache builds a cache that keeps at most capacity entries.
func NewTileCache(capacity int) *TileCache {
	if capacity < 1 {
		capacity = 1
	}
	return &TileCache{
		entries:  make(map[string]*cacheEntry),
		capacity: capacity,
	}
}

// Get returns the cached payload for a key.
func (c *TileCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.reads++
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	c.hits++
	entry.hits++
	data := make([]byte, len(entry.data))
	copy(data, entry.data)
	return data, true
}

// Put stores a payload under a key, evicting the least recently written
// entry when the capacity is exceeded.
func (c *TileCache) Put(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.entries[key]; exists {
		c.entries[key] = &cacheEntry{data: cloneBytes(data), createdAt: time.Now()}
		return
	}
	if len(c.entries) >= c.capacity {
		c.evictOneLocked()
	}
	c.entries[key] = &cacheEntry{data: cloneBytes(data), createdAt: time.Now()}
}

// Invalidate removes the listed keys and returns how many existed.
func (c *TileCache) Invalidate(keys []string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	removed := 0
	for _, key := range keys {
		if _, ok := c.entries[key]; ok {
			delete(c.entries, key)
			removed++
		}
	}
	return removed
}

// Clear removes every entry and returns the previous size.
func (c *TileCache) Clear() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	size := len(c.entries)
	c.entries = make(map[string]*cacheEntry)
	return size
}

// Size returns the number of cached entries.
func (c *TileCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// HitRate returns the ratio of cache reads that found an entry.
func (c *TileCache) HitRate() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.reads == 0 {
		return 0
	}
	return float64(c.hits) / float64(c.reads)
}

func (c *TileCache) evictOneLocked() {
	var oldest string
	var oldestAt time.Time
	for key, entry := range c.entries {
		if oldest == "" || entry.createdAt.Before(oldestAt) {
			oldest = key
			oldestAt = entry.createdAt
		}
	}
	if oldest != "" {
		delete(c.entries, oldest)
	}
}

func cloneBytes(data []byte) []byte {
	out := make([]byte, len(data))
	copy(out, data)
	return out
}
