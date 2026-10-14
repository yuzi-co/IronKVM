package hid

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

// A host that stops reading an endpoint shows up on the write path: f_hid holds
// the report it has not read, the next write finds the endpoint busy, and the
// write timeout turns that into a stall (health.go). That is how it works on
// the vendor kernel.
//
// The mainline kernel's f_hid has keep_latest (ironkvm-dist #111), which
// S03usbdev sets on the keyboard and the relative mouse. Under it a write
// never waits: reports queue behind the one at the controller, and once the
// host stops reading only the newest is kept. That cures the stuck key, and it
// also means no write ever times out, so a host that stopped reading was never
// reported. The same kernel says how long the report at the controller has
// waited for the host, in /sys/class/hidg/hidgN/pending_ms (ironkvm-dist
// #133). A write that finds it at hostReadStallAfter or more fails as a
// timed-out write would, so the stall warning, the web UI's state and the
// release of held input after a failed write all work as they do on the vendor
// kernel. Under keep_latest the release still reaches the queue, as its newest
// report, so the host reads it when it reads again.
//
// The capability is read at run time, once per opened handle: a kernel without
// the file (the vendor kernel, or a mainline one without the patch) keeps the
// write timeout as its only signal, and nothing here runs per report.
const (
	// hostReadStallAfter is how long the report at the controller may wait
	// before the host counts as not reading. It matches HIDG_LATEST_STALE_MS
	// in the keep_latest patch, after which f_hid itself cuts the queue to the
	// newest report, so the kernel and the server agree on when the host
	// stopped. A host that reads waits one polling interval, a few
	// milliseconds even for a firmware that polls slowly.
	hostReadStallAfter = 100 * time.Millisecond

	// hostReadRecheckMin is the shortest delay before a recheck. The kernel
	// counts in jiffies, so a value read just short of the limit may be a
	// tick behind.
	hostReadRecheckMin = 10 * time.Millisecond
)

// errHostNotReading is the failure of a write the kernel took while the host
// had left the report before it unread for hostReadStallAfter. It wraps
// os.ErrDeadlineExceeded because it is the same fault as a write that timed
// out, and classifyWriteResult reports both as a stall.
var errHostNotReading = fmt.Errorf("the host has not read a report for %s: %w",
	hostReadStallAfter, os.ErrDeadlineExceeded)

// hidgClassDir is where the f_hid class devices live. A variable so tests can
// point it at a scratch directory.
var hidgClassDir = "/sys/class/hidg"

// pendingReader says how long the report at the controller has waited for the
// host, and 0 when none is waiting.
type pendingReader interface {
	Pending() (time.Duration, error)
	Close() error
}

// openPendingReader opens the kernel's pending_ms for one device node, and
// returns nil when the kernel has none. A variable so tests can stand in a
// fake source.
var openPendingReader = func(devicePath string) pendingReader {
	file, err := os.Open(filepath.Join(hidgClassDir, filepath.Base(devicePath), "pending_ms"))
	if err != nil {
		return nil
	}
	return &sysfsPending{file: file}
}

// sysfsPending reads pending_ms through one open descriptor. A sysfs
// attribute is read again from offset 0, so each check is one pread.
type sysfsPending struct {
	file *os.File
	// buf is here rather than on the stack, which a read through *os.File
	// would move to the heap on every report.
	buf [24]byte
}

func (s *sysfsPending) Pending() (time.Duration, error) {
	n, err := s.file.ReadAt(s.buf[:], 0)
	if n == 0 {
		if err == nil {
			err = io.ErrUnexpectedEOF
		}
		return 0, err
	}
	ms, err := parseMillis(s.buf[:n])
	if err != nil {
		return 0, err
	}
	return time.Duration(ms) * time.Millisecond, nil
}

func (s *sysfsPending) Close() error {
	return s.file.Close()
}

var errBadPending = errors.New("pending_ms is not a number")

// parseMillis reads the decimal number sysfs prints, without allocating.
func parseMillis(raw []byte) (uint64, error) {
	var ms uint64
	digits := 0
	for _, c := range raw {
		if c == '\n' {
			break
		}
		if c < '0' || c > '9' || digits == 19 {
			return 0, errBadPending
		}
		ms = ms*10 + uint64(c-'0')
		digits++
	}
	if digits == 0 {
		return 0, errBadPending
	}
	return ms, nil
}

// readTimer is the part of *time.Timer the watch uses, so the tests can fire
// it by hand.
type readTimer interface {
	Reset(time.Duration) bool
	Stop() bool
}

// newReadTimer is time.AfterFunc behind readTimer. A variable for the tests.
var newReadTimer = func(d time.Duration, f func()) readTimer {
	return time.AfterFunc(d, f)
}

// hostReadWatch is one endpoint's view of the kernel's pending_ms. Every field
// is guarded by the endpoint's lock, the one hidDevice.mu names.
type hostReadWatch struct {
	probed bool
	reader pendingReader
	// timer rechecks after the last write. A write finds a stall only while
	// writes keep coming; this finds one after the last of them, so a host that
	// stops reading is reported even when nothing more is typed.
	timer readTimer
}

// readerNoLock returns the kernel's source for this endpoint, opening it the
// first time it is asked after the handle was opened, and nil when the kernel
// has none.
func (w *hostReadWatch) readerNoLock(devicePath string) pendingReader {
	if !w.probed {
		w.reader = openPendingReader(devicePath)
		w.probed = true
		if w.reader != nil {
			log.Debugf("%s: the kernel reports how long the host leaves a report unread", devicePath)
		}
	}
	return w.reader
}

// pendingNoLock reads the source, and says false when there is none or the
// read failed. A failed read closes the source; the next write probes again,
// so a gadget rebuilt underneath it is picked up.
func (w *hostReadWatch) pendingNoLock(devicePath string) (time.Duration, bool) {
	reader := w.readerNoLock(devicePath)
	if reader == nil {
		return 0, false
	}
	pending, err := reader.Pending()
	if err != nil {
		log.Debugf("%s: read pending_ms: %s", devicePath, err)
		w.forgetNoLock()
		return 0, false
	}
	return pending, true
}

// forgetNoLock drops the source, so the next write probes the kernel again.
// Call it when the device handle closes: the gadget may come back without the
// file, or with a new one.
func (w *hostReadWatch) forgetNoLock() {
	if w.reader != nil {
		_ = w.reader.Close()
	}
	w.reader = nil
	w.probed = false
	if w.timer != nil {
		w.timer.Stop()
	}
}

// armRecheckNoLock schedules a recheck of the endpoint after the given delay.
// The timer is made once per endpoint and reset by every write after that, so
// a write arms it without allocating.
func (h *Hid) armRecheckNoLock(device hidDevice, after time.Duration) {
	if after < hostReadRecheckMin {
		after = hostReadRecheckMin
	}
	if device.watch.timer == nil {
		device.watch.timer = newReadTimer(after, func() { h.recheckHostReading(device) })
		return
	}
	device.watch.timer.Reset(after)
}

// checkHostReadingNoLock runs after a write the kernel took. It fails the write
// with errHostNotReading when the report at the controller has waited
// hostReadStallAfter or more, and otherwise arms the recheck. Without the
// kernel's source it does nothing, and the write timeout stays the only
// signal.
func (h *Hid) checkHostReadingNoLock(device hidDevice) error {
	if device.watch == nil {
		return nil
	}
	pending, ok := device.watch.pendingNoLock(device.path)
	if !ok {
		return nil
	}
	if pending >= hostReadStallAfter {
		return errHostNotReading
	}
	h.armRecheckNoLock(device, hostReadStallAfter-pending)
	return nil
}

// recheckHostReading runs from the timer. A report still waiting at
// hostReadStallAfter is a host that stopped reading, recorded as a failed
// write would record it; one still waiting but younger is checked again when
// it would reach the limit. It writes nothing: a held key stays held, as it
// does on the vendor kernel when nothing more is typed, and the next write
// fails and releases it.
func (h *Hid) recheckHostReading(device hidDevice) {
	device.mu.Lock()
	defer device.mu.Unlock()

	if device.get() == nil {
		return
	}
	pending, ok := device.watch.pendingNoLock(device.path)
	if !ok || pending == 0 {
		return
	}
	if pending < hostReadStallAfter {
		h.armRecheckNoLock(device, hostReadStallAfter-pending)
		return
	}
	_ = device.note(errHostNotReading)
}
