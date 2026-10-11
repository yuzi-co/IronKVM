package common

import "sync/atomic"

// videoOff is the failsafe mode of service/failsafe: the server runs without
// ever calling kvmv_init, so a libkvm that crashes the process cannot take the
// web UI, the update page and the logs down with it.
var videoOff atomic.Bool

// DisableVideo turns capture off for the life of the process. It must run
// before the first GetKvmVision, which is the call that would initialise
// libkvm; main.go calls it first thing.
//
// The capture gate is closed at once, so every call that would reach the
// library answers as it does after the teardown: no frame, no signal, no
// change. Close then has nothing to deinitialise, which is right, because
// nothing was initialised.
func DisableVideo() {
	videoOff.Store(true)
	captureLifecycle.disable()
}

// VideoDisabled reports whether DisableVideo ran.
func VideoDisabled() bool {
	return videoOff.Load()
}
