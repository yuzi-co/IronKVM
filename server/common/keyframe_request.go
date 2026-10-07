package common

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// KeyframeRequestKind is how the loaded libkvm.so can be asked for a keyframe
// outside its GOP (ironkvm-dist#72).
type KeyframeRequestKind int

const (
	// KeyframeRequestNone: the encoder makes keyframes only at its GOP.
	KeyframeRequestNone KeyframeRequestKind = iota
	// KeyframeRequestV4L2: libkvm-v4l2 (slot B), through set_h264_gop with
	// the GOP unchanged, which sets the encoder's keyframe control.
	KeyframeRequestV4L2
	// KeyframeRequestVendor: Sipeed's library (slot A), through the vendor's
	// CVI_VENC_RequestIDR on libkvm's encoder channel.
	KeyframeRequestVendor
)

func (k KeyframeRequestKind) String() string {
	switch k {
	case KeyframeRequestV4L2:
		return "libkvm-v4l2"
	case KeyframeRequestVendor:
		return "CVI_VENC_RequestIDR"
	default:
		return "none"
	}
}

var keyframeFailureLogged sync.Map

// noteKeyframeRequestFailure logs a failed request once per error code. A
// stream whose request fails still gets the GOP's own keyframes.
func noteKeyframeRequestFailure(code int) {
	if _, seen := keyframeFailureLogged.LoadOrStore(code, struct{}{}); seen {
		return
	}

	switch code {
	case -2:
		log.Infof("keyframe request: no encoder channel carries video yet")
	default:
		log.Warnf("keyframe request: the encoder refused it (%#x); keyframes wait for the GOP", uint32(code))
	}
}
