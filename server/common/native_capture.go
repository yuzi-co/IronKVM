//go:build linux

package common

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// nativeCaptureEnv turns on libkvm-v4l2's native capture worker
// (ironkvm-dist#91). Off by default until it has been measured on a board.
//
// ReadVideo is a blocking cgo call into kvmv_read_video for each frame: 13 to
// 26 ms of scaler and encoder at 1080p on the mainline kernel. The runtime
// hands the P of a thread in a cgo call to another thread once sysmon notices,
// so every frame costs a P handoff and the thread switches that come with it.
// NanoKVM-OS measured 246 handoffs a second with this shape and 8 without it
// (their PR 65), and the server's CPU fell from about 21 to 13 %.
//
// With the worker the library makes that call on a thread of its own. The
// server writes the request into a block it shares with the worker, writes to
// one eventfd, and parks the capture goroutine on another in the netpoller
// until the worker answers. The frame is then copied out of the shared block
// in Go. When to read stays with the capture loop (read_pacer.go), so the
// stream asks for the same frames either way.
//
// Sipeed's library has no worker, and neither has a libkvm-v4l2 older than the
// variable: ReadVideo then makes the per-frame call as before.
const nativeCaptureEnv = "KVM_NATIVE_CAPTURE"

// nativeCaptureWanted reads the variable: on, 1, true or yes.
func nativeCaptureWanted(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "on", "1", "true", "yes":
		return true
	}

	return false
}

// nativeReadTimeout is how long a read waits for the worker. A read in the
// library gives up on its own lock after a second and on a frame after
// another, and a pipeline rebuild takes a few hundred milliseconds, so this is
// only reached by a worker that is stuck.
const nativeReadTimeout = 5 * time.Second

// noImage and retrievingImage mirror the library's -1, IMG_NOT_EXIST, and -5,
// "Retrieving image, please wait", which the stream loops already treat as no
// frame this time.
const (
	noImage         = -1
	retrievingImage = -5
)

// nativeRequest is one kvmv_read_video call's arguments.
type nativeRequest struct {
	width, height, bitrate uint16
	codec, gop, fps        uint8
}

// nativeAnswer is the worker's answer, with the frame already copied into Go
// memory.
type nativeAnswer struct {
	seq    uint32
	result int
	data   []byte

	// times are kvmv_video_times for the frame, CLOCK_MONOTONIC
	// microseconds, ntimes of them; mono and real are the two clocks in
	// nanoseconds, read together after them.
	times  [5]uint64
	ntimes int
	mono   int64
	real   int64
}

// videoTimes puts the answer's times on the wall clock, as
// kvmv_video_times_wall does in C for the per-frame path.
func (a nativeAnswer) videoTimes() (VideoTimes, bool) {
	if a.ntimes < len(a.times) || a.times[0] == 0 {
		return VideoTimes{}, false
	}

	wall := func(us uint64) time.Time {
		return time.Unix(0, a.real-(a.mono-int64(us)*1000))
	}

	return VideoTimes{
		CaptureEnd:  wall(a.times[0]),
		FrameTime:   time.Duration(a.times[1]) * time.Microsecond,
		EncodeStart: wall(a.times[2]),
		EncodeEnd:   wall(a.times[3]),
		Returned:    wall(a.times[4]),
	}, true
}

// nativePort is the block the server shares with the worker. The cgo file
// implements it over the library's struct kvmv_native_capture; the tests
// implement it over Go memory with a goroutine for the worker.
type nativePort interface {
	// put fills in the request and publishes seq last.
	put(req nativeRequest, seq uint32)
	// answered is the seq of the request the worker answered last.
	answered() uint32
	// take copies the answer out. Only valid once answered is the seq the
	// caller is waiting for, and only until the next put.
	take() nativeAnswer
}

// nativeCapture is the server's end of the worker. Only the capture loop
// reads through it, one read at a time.
type nativeCapture struct {
	port      nativePort
	requestFD int
	done      *os.File
	timeout   time.Duration

	seq     uint32
	pending bool // the worker has not answered seq yet
	last    nativeAnswer
	buf     [8]byte
}

// newNativeCapture takes the worker's two eventfds. done is dup'd: the
// library keeps its own, and the dup goes to the netpoller, which needs it to
// be nonblocking, as the library makes it.
func newNativeCapture(port nativePort, requestFD, doneFD int) (*nativeCapture, error) {
	if requestFD < 0 || doneFD < 0 {
		return nil, fmt.Errorf("no eventfds (%d, %d)", requestFD, doneFD)
	}

	fd, err := unix.FcntlInt(uintptr(doneFD), unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("dup the done eventfd: %w", err)
	}

	done := os.NewFile(uintptr(fd), "kvmv-native-capture")
	// A deadline is only possible on a descriptor in the poller, so this
	// also says that a wait will not block the thread.
	if err := done.SetReadDeadline(time.Time{}); err != nil {
		_ = done.Close()
		return nil, fmt.Errorf("the done eventfd is not pollable: %w", err)
	}

	return &nativeCapture{
		port:      port,
		requestFD: requestFD,
		done:      done,
		timeout:   nativeReadTimeout,
	}, nil
}

// errNativeKick is a request the worker could not be told about. The caller
// gives up on the worker for good.
var errNativeKick = errors.New("cannot signal the native capture worker")

// read asks the worker for one frame and waits for it. It answers the
// library's result, -5 when the worker has not answered in time, or
// errNativeKick.
func (n *nativeCapture) read(req nativeRequest) ([]byte, int, error) {
	// The answer to a request that timed out is not wanted any more: its
	// frame is older than the one asked for now. It must still arrive
	// before the next request, because the worker reads one at a time and
	// the frame slot it holds goes back only on the next request.
	if n.pending {
		if !n.wait() {
			return nil, retrievingImage, nil
		}
		n.pending = false
	}

	n.seq++
	if n.seq == 0 {
		n.seq = 1 // the worker starts out having answered 0
	}
	n.port.put(req, n.seq)

	if err := n.kick(); err != nil {
		return nil, noImage, fmt.Errorf("%w: %v", errNativeKick, err)
	}

	if !n.wait() {
		n.pending = true
		log.Warnf("native capture: no answer from the worker in %s", n.timeout)

		return nil, retrievingImage, nil
	}

	n.last = n.port.take()

	return n.last.data, n.last.result, nil
}

// lastTimes are the times of the last answer, for LastVideoTimes.
func (n *nativeCapture) lastTimes() (VideoTimes, bool) {
	return n.last.videoTimes()
}

// kick tells the worker a request is there. A raw write: an eventfd write
// does not block, so the runtime need not be told about it.
func (n *nativeCapture) kick() error {
	one := uint64(1)
	_, _, errno := unix.RawSyscall(unix.SYS_WRITE, uintptr(n.requestFD),
		uintptr(unsafe.Pointer(&one)), unsafe.Sizeof(one))
	if errno != 0 {
		return errno
	}

	return nil
}

// wait parks until the worker has answered seq, or the timeout. An eventfd
// read takes every signal sent since the last one, so a late answer to an
// abandoned request is read here too; answered tells which one this is.
func (n *nativeCapture) wait() bool {
	deadline := time.Now().Add(n.timeout)
	if err := n.done.SetReadDeadline(deadline); err != nil {
		return false
	}

	for n.port.answered() != n.seq {
		if _, err := n.done.Read(n.buf[:]); err != nil {
			if !errors.Is(err, os.ErrDeadlineExceeded) {
				log.Warnf("native capture: waiting for the worker: %v", err)
			}

			return n.port.answered() == n.seq
		}
	}

	return true
}

// close gives the dup back. The worker itself stops with kvmv_deinit.
func (n *nativeCapture) close() {
	_ = n.done.Close()
}

// nativeCaptureLibrary is what choosing the worker needs from the library,
// as two functions so the choice can be tested without one.
type nativeCaptureLibrary struct {
	// available says the library exports the worker.
	available func() bool
	// open starts it and answers its port and eventfds.
	open func() (port nativePort, requestFD, doneFD int, err error)
}

// chooseNativeCapture decides whether reads go through the worker: only when
// the variable asks for it, the library has it and it starts. Every other
// case answers nil, and ReadVideo makes the per-frame call.
func chooseNativeCapture(env string, lib nativeCaptureLibrary) *nativeCapture {
	if !nativeCaptureWanted(env) {
		return nil
	}

	if !lib.available() {
		log.Infof("native capture: %s=%s, but the loaded libkvm has no capture worker (Sipeed's library, or an older libkvm-v4l2): one call per frame",
			nativeCaptureEnv, env)

		return nil
	}

	port, requestFD, doneFD, err := lib.open()
	if err == nil {
		var capture *nativeCapture
		if capture, err = newNativeCapture(port, requestFD, doneFD); err == nil {
			log.Infof("native capture: frames come from libkvm's capture worker (%s)", nativeCaptureEnv)

			return capture
		}
	}

	log.Warnf("native capture: the worker did not start, one call per frame: %v", err)

	return nil
}
