package cache

import (
	"io"
	"sync"
)

type RingBuffer struct {
	data  [][]byte
	size  int
	head  int
	tail  int
	count int
	mu    sync.Mutex
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([][]byte, size),
		size: size,
	}
}

func (r *RingBuffer) Put(frame []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.size == 0 {
		return
	}

	r.data[r.head] = frame
	r.head = (r.head + 1) % r.size
	if r.count == r.size {
		r.tail = (r.tail + 1) % r.size
	} else {
		r.count++
	}
}

func (r *RingBuffer) Get(index int) ([]byte, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if index < 0 || index >= r.count {
		return nil, false
	}
	idx := (r.tail + index) % r.size
	return r.data[idx], true
}

func (r *RingBuffer) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}

func (r *RingBuffer) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.head = 0
	r.tail = 0
	r.count = 0
	r.data = make([][]byte, r.size)
}

func (r *RingBuffer) Reader() io.Reader {
	return &emptyReader{}
}

type emptyReader struct{}

func (emptyReader) Read(p []byte) (int, error) {
	return 0, io.EOF
}
