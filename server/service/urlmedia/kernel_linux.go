//go:build linux

package urlmedia

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// The kernel's NBD ioctls, _IO(0xab, n) in include/uapi/linux/nbd.h.
const (
	nbdSetSock       = 0xab00
	nbdSetBlksize    = 0xab01
	nbdDoIt          = 0xab03
	nbdClearSock     = 0xab04
	nbdClearQue      = 0xab05
	nbdSetSizeBlocks = 0xab07
	nbdDisconnect    = 0xab08
	nbdSetTimeout    = 0xab09
	nbdSetFlags      = 0xab0a

	nbdFlagHasFlags = 1 << 0
	nbdFlagReadOnly = 1 << 1

	// sectorSize is the device's logical block size: a disk image keeps the
	// 512-byte sectors it was made with. The gadget serves a CD in 2048-byte
	// blocks whatever the device's are.
	sectorSize = 512
	// kernelTimeout is the kernel's own limit on a request, in seconds. It
	// is a backstop: the server answers EIO when its retries run out, well
	// before it, so a request never waits on a server that has gone.
	kernelTimeout = 60
	// maxSectorsKB keeps the kernel's requests inside one buffer of
	// serveNBD; a larger one would be streamed.
	maxSectorsKB = "512"
)

var (
	sysBlock = "/sys/block"
	// Two devices, one per drive, and no partition scan: the board has no
	// use for the image's partitions, and the scan would read it at attach.
	modprobe = func() error { return exec.Command("modprobe", "nbd", "nbds_max=2", "max_part=0").Run() }
)

// nbdIndices lists the nbd devices the kernel has, in order.
func nbdIndices() []int {
	entries, _ := filepath.Glob(filepath.Join(sysBlock, "nbd*"))
	var out []int
	for _, e := range entries {
		if n, err := strconv.Atoi(strings.TrimPrefix(filepath.Base(e), "nbd")); err == nil {
			out = append(out, n)
		}
	}
	sort.Ints(out)
	return out
}

var loadOnce sync.Mutex

// kernelNBDAvailable loads the nbd module if it is not loaded, and says
// whether the kernel has NBD devices.
func kernelNBDAvailable() error {
	loadOnce.Lock()
	defer loadOnce.Unlock()
	if len(nbdIndices()) > 0 {
		return nil
	}
	if err := modprobe(); err != nil {
		log.Debugf("urlmedia: modprobe nbd: %s", err)
	}
	if len(nbdIndices()) > 0 {
		return nil
	}
	return errNoNBD
}

// kernelDevice is an nbd device with the kernel's end of a socket pair. A
// goroutine holds NBD_DO_IT, which returns once the connection ends.
type kernelDevice struct {
	path string
	fd   int
	done chan struct{}
	once sync.Once
	// doItErr is what NBD_DO_IT returned, once done is closed.
	doItErr error
}

func (d *kernelDevice) Path() string { return d.path }

// Ended reports whether the connection has ended, which the kernel does by
// itself only when a request outlives its timeout.
func (d *kernelDevice) Ended() bool {
	select {
	case <-d.done:
		return true
	default:
		return false
	}
}

// Disconnect asks the kernel to end the connection, which sends the server
// NBD_CMD_DISC, and waits for NBD_DO_IT to return. The gadget must have let
// go of the device first.
func (d *kernelDevice) Disconnect() error {
	var err error
	d.once.Do(func() {
		if e := unix.IoctlSetInt(d.fd, nbdDisconnect, 0); e != nil {
			err = fmt.Errorf("disconnect %s: %w", d.path, e)
		}
		select {
		case <-d.done:
		case <-time.After(10 * time.Second):
			// The server side is closed by the caller next, which ends the
			// receive thread and so NBD_DO_IT.
			err = errors.Join(err, fmt.Errorf("%s did not disconnect in time", d.path))
		}
	})
	return err
}

// attachNBD connects a free nbd device to a new socket pair, read-only and
// size bytes long, and returns the device and the server's end.
func attachNBD(size int64) (device, net.Conn, error) {
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("socketpair: %w", err)
	}
	serverFile := os.NewFile(uintptr(fds[0]), "nbd-server")
	conn, err := net.FileConn(serverFile)
	_ = serverFile.Close()
	if err != nil {
		_ = unix.Close(fds[1])
		return nil, nil, err
	}
	kernelSock := fds[1]

	for _, n := range nbdIndices() {
		path := fmt.Sprintf("/dev/nbd%d", n)
		// A connected device has a pid; skip it without opening it.
		if _, err := os.Stat(filepath.Join(sysBlock, fmt.Sprintf("nbd%d", n), "pid")); err == nil {
			continue
		}
		fd, err := unix.Open(path, unix.O_RDWR|unix.O_CLOEXEC, 0)
		if err != nil {
			continue
		}
		if err := unix.IoctlSetInt(fd, nbdSetSock, kernelSock); err != nil {
			_ = unix.Close(fd)
			if errors.Is(err, unix.EBUSY) {
				continue
			}
			_ = unix.Close(kernelSock)
			_ = conn.Close()
			return nil, nil, fmt.Errorf("give %s its socket: %w", path, err)
		}
		if err := configure(fd, n, size); err != nil {
			_ = unix.IoctlSetInt(fd, nbdClearSock, 0)
			_ = unix.Close(fd)
			_ = unix.Close(kernelSock)
			_ = conn.Close()
			return nil, nil, fmt.Errorf("configure %s: %w", path, err)
		}
		// The kernel holds its own reference to the socket now.
		_ = unix.Close(kernelSock)

		d := &kernelDevice{path: path, fd: fd, done: make(chan struct{})}
		go d.run()
		if err := waitStarted(n, d.done); err != nil {
			_ = conn.Close()
			<-d.done
			return nil, nil, fmt.Errorf("start %s: %w", path, err)
		}
		return d, conn, nil
	}
	_ = unix.Close(kernelSock)
	_ = conn.Close()
	return nil, nil, errors.New("no free nbd device")
}

// waitStarted waits for NBD_DO_IT to start the device: until then it has
// no size, and the gadget would open an empty disk. The kernel adds the
// device's pid attribute as it starts, and sets the size last.
func waitStarted(n int, done chan struct{}) error {
	dir := filepath.Join(sysBlock, fmt.Sprintf("nbd%d", n))
	deadline := time.Now().Add(5 * time.Second)
	for {
		_, err := os.Stat(filepath.Join(dir, "pid"))
		size, _ := os.ReadFile(filepath.Join(dir, "size"))
		if err == nil && strings.TrimSpace(string(size)) != "0" && len(size) > 0 {
			return nil
		}
		select {
		case <-done:
			return errors.New("the kernel ended the connection at once")
		case <-time.After(10 * time.Millisecond):
		}
		if time.Now().After(deadline) {
			return errors.New("the device did not start")
		}
	}
}

func configure(fd, n int, size int64) error {
	if err := unix.IoctlSetInt(fd, nbdSetBlksize, sectorSize); err != nil {
		return err
	}
	if err := unix.IoctlSetInt(fd, nbdSetSizeBlocks, int(size/sectorSize)); err != nil {
		return err
	}
	if err := unix.IoctlSetInt(fd, nbdSetTimeout, kernelTimeout); err != nil {
		return err
	}
	if err := unix.IoctlSetInt(fd, nbdSetFlags, nbdFlagHasFlags|nbdFlagReadOnly); err != nil {
		return err
	}
	// Best effort: an older kernel may refuse the write.
	_ = os.WriteFile(filepath.Join(sysBlock, fmt.Sprintf("nbd%d", n), "queue", "max_sectors_kb"), []byte(maxSectorsKB), 0o644)
	return nil
}

// run holds NBD_DO_IT on a thread of its own with every signal blocked: the
// kernel waits interruptibly, and a signal there (a child's SIGCHLD, Go's
// preemption signal) would shut the device down under the host.
func (d *kernelDevice) run() {
	runtime.LockOSThread()
	// The thread is not unlocked: with its signal mask changed it is not
	// fit for other goroutines, and the runtime ends it when this returns.
	defer close(d.done)

	var all unix.Sigset_t
	for i := range all.Val {
		all.Val[i] = ^uint64(0)
	}
	if err := unix.PthreadSigmask(unix.SIG_BLOCK, &all, nil); err != nil {
		log.Warnf("urlmedia: block signals for %s: %s", d.path, err)
	}
	if err := unix.IoctlSetInt(d.fd, nbdDoIt, 0); err != nil {
		d.doItErr = err
		log.Infof("urlmedia: %s connection ended: %s", d.path, err)
	}
	_ = unix.IoctlSetInt(d.fd, nbdClearQue, 0)
	_ = unix.IoctlSetInt(d.fd, nbdClearSock, 0)
	_ = unix.Close(d.fd)
}
