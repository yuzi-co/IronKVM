//go:build !novision

package audio

/*
	#cgo CFLAGS: -I${SRCDIR}/../../../include
	#cgo LDFLAGS: ${SRCDIR}/../../../dl_lib/libopus.a -lm

	#include <opus/opus.h>

	// opus_encoder_ctl is variadic, and cgo cannot call a variadic C function.
	// This wrapper is not, so cgo can. Every setting this package needs takes
	// one opus_int32, so one wrapper covers all of them.
	static int kvm_opus_set(OpusEncoder *encoder, int request, opus_int32 value) {
		return opus_encoder_ctl(encoder, request, value);
	}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type opusEncoder struct {
	state *C.OpusEncoder

	// scratch receives the packet from C. One encoder belongs to one stream
	// and Encode is called from one goroutine, so a single buffer is enough
	// and it keeps 50 allocations a second off the heap.
	scratch []byte
}

func newOpusEncoder() (Encoder, error) {
	var status C.int

	state := C.opus_encoder_create(
		C.opus_int32(SampleRate),
		C.int(Channels),
		C.OPUS_APPLICATION_AUDIO,
		&status,
	)
	if state == nil || status != C.OPUS_OK {
		return nil, fmt.Errorf("opus_encoder_create: %s", opusError(status))
	}

	encoder := &opusEncoder{state: state, scratch: make([]byte, maxPacketBytes)}

	// OPUS_APPLICATION_AUDIO rather than VOIP. This carries desktop audio, not
	// a phone call, and VOIP was measured more expensive as well as wrong:
	// mono at 32 kbit/s in VOIP mode costs more of the core than stereo at
	// 96 kbit/s in AUDIO mode, because SILK is not cheap.
	settings := []struct {
		name    string
		request C.int
		value   C.opus_int32
	}{
		{"bitrate", C.OPUS_SET_BITRATE_REQUEST, C.opus_int32(Bitrate)},
		{"complexity", C.OPUS_SET_COMPLEXITY_REQUEST, C.opus_int32(Complexity)},
	}

	for _, setting := range settings {
		if status := C.kvm_opus_set(state, setting.request, setting.value); status != C.OPUS_OK {
			C.opus_encoder_destroy(state)
			return nil, fmt.Errorf("opus_encoder_ctl(%s): %s", setting.name, opusError(status))
		}
	}

	return encoder, nil
}

// Encode hands one chunk to libopus.
//
// The length check is not defensive noise. This codec runs inside the server
// process, unlike arecord, which is a child that the source restarts when it
// dies. opus_encode reads SamplesPerFrame*Channels samples from the pointer it
// is given whatever the slice behind it holds, so a short chunk reads past the
// end and takes the whole server with it. Source.runOnce fills chunks with
// io.ReadFull and cannot produce a short one, but that guarantee lives in
// another file and this is where it has to hold.
func (e *opusEncoder) Encode(pcm []byte, dst []byte) ([]byte, error) {
	if len(pcm) != ChunkBytes {
		return dst, fmt.Errorf("audio: chunk is %d bytes, want %d", len(pcm), ChunkBytes)
	}

	// The samples are S16_LE and this is a little-endian machine, so the byte
	// slice is already an array of opus_int16. A slice from make is aligned
	// well enough for the cast.
	written := C.opus_encode(
		e.state,
		(*C.opus_int16)(unsafe.Pointer(&pcm[0])),
		C.int(SamplesPerFrame),
		(*C.uchar)(unsafe.Pointer(&e.scratch[0])),
		C.opus_int32(len(e.scratch)),
	)
	if written < 0 {
		return dst, fmt.Errorf("opus_encode: %s", opusError(C.int(written)))
	}

	return append(dst, e.scratch[:written]...), nil
}

func (e *opusEncoder) Close() {
	if e.state == nil {
		return
	}

	C.opus_encoder_destroy(e.state)
	e.state = nil
}

func opusError(status C.int) string {
	return C.GoString(C.opus_strerror(status))
}
