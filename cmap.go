package main

import "sync"

type cmap[K comparable, V any] struct {
	items map[K]V
	sync.RWMutex
}

func NewCmap[K comparable, V any]() cmap[K, V] {
	return cmap[K, V]{
		items: make(map[K]V),
	}
}

func (c *cmap[K, V]) Store(key K, value V) {
	c.Lock()
	defer c.Unlock()
	c.items[key] = value
}

func (c *cmap[K, V]) Delete(key K) {
	c.Lock()
	defer c.Unlock()
	delete(c.items, key)
}

func (c *cmap[K, V]) Load(key K) V {
	c.RLock()
	defer c.RUnlock()
	return c.items[key]
}

func (c *cmap[K, V]) LoadAndStore(key K, value V) V {
	c.Lock()
	defer c.Unlock()
	temp := c.items[key]
	c.items[key] = value
	return temp
}

func (c *cmap[K, V]) LoadAndDelete(key K) V {
	c.Lock()
	defer c.Unlock()
	temp := c.items[key]
	delete(c.items, key)
	return temp
}

func (c *cmap[K, V]) RRange(f func(K, V) bool) {
	c.RLock()
	defer c.RUnlock()
	for key := range c.items {
		value := c.items[key]
		cont := f(key, value)
		if !cont {
			return
		}
	}
}

func (c *cmap[K, V]) Range(f func(K, V) bool) {
	c.Lock()
	defer c.Unlock()
	for key := range c.items {
		value := c.items[key]
		cont := f(key, value)
		if !cont {
			return
		}
	}
}

func (c *cmap[K, V]) Count() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.items)
}
