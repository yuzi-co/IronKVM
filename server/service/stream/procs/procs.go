// Package procs lends the Go runtime a second P while a stream that waits in
// cgo runs on a board with one core.
//
// Both the MJPEG stream and the WebRTC video stream spend most of each frame
// in a cgo call into libkvm that waits for hardware: on the mainline kernel
// about 10 ms for the scaler and the JPEG unit, 13 to 17 ms for a 1080p H.264
// picture. A cgo call keeps its P until sysmon takes it back, and sysmon backs
// off to a 10 ms check when it has had nothing to do. With GOMAXPROCS 1 the
// goroutines that encrypt and send the previous frame then wait through most
// of that call with the core idle, and when the call returns the capture
// goroutine waits in turn behind them for the one P.
//
// Measured in ironkvm-dist trial 18 (MJPEG over HTTPS at 1080p: 21.7 to 22.7
// fps with the core 83% busy, 23.3 to 24.0 fps at 95%) and trial 56 (WebRTC
// H.264 at 1080p60 on the mainline slot: one viewer 55.8 to 57.9 fps, two 48
// to 50; with a second P 60.0 to 60.7 and 57.5 to 59.7).
//
// Only while such a stream runs: with two Ps a woken M spins looking for work,
// which cost the H.264 direct stream about 6% of the core when it was set for
// the whole process (trial 18). An explicit GOMAXPROCS in the environment, or
// a board with more than one core, is left alone.
package procs

import (
	"os"
	"runtime"
	"sync"
)

var (
	mutex    sync.Mutex
	holders  int
	previous int
)

// setMaxProcs and applies are variables so a test can count the changes
// without touching the runtime.
var (
	setMaxProcs = runtime.GOMAXPROCS
	applies     = func() bool {
		return runtime.NumCPU() == 1 && os.Getenv("GOMAXPROCS") == ""
	}
)

// Two raises GOMAXPROCS to 2 until the returned function runs.
//
// Holders are counted, so two streams that overlap share the one change: the
// first raises it, the last one out puts the old value back. The returned
// function is safe to call more than once.
func Two() func() {
	if !applies() {
		return func() {}
	}

	mutex.Lock()
	if holders == 0 {
		previous = setMaxProcs(2)
	}
	holders++
	mutex.Unlock()

	var once sync.Once

	return func() {
		once.Do(func() {
			mutex.Lock()
			defer mutex.Unlock()

			holders--
			if holders == 0 {
				setMaxProcs(previous)
			}
		})
	}
}
