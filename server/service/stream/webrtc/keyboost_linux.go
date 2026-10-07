package webrtc

import (
	"runtime"

	"golang.org/x/sys/unix"
)

// keyBoost raises the calling writer's thread to keyNice while it writes a
// keyframe and puts it back after. Linux applies PRIO_PROCESS with a thread id
// to that thread alone, so the goroutine is locked to its thread for its whole
// life; when it returns, still locked, Go ends the thread rather than reuse it.
type keyBoost struct {
	on   bool
	tid  int
	nice int
	base int
}

// setThreadNice and threadNice are variables so a test can refuse them.
var (
	setThreadNice = func(tid int, nice int) error {
		return unix.Setpriority(unix.PRIO_PROCESS, tid, nice)
	}

	threadNice = func(tid int) (int, error) {
		// The raw system call answers 20 - nice.
		prio, err := unix.Getpriority(unix.PRIO_PROCESS, tid)

		return 20 - prio, err
	}
)

// newKeyBoost must run on the writer goroutine.
func newKeyBoost() *keyBoost {
	nice := keyNice()
	if nice == 0 || keyBoostOff.Load() {
		return &keyBoost{}
	}

	runtime.LockOSThread()

	tid := unix.Gettid()
	base, err := threadNice(tid)
	if err != nil {
		base = 0
	}

	if nice >= base {
		return &keyBoost{}
	}

	return &keyBoost{on: true, tid: tid, nice: nice, base: base}
}

func (b *keyBoost) raise() {
	if !b.on || keyBoostOff.Load() {
		return
	}

	if err := setThreadNice(b.tid, b.nice); err != nil {
		b.on = false
		keyBoostFailed(b.nice, err)
	}
}

func (b *keyBoost) lower() {
	if !b.on {
		return
	}

	_ = setThreadNice(b.tid, b.base)
}
