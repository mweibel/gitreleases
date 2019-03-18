package main

import (
	"sync"
	"time"
)

type item struct {
	value      string
	err        error
	lastAccess int64
}

type Cacher interface {
	Put(k, v string, err error)
	Get(k string) (v string, err error)
}

type GitReleasesCache struct {
	items map[string]*item
	l     sync.RWMutex
}

// NewCache is a simple cache with expiration time (TTL).
//
// I was a bit lazy so I took the code mostly from:
// https://stackoverflow.com/questions/25484122/map-with-ttl-option-in-go
//
// I just changed the mutex to an RWMutex and added the possibility to do specify the ticking time.
func NewCache(ln int, maxTTL int, tickInterval time.Duration) (m *GitReleasesCache) {
	m = &GitReleasesCache{items: make(map[string]*item, ln)}
	go func() {
		for now := range time.Tick(tickInterval) {
			m.l.Lock()
			for k, v := range m.items {
				if now.Unix()-v.lastAccess > int64(maxTTL) {
					delete(m.items, k)
				}
			}
			m.l.Unlock()
		}
	}()
	return
}

// Put adds `v` using the key `k` to the cache. Error is optional and can be used to store an error in the cache for faster error lookups.
func (m *GitReleasesCache) Put(k, v string, err error) {
	m.l.Lock()
	it, ok := m.items[k]
	if !ok {
		it = &item{value: v, err: err}
		m.items[k] = it
	}
	it.lastAccess = time.Now().Unix()
	m.l.Unlock()
}

// Get retrieves by key `k` the value. If `err` is non nil, this probably means an error has been cached explicitely.
//
// A not cached value is indicated using an empty string for `v` and a nil error.
func (m *GitReleasesCache) Get(k string) (v string, err error) {
	m.l.RLock()
	if it, ok := m.items[k]; ok {
		v = it.value
		err = it.err
		it.lastAccess = time.Now().Unix()
	}
	m.l.RUnlock()
	return
}
