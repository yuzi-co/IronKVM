package audio

// The capture format. arecord is started with these and the encoder is created
// with these, so they are one set of constants rather than two sets that have
// to agree.
const (
	SampleRate = 48000

	// Channels is 2 because the gadget captures stereo and this is what the
	// managed host sends. Downmixing would save about 2% of the core and throw
	// away half the information.
	Channels = 2

	// SamplesPerFrame is 20 ms per channel, which is the period arecord is
	// asked for and the frame size Opus is given.
	SamplesPerFrame = 960

	// Bitrate and Complexity were chosen by measurement on the device. At
	// complexity 3 the encoder needs 5.60% of the core; the 129-tap FIR and
	// G.711 path it replaced needed 4.66%. Complexity 5 costs 7.88% and buys
	// little at this bitrate. Bitrate barely moves the CPU cost at all: four
	// times the rate costs one fifth more. See tools/opusbench.
	Bitrate    = 96000
	Complexity = 3

	// maxPacketBytes bounds one encoded frame. 20 ms of stereo at 96 kbit/s is
	// about 240 bytes, and RFC 6716 caps a single-frame Opus packet at 1275
	// bytes total (not per channel), so this has room to spare.
	maxPacketBytes = 4000
)

// Encoder turns one 20 ms chunk of 48 kHz stereo S16_LE into one packet.
//
// It is an interface because the implementation is a static riscv64 archive:
// it links on the device and nowhere else. Every test of the pipeline supplies
// its own encoder, and only encoder_opus_test.go exercises the real one.
type Encoder interface {
	// Encode appends the packet to dst and returns the extended slice, which
	// is the convention the rest of this package uses. Passing dst[:0] reuses
	// the caller's buffer.
	//
	// pcm must be 2-byte aligned: the libopus implementation casts its first
	// byte to a *C.opus_int16, and an unaligned pointer is undefined behavior
	// in C. Every caller in this codebase gets pcm from make([]byte, ...) or
	// a slice of one, which the Go allocator always aligns well enough. A
	// future implementer or caller that slices pcm from an arbitrary byte
	// offset needs to preserve that alignment.
	Encode(pcm []byte, dst []byte) ([]byte, error)

	// Close releases the encoder. It is called once, from the goroutine that
	// owns the stream.
	Close()
}
