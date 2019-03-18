package main

import (
	"sync"
	"time"
)

type item struct {
	value      string
	lastAccess int64
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

func (m *GitReleasesCache) Put(k, v string) {
	m.l.Lock()
	it, ok := m.items[k]
	if !ok {
		it = &item{value: v}
		m.items[k] = it
	}
	it.lastAccess = time.Now().Unix()
	m.l.Unlock()
}

func (m *GitReleasesCache) Get(k string) (v string) {
	m.l.RLock()
	if it, ok := m.items[k]; ok {
		v = it.value
		it.lastAccess = time.Now().Unix()
	}
	m.l.RUnlock()
	return
}
