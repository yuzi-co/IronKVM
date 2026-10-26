package urlmedia

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

// ErrOutOfRange is a read that does not fit inside the image.
var ErrOutOfRange = errors.New("read outside the image")

// backend is what the cache reads from, and what NBD serves.
type backend interface {
	Size() int64
	ReadAt(ctx context.Context, p []byte, off int64) error
}

// The board's cache: 16 blocks of 512 KiB, 8 MiB in all, out of 256 MB. A
// host booting an installer reads mostly in order, and the kernel's own
// read-ahead asks for up to 128 KiB at a time, so one block answers several
// requests and the 4 blocks read ahead keep 2 MiB in flight on a long read.
const (
	DefaultBlockSize = 512 << 10
	DefaultBlocks    = 16
	DefaultReadAhead = 4
)

// CacheStats counts what the cache did.
type CacheStats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Prefetched int64 `json:"prefetched"`
}

type block struct {
	index int64
	data  []byte
	elem  *list.Element
}

// fetch is a block read in flight, which every reader of the block waits on.
type fetch struct {
	done chan struct{}
	data []byte
	err  error
}

// Cache is a bounded LRU of aligned blocks over a backend, with read-ahead
// once the reads run in order.
type Cache struct {
	src       backend
	blockSize int64
	capacity  int
	readAhead int

	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.Mutex
	blocks   map[int64]*block
	lru      *list.List // front is the most recently used
	inflight map[int64]*fetch
	last     int64 // the last block a read touched
	ahead    bool  // a read-ahead is running

	hits, misses, prefetched atomic.Int64
}

// NewCache caches blocks of blockSize bytes, at most blocks of them, and
// reads up to readAhead blocks ahead of a sequential reader.
func NewCache(src backend, blockSize, blocks, readAhead int) *Cache {
	ctx, cancel := context.WithCancel(context.Background())
	return &Cache{
		src:       src,
		blockSize: int64(blockSize),
		capacity:  max(blocks, 1),
		readAhead: readAhead,
		ctx:       ctx,
		cancel:    cancel,
		blocks:    map[int64]*block{},
		lru:       list.New(),
		inflight:  map[int64]*fetch{},
		last:      -2,
	}
}

// Size is the backend's size.
func (c *Cache) Size() int64 { return c.src.Size() }

// Stats returns the counters.
func (c *Cache) Stats() CacheStats {
	return CacheStats{Hits: c.hits.Load(), Misses: c.misses.Load(), Prefetched: c.prefetched.Load()}
}

// Close stops the read-ahead.
func (c *Cache) Close() { c.cancel() }

func (c *Cache) cachedBlocks() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.blocks)
}

// ReadAt fills p from off.
func (c *Cache) ReadAt(ctx context.Context, p []byte, off int64) error {
	size := c.src.Size()
	if off < 0 || off+int64(len(p)) > size {
		return ErrOutOfRange
	}
	for len(p) > 0 {
		index := off / c.blockSize
		data, err := c.block(ctx, index, false)
		if err != nil {
			return err
		}
		n := copy(p, data[off-index*c.blockSize:])
		p = p[n:]
		off += int64(n)
		c.noteRead(index)
	}
	return nil
}

// noteRead starts a read-ahead when this read follows the last one.
func (c *Cache) noteRead(index int64) {
	if c.readAhead <= 0 {
		return
	}
	c.mu.Lock()
	sequential := index == c.last+1 || (index == c.last && c.last >= 0)
	c.last = index
	start := c.ahead
	if sequential && !c.ahead {
		c.ahead = true
	}
	c.mu.Unlock()
	if sequential && !start {
		go c.prefetch(index)
	}
}

// prefetch reads the blocks after index that are not cached yet.
func (c *Cache) prefetch(index int64) {
	defer func() {
		c.mu.Lock()
		c.ahead = false
		c.mu.Unlock()
	}()
	last := (c.src.Size() - 1) / c.blockSize
	for i := index + 1; i <= index+int64(c.readAhead) && i <= last; i++ {
		if c.ctx.Err() != nil {
			return
		}
		if _, err := c.block(c.ctx, i, true); err != nil {
			return
		}
	}
}

// block returns the block's bytes, from the cache or the backend. Two
// readers of one missing block share one read.
func (c *Cache) block(ctx context.Context, index int64, ahead bool) ([]byte, error) {
	c.mu.Lock()
	if b, ok := c.blocks[index]; ok {
		c.lru.MoveToFront(b.elem)
		c.mu.Unlock()
		if !ahead {
			c.hits.Add(1)
		}
		return b.data, nil
	}
	if f, ok := c.inflight[index]; ok {
		c.mu.Unlock()
		if !ahead {
			// A block the read-ahead is already fetching counts as a hit.
			c.hits.Add(1)
		}
		select {
		case <-f.done:
			return f.data, f.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	f := &fetch{done: make(chan struct{})}
	c.inflight[index] = f
	c.mu.Unlock()

	if ahead {
		c.prefetched.Add(1)
	} else {
		c.misses.Add(1)
	}
	start := index * c.blockSize
	n := min(c.blockSize, c.src.Size()-start)
	data := make([]byte, n)
	// The read is the cache's, not the caller's: a reader that gives up
	// must not fail the others waiting on the block.
	err := c.src.ReadAt(c.ctx, data, start)

	c.mu.Lock()
	delete(c.inflight, index)
	if err == nil {
		b := &block{index: index, data: data}
		b.elem = c.lru.PushFront(b)
		c.blocks[index] = b
		for c.lru.Len() > c.capacity {
			old := c.lru.Back()
			c.lru.Remove(old)
			delete(c.blocks, old.Value.(*block).index)
		}
	} else {
		data = nil
	}
	f.data, f.err = data, err
	close(f.done)
	c.mu.Unlock()
	return data, err
}
