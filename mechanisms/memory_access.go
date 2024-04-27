package mechanisms

import (
	"container/list"
	"log"
	"sync"
)

type CacheMemoryAccess interface {
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	Delete(key string)
}

type cache struct {
	data      map[string]*list.Element // map of keys to list element
	evictList *list.List               // list to keep track of eviction order
	maxSize   int                      // maximum size of the cache
	mu        sync.Mutex               // mutex for thread safety
}

type entry struct {
	key   string
	value interface{}
}

func NewCacheMemoryAccess(maxSize int) CacheMemoryAccess {
	return &cache{
		data:      make(map[string]*list.Element),
		evictList: list.New(),
		maxSize:   maxSize,
	}
}

func (c *cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key already exists, update the value and move the entry to the front of the eviction list
	if elem, ok := c.data[key]; ok {
		elem.Value.(*entry).value = value
		c.evictList.MoveToFront(elem)
		log.Printf("Updated value for key %s", key)
		return
	}

	// Otherwise, add a new entry to the cache
	elem := c.evictList.PushFront(&entry{key, value})
	c.data[key] = elem

	// If cache size exceeds maximum size, evict the least recently used entry
	if c.evictList.Len() > c.maxSize {
		last := c.evictList.Back()
		c.evictList.Remove(last)
	}
	log.Printf("Added new entry with key %s", key)
}

func (c *cache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the key exists in the cache
	if elem, ok := c.data[key]; ok {
		// Move the entry to the front of the eviction list
		c.evictList.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	log.Printf("Key %s not found", key)
	return nil, false
}

func (c *cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the key exists in the cache
	if elem, ok := c.data[key]; ok {
		c.removeElement(elem)
		log.Printf("Deleted key %s", key)
		return
	}
	log.Printf("Key %s not found, cannot delete", key)
}

func (c *cache) removeElement(elem *list.Element) {
	c.evictList.Remove(elem)
	delete(c.data, elem.Value.(*entry).key)
}
