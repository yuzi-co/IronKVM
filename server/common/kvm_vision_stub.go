//go:build novision

package common

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// captureLifecycle mirrors the real implementation so the state machine behaves
// the same way in a build with no hardware behind it.
var captureLifecycle = newCaptureGate()

// This stub replaces the cgo capture bindings when the "novision" build tag is
// set. The real implementation links against libkvm in dl_lib through cgo, so
// building it needs CGO_ENABLED=1 and the riscv64 cross-compiler. Without this
// stub, every package that reaches common is therefore out of reach of go vet
// and go test on a workstation. It is never used in a device build.

var (
	kvmVision     *KvmVision
	kvmVisionOnce sync.Once
)

type KvmVision struct{}

func GetKvmVision() *KvmVision {
	kvmVisionOnce.Do(func() {
		kvmVision = &KvmVision{}
		log.Debugf("kvm vision stub initialized")
	})

	return kvmVision
}

// One read tracker per encoding, mirroring the real implementation so that the
// call shape it uses is type-checked by a build that has no libkvm behind it.
var (
	mjpegReads captureReadLog
	h264Reads  captureReadLog
)

func (k *KvmVision) ReadMjpeg(width uint16, height uint16, quality uint16) (data []byte, result int) {
	result = -1
	reportCaptureRead(&mjpegReads, result)

	return nil, result
}

func (k *KvmVision) ReadH264(width uint16, height uint16, bitRate uint16) (data []byte, result int) {
	result = -1
	reportCaptureRead(&h264Reads, result)

	return nil, result
}

func (k *KvmVision) SetHDMI(enable bool) int {
	return 0
}

func (k *KvmVision) SetGop(gop uint8) {}

func (k *KvmVision) SetFrameDetect(frame uint8) {}

func (k *KvmVision) Close() {
	captureLifecycle.stop(func() {})
}

func (k *KvmVision) StopCapture() {
	captureLifecycle.stop(func() {})
}

func (k *KvmVision) ResumeCapture() {
	captureLifecycle.resume(func() {})
}

func (k *KvmVision) IsCapturing() bool {
	return captureLifecycle.isLive()
}
