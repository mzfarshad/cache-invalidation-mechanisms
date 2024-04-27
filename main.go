package main

import (
	"log"
	"time"

	"github.com/mzfarshad/cache-invalidation-mechanisms/mechanisms"
)

type category string

const (
	cars category = "cars"
)

type status string

const (
	color status = "color"
	name  status = "name"
	model status = "model"
	a     status = "a"
)

func main() {
	newCacheMemory := mechanisms.NewCacheMemory(3)

	newCacheMemory.Set(string(cars), string(color), "white")
	newCacheMemory.Set(string(cars), string(name), "pride")
	newCacheMemory.Set(string(cars), string(model), 2010)
	newCacheMemory.Set(string(cars), string(a), "qw")

	log.Println("-----------------------------------------")
	newCacheMemory.Delete(string(cars), string(a))

	newCacheAccess := mechanisms.NewCacheAccess()

	newCacheAccess.Set(string(cars), string(color), "black")
	newCacheAccess.Set(string(cars), string(name), "samand")

	name, err := newCacheAccess.Get(string(cars), string(name))
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(name)
	log.Println("-----------------------------------------")

	newCacheTTL := mechanisms.NewCacheTTL(10 * time.Second)

	newCacheTTL.Set(string(cars), string(color), "blue")
	time.Sleep(15 * time.Second)

	color, err := newCacheTTL.Get(string(cars), string(color))
	if err != nil {
		log.Println(err)
	} else {
		log.Println(color.(string))
	}
}
