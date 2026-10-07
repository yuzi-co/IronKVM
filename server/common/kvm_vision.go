//go:build !novision

package common

/*
	#cgo CFLAGS: -I../include
	#cgo LDFLAGS: -L../dl_lib -lkvm
	#include <dlfcn.h>
	#include "kvm_vision.h"

	#ifndef RTLD_DEFAULT
	#define RTLD_DEFAULT ((void *)0)
	#endif

	// kvmv_codec_supported is an extension of libkvm-v4l2 and absent from
	// Sipeed's library, so it is found at run time rather than linked: a
	// weak reference would be resolved, or not, by whichever libkvm.so the
	// server was linked against, not by the one it runs with.
	// Answers -1 when the library does not have it.
	static int kvmv_codec_supported_lookup(uint8_t codec)
	{
		uint8_t (*fn)(uint8_t) =
			(uint8_t (*)(uint8_t))dlsym(RTLD_DEFAULT, "kvmv_codec_supported");

		if (fn == NULL)
			return -1;
		return fn(codec);
	}

	// kvmv_set_keep_aspect is another libkvm-v4l2 extension (ironkvm-dist#37):
	// 1 keeps the source's shape at the requested height, 0 stretches it to
	// the requested size. Answers -1 when the library does not have it, as
	// Sipeed's does not; that library always stretches.
	static int kvmv_set_keep_aspect_lookup(uint8_t keep)
	{
		uint8_t (*fn)(uint8_t) =
			(uint8_t (*)(uint8_t))dlsym(RTLD_DEFAULT, "kvmv_set_keep_aspect");

		if (fn == NULL)
			return -1;
		return fn(keep);
	}

	static int kvmv_keep_aspect_available(void)
	{
		return dlsym(RTLD_DEFAULT, "kvmv_set_keep_aspect") != NULL;
	}

	// Keyframe requests (ironkvm-dist#72, trial 64). kvm_vision.h declares
	// none. libkvm-v4l2 asks its encoder for a keyframe on every
	// set_h264_gop, without a rebuild while the GOP stays the same, so the
	// server uses that there. Sipeed's library rebuilds its encoder on
	// set_h264_gop instead, so on it the request goes to the vendor's
	// CVI_VENC_RequestIDR in libvenc.so, which libkvm already loads, on the
	// channel that carries H.264 or H.265. Both are found at run time.
	static int kvmv_is_v4l2_library(void)
	{
		return dlsym(RTLD_DEFAULT, "kvmv_codec_supported") != NULL;
	}

	typedef int (*cvi_request_idr_fn)(int, int);
	typedef int (*cvi_get_chn_attr_fn)(int, void *);

	static int vendor_venc_chn = -1;

	// VENC_CHN_ATTR_S opens with VENC_ATTR_S, which opens with the payload
	// type: PT_H264 is 96, PT_H265 265. Nothing else is read. The buffer is
	// far larger than the structure, and only the capture loop calls this.
	static int vendor_venc_is_video(cvi_get_chn_attr_fn get_attr, int chn)
	{
		static union {
			int type;
			unsigned char raw[8192];
		} attr;

		memset(&attr, 0, sizeof(attr));
		if (get_attr(chn, &attr) != 0)
			return 0;
		return attr.type == 96 || attr.type == 265;
	}

	static int kvmv_vendor_keyframe_available(void)
	{
		return dlsym(RTLD_DEFAULT, "CVI_VENC_RequestIDR") != NULL &&
		       dlsym(RTLD_DEFAULT, "CVI_VENC_GetChnAttr") != NULL;
	}

	// Answers 0 when the encoder took the request, -1 without the vendor
	// calls, -2 when no channel carries video (none built yet), else the
	// vendor's error.
	static int kvmv_vendor_request_keyframe(void)
	{
		cvi_request_idr_fn request =
			(cvi_request_idr_fn)dlsym(RTLD_DEFAULT, "CVI_VENC_RequestIDR");
		cvi_get_chn_attr_fn get_attr =
			(cvi_get_chn_attr_fn)dlsym(RTLD_DEFAULT, "CVI_VENC_GetChnAttr");
		int chn;
		int ret;

		if (request == NULL || get_attr == NULL)
			return -1;
		if (vendor_venc_chn < 0 || !vendor_venc_is_video(get_attr, vendor_venc_chn)) {
			vendor_venc_chn = -1;
			for (chn = 0; chn < 16; chn++) {
				if (vendor_venc_is_video(get_attr, chn)) {
					vendor_venc_chn = chn;
					break;
				}
			}
		}
		if (vendor_venc_chn < 0)
			return -2;
		// bInstant: the next picture, not the next GOP.
		ret = request(vendor_venc_chn, 1);
		if (ret != 0)
			vendor_venc_chn = -1;
		return ret;
	}
*/
import "C"
import (
	"sync"
	"unsafe"

	log "github.com/sirupsen/logrus"
)

var (
	kvmVision     *KvmVision
	kvmVisionOnce sync.Once

	// captureLifecycle keeps frame reads out of the way of teardown. See
	// capture_gate.go for why libkvm cannot be trusted to do this itself.
	captureLifecycle = newCaptureGate()
)

// imgNotExist mirrors IMG_NOT_EXIST from kvm_vision.h.
const imgNotExist = -1

// One read tracker per encoding, because the two are read independently and a
// shared tracker would report a change every time they disagreed. service/hid
// keeps one health record per endpoint for the same reason.
var (
	mjpegReads captureReadLog
	h264Reads  captureReadLog
)

// KvmVision carries no state. The lifecycle lives in captureLifecycle, which
// holds the read lock across each call rather than exposing a flag to test
// beforehand - see capture_gate.go for why a flag would not close the race.
type KvmVision struct{}

func GetKvmVision() *KvmVision {
	kvmVisionOnce.Do(func() {
		kvmVision = &KvmVision{}

		logLevel := C.uint8_t(0)
		C.kvmv_init(logLevel)
		log.Debugf("kvm vision initialized")
	})

	return kvmVision
}

func (k *KvmVision) ReadMjpeg(width uint16, height uint16, quality uint16) (data []byte, result int) {
	var (
		kvmData  *C.uint8_t
		dataSize C.uint32_t
	)

	// A read after the teardown answers IMG_NOT_EXIST rather than reaching
	// kvmv_read_img with the mutex already destroyed. The streamers treat that
	// as "no frame this time", which is what they do on a live board whenever
	// libkvm has nothing ready.
	if !captureLifecycle.withLive(func() {
		result = int(C.kvmv_read_img(
			C.uint16_t(width),
			C.uint16_t(height),
			C.uint8_t(0),
			C.uint16_t(quality),
			&kvmData,
			&dataSize,
		))

		reportCaptureRead(&mjpegReads, result)
		if result < 0 {
			return
		}
		defer C.free_kvmv_data(&kvmData)

		data = C.GoBytes(unsafe.Pointer(kvmData), C.int(dataSize))
	}) {
		return nil, imgNotExist
	}

	return
}

func (k *KvmVision) ReadH264(width uint16, height uint16, bitRate uint16) (data []byte, result int) {
	var (
		kvmData  *C.uint8_t
		dataSize C.uint32_t
	)

	// A read after the teardown answers IMG_NOT_EXIST rather than reaching
	// kvmv_read_img with the mutex already destroyed. The streamers treat that
	// as "no frame this time", which is what they do on a live board whenever
	// libkvm has nothing ready.
	if !captureLifecycle.withLive(func() {
		result = int(C.kvmv_read_img(
			C.uint16_t(width),
			C.uint16_t(height),
			C.uint8_t(1),
			C.uint16_t(bitRate),
			&kvmData,
			&dataSize,
		))

		reportCaptureRead(&h264Reads, result)
		if result < 0 {
			return
		}
		defer C.free_kvmv_data(&kvmData)

		data = C.GoBytes(unsafe.Pointer(kvmData), C.int(dataSize))
	}) {
		return nil, imgNotExist
	}

	return
}

// ReadVideo reads one encoded access unit in the codec asked for.
//
// ReadH264 above goes through kvmv_read_img, which can only ever ask for
// H.264. This goes through kvmv_read_video, which takes the codec, the gop and
// the frame rate, so the encoder is configured by the caller rather than by
// whatever the module last remembered.
//
// The codec numbering here is libkvm's public one, CodecH264 and CodecH265.
// libkvm converts to mmf's, which runs the other way.
func (k *KvmVision) ReadVideo(width uint16, height uint16, codec uint8, bitRate uint16, gop uint8, fps uint8) (data []byte, result int) {
	var (
		kvmData  *C.uint8_t
		dataSize C.uint32_t
	)

	if !captureLifecycle.withLive(func() {
		result = int(C.kvmv_read_video(
			C.uint16_t(width),
			C.uint16_t(height),
			C.uint8_t(codec),
			C.uint16_t(bitRate),
			C.uint8_t(gop),
			C.uint8_t(fps),
			&kvmData,
			&dataSize,
		))

		reportCaptureRead(&h264Reads, result)
		if result < 0 {
			return
		}
		defer C.free_kvmv_data(&kvmData)

		data = C.GoBytes(unsafe.Pointer(kvmData), C.int(dataSize))
	}) {
		return nil, imgNotExist
	}

	return
}

func (k *KvmVision) SetHDMI(enable bool) int {
	hdmiEnable := C.uint8_t(0)
	if enable {
		hdmiEnable = C.uint8_t(1)
	}

	result := -1
	if !captureLifecycle.withLive(func() {
		result = hdmiControlResult(int(C.kvmv_hdmi_control(hdmiEnable)))
	}) {
		return -1
	}

	if result < 0 {
		log.Errorf("failed to set hdmi to %t: the library declined, which is what alpha and beta boards always do", enable)
	}

	return result
}

func (k *KvmVision) HasHDMISignal() bool {
	active := false
	captureLifecycle.withLive(func() {
		active = C.kvmv_hdmi_signal_active() != 0
	})

	return active
}

func (k *KvmVision) SetGop(gop uint8) {
	_gop := C.uint8_t(gop)
	captureLifecycle.withLive(func() {
		C.set_h264_gop(_gop)
	})
}

// KeyframeRequests says how the loaded library can be asked for a keyframe
// outside its GOP, or that it cannot. It needs no kvmv_init.
func (k *KvmVision) KeyframeRequests() KeyframeRequestKind {
	if C.kvmv_is_v4l2_library() != 0 {
		return KeyframeRequestV4L2
	}
	if C.kvmv_vendor_keyframe_available() != 0 {
		return KeyframeRequestVendor
	}

	return KeyframeRequestNone
}

// RequestKeyframe asks the encoder to make its next picture a keyframe. gop
// must be the GOP the encoder holds: libkvm-v4l2 takes the request through
// set_h264_gop, and another value there would also change the GOP. It answers
// whether the request was made. Only the capture loop calls it, between reads.
func (k *KvmVision) RequestKeyframe(gop uint8) bool {
	kind := k.KeyframeRequests()
	made := false

	captureLifecycle.withLive(func() {
		switch kind {
		case KeyframeRequestV4L2:
			C.set_h264_gop(C.uint8_t(gop))
			made = true
		case KeyframeRequestVendor:
			ret := int(C.kvmv_vendor_request_keyframe())
			made = ret == 0
			if !made {
				noteKeyframeRequestFailure(ret)
			}
		}
	})

	return made
}

// SetFPS tells the encoder what frame rate the capture loop is feeding it, so
// its rate controller can work out what one frame may cost. It answers false
// when the library is older than the call, which is possible because a library
// and a server binary are deployed one file at a time here.
func (k *KvmVision) SetFPS(fps uint8) bool {
	_fps := C.uint8_t(fps)
	available := false
	captureLifecycle.withLive(func() {
		available = C.set_h264_fps_if_available(_fps) != 0
	})

	return available
}

// SetCaptureFPS tells the capture channel how many frames a second the stream
// loop is going to take from it. The channel came up handing out every frame
// the source produced, which on a 60Hz source is twice what a default stream
// reads, and every frame nobody takes is still written to memory.
//
// It answers false when the library is older than the call, in which case the
// channel keeps handing out everything, which is what it did before this
// existed.
func (k *KvmVision) SetCaptureFPS(fps uint8) bool {
	_fps := C.uint8_t(fps)
	available := false
	captureLifecycle.withLive(func() {
		available = C.set_capture_fps_if_available(_fps) != 0
	})

	return available
}

// libraryCodecSupported asks the loaded libkvm.so. It needs no kvmv_init, so
// the screen settings can be checked before capture has started.
func libraryCodecSupported(codec uint8) bool {
	switch C.kvmv_codec_supported_lookup(C.uint8_t(codec)) {
	case -1:
		return vendorCodecSupported(codec)
	case 0:
		return false
	default:
		return true
	}
}

// libraryAspectSupported reports whether the loaded libkvm.so can keep the
// source's aspect ratio (libkvm-v4l2's kvmv_set_keep_aspect). Like
// libraryCodecSupported it needs no kvmv_init.
func libraryAspectSupported() bool {
	return C.kvmv_keep_aspect_available() != 0
}

// librarySetKeepAspect tells the library how to fit a source of another shape.
// It answers false when the library has no such choice.
func librarySetKeepAspect(keep bool) bool {
	value := C.uint8_t(0)
	if keep {
		value = 1
	}

	return C.kvmv_set_keep_aspect_lookup(value) >= 0
}

func (k *KvmVision) SetFrameDetect(frame uint8) {
	_frame := C.uint8_t(frame)
	captureLifecycle.withLive(func() {
		C.set_frame_detact(_frame)
	})
}

func (k *KvmVision) Close() {
	captureLifecycle.stop(func() {
		C.kvmv_deinit()
	})
	log.Debugf("stop kvm vision...")
}

// There used to be StopCapture, ResumeCapture and IsCapturing here. Their only
// caller was service/vm/hdmi_idle.go, which stopped capture after an idle
// timeout, and the 2026-08-01 rebase dropped that file - so the three methods
// survived their consumer by months with nothing calling them.
//
// The gate itself stays, and it is what keeps a read from reaching a destroyed
// vi_mutex while Close is running. What it no longer carries is a way back:
// see the note at the end of capture_gate.go for why nothing on a device can
// stop capture without also ending the process.
