package main

import "sync"

type concurrentMap[K comparable, V any] struct {
	items map[K]V
	sync.RWMutex
}

func NewConcurrentMap[K comparable, V any]() concurrentMap[K, V] {
	return concurrentMap[K, V]{
		items: make(map[K]V),
	}
}

func (c *concurrentMap[K, V]) Store(key K, value V) {
	c.Lock()
	defer c.Unlock()
	c.items[key] = value
}

func (c *concurrentMap[K, V]) Delete(key K) {
	c.Lock()
	defer c.Unlock()
	delete(c.items, key)
}

func (c *concurrentMap[K, V]) Load(key K) V {
	c.RLock()
	defer c.RUnlock()
	return c.items[key]
}

func (c *concurrentMap[K, V]) LoadAndStore(key K, value V) V {
	c.Lock()
	defer c.Unlock()
	temp := c.items[key]
	c.items[key] = value
	return temp
}

func (c *concurrentMap[K, V]) LoadAndDelete(key K) V {
	c.Lock()
	defer c.Unlock()
	temp := c.items[key]
	delete(c.items, key)
	return temp
}

func (c *concurrentMap[K, V]) RRange(f func(K, V) bool) {
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

func (c *concurrentMap[K, V]) Range(f func(K, V) bool) {
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

func (c *concurrentMap[K, V]) Count() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.items)
}

func (c *concurrentMap[K, V]) RApplyToOne(finder func(K, V) bool, applier func(K, V)) {
	c.RRange(func(k K, v V) bool {
		if finder(k, v) {
			applier(k, v)
			return false
		}
		return true
	})
}

func (c *concurrentMap[K, V]) ApplyToOne(finder func(K, V) bool, applier func(K, V)) {
	c.Range(func(k K, v V) bool {
		if finder(k, v) {
			applier(k, v)
			return false
		}
		return true
	})
}
