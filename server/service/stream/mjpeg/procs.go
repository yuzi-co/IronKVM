package mjpeg

import "NanoKVM-Server/service/stream/procs"

// moreProcs gives the runtime a second P while an MJPEG stream runs on a board
// with one core, and returns the function that puts the old value back.
//
// Every picture is a cgo call into libkvm that mostly waits for hardware: on
// the mainline kernel about 10 ms for the scaler and the JPEG unit. With one
// P the goroutine that encrypts and sends the previous picture often cannot
// run during that wait, and the core idles. See procs.Two, which the WebRTC
// video stream shares, so that the two never undo each other's change.
//
// A variable so a test can count the calls without touching the runtime.
var moreProcs = procs.Two
