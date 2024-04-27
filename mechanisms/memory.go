package mechanisms

import (
	"container/list"
	"fmt"
	"log"
	"sync"
)

type CacheMemory interface {
	Set(category, key string, val interface{})
	Get(category, key string) (interface{}, error)
	Delete(category, key string)
}

type cacheMemory struct {
	data           map[string]map[string]*entryCacheMemory
	mutex          sync.Mutex
	evicationMutex sync.Mutex
	evication      *list.List
	sizeKey        int
}

type entryCacheMemory struct {
	value    interface{}
	key      string
	category string
}

func NewCacheMemory(sizeKey int) CacheMemory {
	return &cacheMemory{
		data:           make(map[string]map[string]*entryCacheMemory),
		sizeKey:        sizeKey,
		evication:      list.New(),
		mutex:          sync.Mutex{},
		evicationMutex: sync.Mutex{},
	}
}

func (c *cacheMemory) Set(category, key string, val interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.data[category]; !ok {
		c.data[category] = make(map[string]*entryCacheMemory)
	}
	if entry, ok := c.data[category][key]; ok {
		entry.value = val
		log.Printf("Updated [%s][%s] : %v", category, key, val)
	} else {
		if len(c.data[category]) > c.sizeKey-1 {
			err := c.evicatLastKey(category, key)
			if err != nil {
				log.Println(err)
			}
		}
		c.data[category][key] = &entryCacheMemory{
			value:    val,
			category: category,
			key:      key,
		}
		c.evication.PushFront(&entryCacheMemory{
			category: category,
			key:      key,
			value:    val,
		})
		log.Printf("Successfully added [%s][%s] : %v", category, key, val)
	}
}

func (c *cacheMemory) Get(category, key string) (interface{}, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entries, ok := c.data[category]; ok {
		if entry, ok := entries[key]; ok {
			return entry.value, nil
		}
		return nil, fmt.Errorf("not found key: %s", key)
	}
	return nil, fmt.Errorf("not found category: %s", category)
}

func (c *cacheMemory) Delete(category, key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entries, ok := c.data[category]; ok {
		if entry, ok := entries[key]; ok {
			delete(c.data, entry.key)
			log.Printf("Successfully deleted category: %s, key: %s", category, key)
			return
		}
		log.Printf("not found key: %s", key)
	}
	log.Printf("not found category: %s", category)
}

func (c *cacheMemory) evicatLastKey(category, key string) error {
	c.evicationMutex.Lock()
	defer c.evicationMutex.Unlock()

	last := c.evication.Back()
	if last != nil {
		entry := last.Value.(*entryCacheMemory)
		delete(c.data[category], entry.key)
		c.evication.Remove(last)
		log.Printf("Due to exceeding the allowed number of keys, key: %s was removed and key: %s was replaced",
			entry.key, key)
		return nil
	}
	return fmt.Errorf("there are no entities to remove from the bottom of the list")
}
