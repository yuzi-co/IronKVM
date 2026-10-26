package urlmedia

import (
	"bytes"
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"
)

func mustIP(t *testing.T, s string) net.IP {
	t.Helper()
	ip := net.ParseIP(s)
	if ip == nil {
		t.Fatalf("bad IP %q", s)
	}
	return ip
}

// memSource is a source in memory that counts its reads and can fail.
type memSource struct {
	mu    sync.Mutex
	data  []byte
	reads []int64 // offsets read
	fail  error
	delay time.Duration
}

func (m *memSource) Size() int64 { return int64(len(m.data)) }

func (m *memSource) ReadAt(ctx context.Context, p []byte, off int64) error {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reads = append(m.reads, off)
	if m.fail != nil {
		return m.fail
	}
	if off < 0 || off+int64(len(p)) > int64(len(m.data)) {
		return errors.New("read past the end")
	}
	copy(p, m.data[off:])
	return nil
}

func (m *memSource) readCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.reads)
}

func (m *memSource) setFail(err error) {
	m.mu.Lock()
	m.fail = err
	m.mu.Unlock()
}

func TestCacheReadsAcrossBlocksAndTheShortLastBlock(t *testing.T) {
	src := &memSource{data: testImage(10*1024 + 300)}
	c := NewCache(src, 1024, 4, 0)
	defer c.Close()

	for _, tc := range []struct{ off, n int }{{0, 10}, {1000, 100}, {3000, 3000}, {10*1024 + 200, 100}, {0, 10*1024 + 300}} {
		p := make([]byte, tc.n)
		if err := c.ReadAt(context.Background(), p, int64(tc.off)); err != nil {
			t.Fatalf("ReadAt(%d, %d): %v", tc.off, tc.n, err)
		}
		if !bytes.Equal(p, src.data[tc.off:tc.off+tc.n]) {
			t.Fatalf("ReadAt(%d, %d) returned the wrong bytes", tc.off, tc.n)
		}
	}
}

func TestCacheRefusesReadsPastTheEnd(t *testing.T) {
	src := &memSource{data: testImage(4096)}
	c := NewCache(src, 1024, 4, 0)
	defer c.Close()

	if err := c.ReadAt(context.Background(), make([]byte, 10), 4090); !errors.Is(err, ErrOutOfRange) {
		t.Fatalf("err = %v, want ErrOutOfRange", err)
	}
}

func TestCacheHitsDoNotReadTheSource(t *testing.T) {
	src := &memSource{data: testImage(8192)}
	c := NewCache(src, 1024, 4, 0)
	defer c.Close()

	p := make([]byte, 100)
	for i := 0; i < 5; i++ {
		if err := c.ReadAt(context.Background(), p, 50); err != nil {
			t.Fatal(err)
		}
	}
	if n := src.readCount(); n != 1 {
		t.Fatalf("%d source reads for one block read five times, want 1", n)
	}
	st := c.Stats()
	if st.Hits != 4 || st.Misses != 1 {
		t.Fatalf("stats = %+v, want 4 hits and 1 miss", st)
	}
}

func TestCacheEvictsTheLeastRecentlyUsedBlock(t *testing.T) {
	src := &memSource{data: testImage(8 * 1024)}
	c := NewCache(src, 1024, 2, 0)
	defer c.Close()

	read := func(block int) {
		t.Helper()
		if err := c.ReadAt(context.Background(), make([]byte, 1), int64(block*1024)); err != nil {
			t.Fatal(err)
		}
	}
	read(0)
	read(1)
	read(0) // 0 is now newer than 1
	read(2) // evicts 1
	before := src.readCount()
	read(0)
	if src.readCount() != before {
		t.Fatal("block 0 was evicted, but block 1 was the least recently used")
	}
	read(1)
	if src.readCount() != before+1 {
		t.Fatal("block 1 was still cached past the capacity")
	}
	if n := c.cachedBlocks(); n > 2 {
		t.Fatalf("%d blocks cached, capacity 2", n)
	}
}

func TestCacheDoesNotKeepErrors(t *testing.T) {
	src := &memSource{data: testImage(4096)}
	c := NewCache(src, 1024, 4, 0)
	defer c.Close()

	src.setFail(errors.New("network down"))
	if err := c.ReadAt(context.Background(), make([]byte, 10), 0); err == nil {
		t.Fatal("ReadAt succeeded with the source failing")
	}
	src.setFail(nil)
	p := make([]byte, 10)
	if err := c.ReadAt(context.Background(), p, 0); err != nil {
		t.Fatalf("ReadAt after the source recovered: %v", err)
	}
	if !bytes.Equal(p, src.data[:10]) {
		t.Fatal("wrong bytes after recovery")
	}
}

func TestCacheReadsOneBlockOnceForConcurrentReaders(t *testing.T) {
	src := &memSource{data: testImage(4096), delay: 50 * time.Millisecond}
	c := NewCache(src, 1024, 4, 0)
	defer c.Close()

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.ReadAt(context.Background(), make([]byte, 10), 100); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if n := src.readCount(); n != 1 {
		t.Fatalf("%d source reads for one block, want 1", n)
	}
}

func TestCacheReadsAheadOnSequentialReads(t *testing.T) {
	src := &memSource{data: testImage(16 * 1024)}
	c := NewCache(src, 1024, 8, 3)
	defer c.Close()

	p := make([]byte, 1024)
	_ = c.ReadAt(context.Background(), p, 0)
	_ = c.ReadAt(context.Background(), p, 1024) // sequential: blocks 2..4 are fetched ahead

	deadline := time.Now().Add(2 * time.Second)
	for c.cachedBlocks() < 5 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	before := src.readCount()
	for b := 2; b <= 4; b++ {
		if err := c.ReadAt(context.Background(), p, int64(b*1024)); err != nil {
			t.Fatal(err)
		}
	}
	// Reading 2..4 was all hits; the reads may have started the next
	// window, but none of 2..4 came from the source again.
	src.mu.Lock()
	later := append([]int64(nil), src.reads[before:]...)
	src.mu.Unlock()
	for _, off := range later {
		if off >= 2*1024 && off <= 4*1024 {
			t.Fatalf("block at %d was read again, not ahead", off)
		}
	}
	if st := c.Stats(); st.Prefetched < 3 {
		t.Fatalf("stats = %+v, want at least 3 blocks read ahead", st)
	}
}

func TestCacheDoesNotReadAheadOnRandomReads(t *testing.T) {
	src := &memSource{data: testImage(16 * 1024)}
	c := NewCache(src, 1024, 8, 3)
	defer c.Close()

	for _, b := range []int{7, 2, 12, 0} {
		_ = c.ReadAt(context.Background(), make([]byte, 1), int64(b*1024))
	}
	time.Sleep(50 * time.Millisecond)
	if n := src.readCount(); n != 4 {
		t.Fatalf("%d source reads for 4 random blocks, want 4", n)
	}
}
