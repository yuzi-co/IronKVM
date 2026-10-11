package hostlog

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

// DiskDir holds the optional on-disk log. It is on /data, which both slots
// share and which survives an update.
const DiskDir = "/data/ironkvm/hostlog"

const (
	diskCurrent  = "hostlog.jsonl"
	diskPrevious = "hostlog.jsonl.1"
)

// Disk keeps the log in two files that take turns: lines go to the current
// one, and when it holds half the cap it becomes the previous one and the old
// previous one is deleted. The two never total more than the cap.
type Disk struct {
	dir string

	mu   sync.Mutex
	max  int64
	f    *os.File
	w    *bufio.Writer
	size int64
}

// NewDisk returns a log in dir that holds up to maxBytes.
func NewDisk(dir string, maxBytes int64) *Disk {
	return &Disk{dir: dir, max: maxBytes}
}

// SetMax changes the cap. A smaller one takes effect at the next rotation.
func (d *Disk) SetMax(maxBytes int64) {
	d.mu.Lock()
	d.max = maxBytes
	d.mu.Unlock()
}

func (d *Disk) openLocked() error {
	if d.f != nil {
		return nil
	}
	if err := os.MkdirAll(d.dir, 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(d.dir, diskCurrent), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	st, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return err
	}
	d.f, d.w, d.size = f, bufio.NewWriterSize(f, 16*1024), st.Size()
	return nil
}

// Append writes one entry. It buffers: Flush makes it durable.
func (d *Disk) Append(e Entry) error {
	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	line = append(line, '\n')

	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.openLocked(); err != nil {
		return err
	}
	if d.size+int64(len(line)) > d.max/2 && d.size > 0 {
		if err := d.rotateLocked(); err != nil {
			return err
		}
	}
	n, err := d.w.Write(line)
	d.size += int64(n)
	return err
}

func (d *Disk) rotateLocked() error {
	if err := d.closeLocked(); err != nil {
		return err
	}
	cur := filepath.Join(d.dir, diskCurrent)
	if err := os.Rename(cur, filepath.Join(d.dir, diskPrevious)); err != nil {
		return err
	}
	return d.openLocked()
}

func (d *Disk) closeLocked() error {
	if d.f == nil {
		return nil
	}
	err := d.w.Flush()
	if cerr := d.f.Close(); err == nil {
		err = cerr
	}
	d.f, d.w, d.size = nil, nil, 0
	return err
}

// Flush writes what is buffered.
func (d *Disk) Flush() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.w == nil {
		return nil
	}
	return d.w.Flush()
}

// Close flushes and closes the files. The log can be appended to again.
func (d *Disk) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.closeLocked()
}

// ReadAll returns the stored entries, oldest first. A line that does not parse
// is skipped: a power cut can tear the last one.
func (d *Disk) ReadAll() []Entry {
	_ = d.Flush()

	var out []Entry
	for _, name := range []string{diskPrevious, diskCurrent} {
		data, err := os.ReadFile(filepath.Join(d.dir, name))
		if err != nil {
			continue
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var e Entry
			if json.Unmarshal(line, &e) == nil {
				out = append(out, e)
			}
		}
	}
	return out
}

// Size returns how many bytes the files hold.
func (d *Disk) Size() int64 {
	_ = d.Flush()

	var total int64
	for _, name := range []string{diskPrevious, diskCurrent} {
		if st, err := os.Stat(filepath.Join(d.dir, name)); err == nil {
			total += st.Size()
		}
	}
	return total
}

// Clear deletes the files.
func (d *Disk) Clear() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_ = d.closeLocked()
	var first error
	for _, name := range []string{diskPrevious, diskCurrent} {
		if err := os.Remove(filepath.Join(d.dir, name)); err != nil && !errors.Is(err, os.ErrNotExist) && first == nil {
			first = err
		}
	}
	return first
}
