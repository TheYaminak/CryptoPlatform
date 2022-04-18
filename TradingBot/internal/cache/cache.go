package cache

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"

	gc "github.com/patrickmn/go-cache"
)

type Cache struct {
	Cache *gc.Cache
}

// NewCache func return cache.
func NewCache() *Cache {
	var c Cache

	c.loadCache()

	return &c
}

// loadCache func load cache from file.
func (c *Cache) loadCache() {
	jsonFile, err := os.Open("internal/cache/cache.json")
	if err != nil {
		log.Fatalf("os.Open(cache.json) err: %v\n", err)
	}
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Fatalf("ioutil.ReadAll(jsonFile) err: %v\n", err)
	}

	if len(byteValue) == 0 {
		c.Cache = gc.New(0, 0)

		return
	}

	var items map[string]gc.Item

	err = json.Unmarshal(byteValue, &items)
	if err != nil {
		log.Fatalf("loadCache() json.Unmarshal(byteValue) err: %v\n", err)
	}

	c.Cache = gc.NewFrom(0, 0, items)
}

// SaveCache func save cache to file.
func (c *Cache) SaveCache() {
	items := c.Cache.Items()

	b, err := json.Marshal(items)
	if err != nil {
		log.Printf("json.Marshal(items) err: %v\n", err)
	}

	file, err := os.OpenFile("internal/cache/cache.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Printf("os.OpenFile(cache.json) err: %v\n", err)
	}

	defer file.Close()

	// encoder := json.NewEncoder(file)
	// err = encoder.Encode(b)
	_, err = io.WriteString(file, string(b))
	if err != nil {
		log.Printf("SaveCache() io.WriteString() err: %v\n", err)
	}
}
