package mechanisms

import (
	"container/list"
	"fmt"
	"log"
	"sync"
)

type CacheAccess interface {
	Set(category, key string, val interface{})
	Get(category, key string) (interface{}, error)
	Delete(category, key string)
}

type cacheAccess struct {
	data      map[string]map[string]*entryCacheAccess
	evication *list.List
	mutex     sync.Mutex
}

type entryCacheAccess struct {
	value    interface{}
	category string
	key      string
	elem     *list.Element
}

func NewCacheAccess() CacheAccess {
	return &cacheAccess{
		data:      make(map[string]map[string]*entryCacheAccess),
		evication: list.New(),
		mutex:     sync.Mutex{},
	}
}

func (c *cacheAccess) Set(category, key string, val interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.data[category]; !ok {
		c.data[category] = make(map[string]*entryCacheAccess)
		log.Printf("added new category: %s in cacheAccess", category)
	}
	if entry, ok := c.data[category][key]; ok {
		entry.value = val
		c.evication.MoveToFront(entry.elem)
		log.Printf("Updated [%s][%s] : %v", category, key, val)
	} else {
		elem := c.evication.PushFront(&entryCacheAccess{
			value:    val,
			key:      key,
			category: category,
		})
		c.data[category][key] = &entryCacheAccess{
			value: val,
			elem:  elem,
		}
		log.Printf("Successfully added [%s][%s]: %v in cacheAccess", category, key, val)
	}

}

func (c *cacheAccess) Get(category, key string) (interface{}, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entries, ok := c.data[category]; ok {
		if entry, ok := entries[key]; ok {
			c.evication.MoveToFront(entry.elem)
			return entry.value, nil
		} else {
			return nil, fmt.Errorf("key not found")
		}
	}
	return nil, fmt.Errorf("category not found")
}

func (c *cacheAccess) Delete(category, key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if entries, ok := c.data[category]; ok {
		if entry, ok := entries[key]; ok {
			c.evication.Remove(entry.elem)
			delete(c.data, entry.key)
			log.Printf("Successfully deleted key : %s\n", key)
			return
		} else {
			log.Printf("not found key : %s", key)
			return
		}
	}
	log.Printf("not found category : %s", category)
}
