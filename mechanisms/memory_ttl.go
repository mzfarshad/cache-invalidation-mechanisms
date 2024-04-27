package mechanisms

import (
	"container/list"
	"sync"
	"time"
)

// Cache interface defines the methods for caching.
type CacheTTLMemory interface {
	Set(category, key string, value interface{})
	Get(category, key string) (interface{}, bool)
}

// MemoryCache implements the Cache interface using an in-memory store with time-based expiration.
type ttlMemoryCache struct {
	data     map[string]map[string]*cacheEntry // Data stored in memory
	ttl      time.Duration                     // Time-to-live for cache entries
	mu       sync.Mutex                        // Mutex for synchronization
	eviction *list.List                        // Eviction list for tracking least recently used items
	capacity int                               // Capacity of the cache
}

type cacheEntry struct {
	value     interface{}   // Cached value
	createdAt time.Time     // Time when the cache entry was created
	elem      *list.Element // Pointer to the corresponding element in the eviction list
	category  string        // Category of the cache entry
	key       string        // Key of the cache entry
}

// NewMemoryCache creates a new instance of MemoryCache with the specified time-to-live (TTL) and capacity.
func NewCacheTTLMemory(ttl time.Duration, capacity int) CacheTTLMemory {
	return &ttlMemoryCache{
		data:     make(map[string]map[string]*cacheEntry),
		ttl:      ttl,
		mu:       sync.Mutex{},
		eviction: list.New(),
		capacity: capacity,
	}
}

// Set adds a new key-value pair to the cache under the specified category.
func (t *ttlMemoryCache) Set(category, key string, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Create the category map if it doesn't exist
	if _, ok := t.data[category]; !ok {
		t.data[category] = make(map[string]*cacheEntry)
	}

	// Update existing entry if found, otherwise add a new one
	if entry, ok := t.data[category][key]; ok {
		entry.value = value
		entry.createdAt = time.Now()
		t.eviction.MoveToFront(entry.elem)
	} else {
		// Evict least recently used entry if cache is full
		if t.eviction.Len() >= t.capacity {
			last := t.eviction.Back()
			delete(t.data[last.Value.(*cacheEntry).category], last.Value.(*cacheEntry).key)
			t.eviction.Remove(last)
		}
		// Add new cache entry
		elem := t.eviction.PushFront(&cacheEntry{
			category:  category,
			key:       key,
			value:     value,
			createdAt: time.Now(),
		})
		t.data[category][key] = &cacheEntry{
			value: value,
			elem:  elem,
		}
	}
}

// Get retrieves the value associated with the given key from the cache under the specified category.
func (t *ttlMemoryCache) Get(category, key string) (interface{}, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if category exists
	if entries, ok := t.data[category]; ok {
		// Check if entry exists
		if entry, ok := entries[key]; ok {
			// Check if entry has expired
			if time.Since(entry.createdAt) > t.ttl {
				// Remove expired entry from cache and eviction list
				delete(entries, key)
				t.eviction.Remove(entry.elem)
				return nil, false
			}
			// Entry is valid, return value
			return entry.value, true
		}
	}
	// Entry not found
	return nil, false
}
