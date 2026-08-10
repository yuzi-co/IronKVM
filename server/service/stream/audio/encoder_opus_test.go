//go:build !novision

package audio

import (
	"fmt"
	"math"
	"runtime"
	"sync"
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

// TestOpusEncoderUnderGOMAXPROCSConcurrency forces opus_encode onto
// Go-created Ms rather than the runtime's first M.
//
// cgocall runs opus_encode on the calling M's g0 stack. This binary is linked
// with riscv64-unknown-linux-musl-gcc, and musl sizes a new thread's stack at
// 128 KB by default -- glibc's default is 8 MB. libopus here is a float build
// configured without --enable-alloca, so the CELT encoder uses C99 VLAs sized
// from the frame on whatever stack it is given. A VLA that overflows that
// stack is a SIGSEGV in this process, not a recoverable Go panic, and on this
// board a crashed server costs a reboot rather than a restarted child.
//
// Neither existing proof covers the Go-created-M case. tools/opusbench calls
// opus_encode from a static C binary's main(), which runs on the process's
// own large stack. The other tests in this file run sequentially and call
// newOpusEncoder directly, so they very likely stay on the runtime's first M
// -- whose g0 is also the large main stack, not a musl thread stack.
// Production runs the encoder on a goroutine that can land on any M the
// runtime creates.
//
// Raising GOMAXPROCS and running many independent encoders on many goroutines
// concurrently is what actually forces the runtime to create new Ms, each
// with its own musl-sized g0, and keeps enough of them busy encoding real
// frames that a marginal stack would show itself here rather than in
// production.
func TestOpusEncoderUnderGOMAXPROCSConcurrency(t *testing.T) {
	const goroutines = 8
	const encodesPerGoroutine = 500

	previous := runtime.GOMAXPROCS(goroutines * 2)
	t.Cleanup(func() { runtime.GOMAXPROCS(previous) })

	signal := tone()

	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for g := range goroutines {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			encoder, err := newOpusEncoder()
			if err != nil {
				errs <- fmt.Errorf("goroutine %d: failed to create the encoder: %w", id, err)
				return
			}
			defer encoder.Close()

			for i := range encodesPerGoroutine {
				packet, err := encoder.Encode(signal, nil)
				if err != nil {
					errs <- fmt.Errorf("goroutine %d, encode %d: %w", id, i, err)
					return
				}

				if len(packet) < 50 {
					errs <- fmt.Errorf("goroutine %d, encode %d: packet is %d bytes, too small to carry a tone",
						id, i, len(packet))
					return
				}

				if len(packet) > maxPacketBytes {
					errs <- fmt.Errorf("goroutine %d, encode %d: packet is %d bytes, over the %d byte buffer",
						id, i, len(packet), maxPacketBytes)
					return
				}
			}
		}(g)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
