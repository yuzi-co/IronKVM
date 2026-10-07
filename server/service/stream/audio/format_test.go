package audio

import (
	"slices"
	"testing"
)

// The host format is the one every existing path was built on, so making the
// source configurable must not move it.
func TestTheHostFormatIsTheGadget(t *testing.T) {
	if HostFormat.Device != CaptureDevice || HostFormat.Channels != Channels || HostFormat.Bitrate != Bitrate {
		t.Fatalf("host format %+v drifted from the package constants", HostFormat)
	}
	if HostFormat.ChunkBytes() != ChunkBytes {
		t.Fatalf("host chunk is %d bytes, want %d", HostFormat.ChunkBytes(), ChunkBytes)
	}
	if got := newArecord().Args; !slices.Equal(got, arecordFor(HostFormat).Args) {
		t.Fatalf("newArecord runs %v", got)
	}
}

// A second source reads its own device with its own channel count, and its
// chunks are sized to match.
func TestAnotherFormatReadsItsOwnDevice(t *testing.T) {
	format := Format{Device: "hw:sg2002onboard,1", Channels: 1, Bitrate: 32000}

	args := arecordFor(format).Args
	want := []string{"arecord", "-D", "hw:sg2002onboard,1", "-f", "S16_LE", "-r", "48000", "-c", "1", "-t", "raw", "--period-size=960"}
	if !slices.Equal(args, want) {
		t.Fatalf("arecord args %v, want %v", args, want)
	}

	if got := NewSourceFor(format).chunkBytes; got != 1920 {
		t.Fatalf("mono chunk is %d bytes, want 1920", got)
	}
}

// The filter sees each chunk before the encoder does, and what it changes is
// what gets encoded.
func TestTheFilterRunsBeforeTheEncoder(t *testing.T) {
	encoder := &fakeEncoder{packet: []byte{1}}
	stream := NewStreamFor(Format{Device: "x", Channels: 1, Bitrate: 32000}, func(chunk []byte) {
		for i := range chunk {
			chunk[i] = 7
		}
	})
	stream.encoder = encoder

	stream.consume(make([]byte, 1920))

	if encoder.count() != 1 {
		t.Fatalf("encoded %d chunks, want 1", encoder.count())
	}
	for _, b := range encoder.chunks[0] {
		if b != 7 {
			t.Fatal("the encoder saw the chunk before the filter changed it")
		}
	}
}
