package cache

import (
	"os"
	"path/filepath"
	"sync"
)

type ThumbnailCache struct {
	baseDir string
	maxSize int64
	mu      sync.Mutex
	index   map[string]int64
}

func NewThumbnailCache(baseDir string, maxSize int64) (*ThumbnailCache, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &ThumbnailCache{
		baseDir: baseDir,
		maxSize: maxSize,
		index:   make(map[string]int64),
	}, nil
}

func (c *ThumbnailCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := filepath.Join(c.baseDir, key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	fi, err := os.Stat(path)
	if err == nil {
		c.index[key] = fi.Size()
	}
	return data, true
}

func (c *ThumbnailCache) Put(key string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := filepath.Join(c.baseDir, key)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	c.index[key] = int64(len(data))
	c.evict()
	return nil
}

func (c *ThumbnailCache) evict() {
	if c.maxSize <= 0 {
		return
	}

	var total int64
	for _, size := range c.index {
		total += size
	}

	if total <= c.maxSize {
		return
	}

	type entry struct {
		key  string
		path string
		size int64
	}

	var entries []entry
	for k, s := range c.index {
		entries = append(entries, entry{key: k, path: filepath.Join(c.baseDir, k), size: s})
	}

	for _, e := range entries {
		os.Remove(e.path)
		total -= e.size
		delete(c.index, e.key)
		if total <= c.maxSize {
			break
		}
	}
}

func (c *ThumbnailCache) Clear(ctx interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	files, _ := filepath.Glob(filepath.Join(c.baseDir, "*"))
	for _, f := range files {
		os.Remove(f)
	}
	c.index = make(map[string]int64)
	return nil
}
