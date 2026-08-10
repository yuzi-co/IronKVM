//go:build !novision

package audio

import (
	"math"
	"testing"
)

// These run on the device. The archive is riscv64, so the encoder does not
// link anywhere else. Build with `go test -c` in the builder image and run the
// binary on the board; the command is in the plan and in CLAUDE.md.

// tone fills one chunk with a 440 Hz sine at 48 kHz, stereo, S16_LE. Silence
// would encode to a handful of bytes and would not show that the encoder is
// carrying a signal.
func tone() []byte {
	chunk := make([]byte, ChunkBytes)

	for i := range SamplesPerFrame {
		value := int16(20000 * math.Sin(2*math.Pi*440*float64(i)/SampleRate))

		for c := range Channels {
			offset := (i*Channels + c) * 2
			chunk[offset] = byte(value)
			chunk[offset+1] = byte(value >> 8)
		}
	}

	return chunk
}

func TestOpusEncoderEncodesAChunk(t *testing.T) {
	encoder, err := newOpusEncoder()
	if err != nil {
		t.Fatalf("failed to create the encoder: %s", err)
	}
	t.Cleanup(encoder.Close)

	packet, err := encoder.Encode(tone(), nil)
	if err != nil {
		t.Fatalf("failed to encode: %s", err)
	}

	// 20 ms at 96 kbit/s is about 240 bytes. A packet of two or three bytes is
	// what Opus emits for silence, and it would mean the samples never arrived.
	if len(packet) < 50 {
		t.Errorf("the packet is %d bytes, which is too small to carry a tone", len(packet))
	}

	if len(packet) > maxPacketBytes {
		t.Errorf("the packet is %d bytes, over the %d byte buffer", len(packet), maxPacketBytes)
	}
}

func TestOpusEncoderAppendsToDestination(t *testing.T) {
	encoder, err := newOpusEncoder()
	if err != nil {
		t.Fatalf("failed to create the encoder: %s", err)
	}
	t.Cleanup(encoder.Close)

	dst := []byte{0xAA, 0xBB}

	packet, err := encoder.Encode(tone(), dst)
	if err != nil {
		t.Fatalf("failed to encode: %s", err)
	}

	if len(packet) <= 2 || packet[0] != 0xAA || packet[1] != 0xBB {
		t.Errorf("Encode did not append to the destination it was given")
	}
}

// The chunk length is checked on the Go side because the C call reads
// SamplesPerFrame*Channels samples whatever the slice actually holds. A short
// chunk would read past the end, and this codec runs in the server process.
func TestOpusEncoderRejectsAShortChunk(t *testing.T) {
	encoder, err := newOpusEncoder()
	if err != nil {
		t.Fatalf("failed to create the encoder: %s", err)
	}
	t.Cleanup(encoder.Close)

	if _, err := encoder.Encode(make([]byte, ChunkBytes-2), nil); err == nil {
		t.Error("Encode accepted a short chunk")
	}
}

func TestOpusEncoderRejectsAnEmptyChunk(t *testing.T) {
	encoder, err := newOpusEncoder()
	if err != nil {
		t.Fatalf("failed to create the encoder: %s", err)
	}
	t.Cleanup(encoder.Close)

	if _, err := encoder.Encode(nil, nil); err == nil {
		t.Error("Encode accepted an empty chunk")
	}
}

// Encode is exported, so a caller can reach it after Close. Without a guard
// that hands opus_encode a NULL state pointer, which is a SIGSEGV inside the
// codec running in the server process.
func TestOpusEncoderEncodeAfterCloseReturnsError(t *testing.T) {
	encoder, err := newOpusEncoder()
	if err != nil {
		t.Fatalf("failed to create the encoder: %s", err)
	}

	encoder.Close()

	if _, err := encoder.Encode(tone(), nil); err == nil {
		t.Error("Encode after Close did not return an error")
	}
}

// Close is called from a defer in the pipeline this encoder feeds, so a
// second call reaching it (teardown racing an earlier error path, for
// example) must be a no-op rather than a double free.
func TestOpusEncoderCloseTwiceIsSafe(t *testing.T) {
	encoder, err := newOpusEncoder()
	if err != nil {
		t.Fatalf("failed to create the encoder: %s", err)
	}

	encoder.Close()
	encoder.Close()
}

// The encoder is reused across frames fifty times a second in production;
// nothing above exercises more than a single call, so this checks the second
// call succeeds with the state the first call left behind.
func TestOpusEncoderEncodesTwoChunksInSequence(t *testing.T) {
	encoder, err := newOpusEncoder()
	if err != nil {
		t.Fatalf("failed to create the encoder: %s", err)
	}
	t.Cleanup(encoder.Close)

	first, err := encoder.Encode(tone(), nil)
	if err != nil {
		t.Fatalf("failed to encode the first chunk: %s", err)
	}
	if len(first) < 50 {
		t.Errorf("the first packet is %d bytes, which is too small to carry a tone", len(first))
	}

	second, err := encoder.Encode(tone(), nil)
	if err != nil {
		t.Fatalf("failed to encode the second chunk: %s", err)
	}
	if len(second) < 50 {
		t.Errorf("the second packet is %d bytes, which is too small to carry a tone", len(second))
	}
}
