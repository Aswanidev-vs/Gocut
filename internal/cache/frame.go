package cache

import (
	"sync"
	"time"
)

type frameEntry struct {
	data    []byte
	expires time.Time
}

type FrameCache struct {
	mu      sync.Mutex
	entries map[string]frameEntry
	maxSize int
	ttl     time.Duration
}

func NewFrameCache(maxSize int, ttl time.Duration) *FrameCache {
	return &FrameCache{
		entries: make(map[string]frameEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

func (c *FrameCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expires) {
		delete(c.entries, key)
		return nil, false
	}
	return e.data, true
}

func (c *FrameCache) Put(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.maxSize > 0 && len(c.entries) >= c.maxSize {
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}

	c.entries[key] = frameEntry{
		data:    data,
		expires: time.Now().Add(c.ttl),
	}
}

func (c *FrameCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]frameEntry)
}

func (c *FrameCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}
