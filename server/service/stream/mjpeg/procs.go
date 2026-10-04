package mjpeg

import (
	"os"
	"runtime"
)

// moreProcs gives the runtime a second P while an MJPEG stream runs on a board
// with one core, and returns the function that puts the old value back.
//
// Every picture is a cgo call into libkvm that mostly waits for hardware: on
// the mainline kernel about 10 ms for the scaler and the JPEG unit. A cgo call
// keeps its P until sysmon takes it back, so with one P the goroutine that
// encrypts and sends the previous picture often cannot run during that wait,
// and the core idles. Measured in ironkvm-dist trial 18, MJPEG over HTTPS at
// 1080p: on the mainline slot 21.7 to 22.7 fps with the core 83% busy, 23.3
// to 24.0 fps at 95%; on the vendor slot, whose reads are shorter, 26.7 to
// 27.1 fps, 27.4 to 28.9 fps.
//
// Only while the stream runs: with two Ps a woken M spins looking for work,
// which cost the H.264 stream about 6% of the core when it was set for the
// whole process. An explicit GOMAXPROCS in the environment is left alone.
//
// A variable so a test can count the calls without touching the runtime.
var moreProcs = func() func() {
	if runtime.NumCPU() != 1 || os.Getenv("GOMAXPROCS") != "" {
		return func() {}
	}

	previous := runtime.GOMAXPROCS(2)

	return func() { runtime.GOMAXPROCS(previous) }
}
