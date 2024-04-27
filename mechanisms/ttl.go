package mechanisms

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type CacheTTL interface {
	Set(category, key string, val interface{})
	Get(category, key string) (interface{}, error)
	Delete(category, key string)
}

type cacheTTL struct {
	data  map[string]map[string]*entryCacheTTL
	mutex sync.Mutex
	ttl   time.Duration
}

type entryCacheTTL struct {
	value     interface{}
	key       string
	category  string
	createdAt time.Time
}

func NewCacheTTL(ttl time.Duration) CacheTTL {
	return &cacheTTL{
		data:  make(map[string]map[string]*entryCacheTTL),
		mutex: sync.Mutex{},
		ttl:   ttl,
	}
}

func (c *cacheTTL) Set(category, key string, val interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.data[category]; !ok {
		c.data[category] = make(map[string]*entryCacheTTL)
	}
	if entry, ok := c.data[category][key]; ok {
		entry.value = val
		entry.createdAt = time.Now()
		log.Printf("Updated [%s][%s] : %v", category, key, val)
	} else {
		c.data[category][key] = &entryCacheTTL{
			value:     val,
			key:       key,
			category:  category,
			createdAt: time.Now(),
		}
		log.Printf("Successfully added [%s][%s]: %v", category, key, val)
	}
}

func (c *cacheTTL) Get(category, key string) (interface{}, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entries, ok := c.data[category]; ok {
		if entry, ok := entries[key]; ok {
			if time.Since(entry.createdAt) > c.ttl {
				delete(entries, entry.key)
				return nil, fmt.Errorf("%s timed out in cache", key)
			}
			return entry.value, nil
		}
	}
	return nil, fmt.Errorf("not found category : %s in cache", category)
}

func (c *cacheTTL) Delete(category, key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entries, ok := c.data[category]; ok {
		if entry, ok := entries[key]; ok {
			delete(entries, entry.key)
		}
		log.Printf("The key: %s not found", key)
		return
	}
	log.Printf("the category: %s not found", category)
}
