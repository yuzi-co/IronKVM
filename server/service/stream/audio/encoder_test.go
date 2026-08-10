package audio

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeEncoder stands in for libopus. The real encoder is a riscv64 archive and
// does not link on a workstation, so every test of the pipeline uses this.
type fakeEncoder struct {
	mutex  sync.Mutex
	chunks [][]byte
	packet []byte
	err    error
	closed bool
}

func (f *fakeEncoder) Encode(pcm []byte, dst []byte) ([]byte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.chunks = append(f.chunks, append([]byte(nil), pcm...))

	if f.err != nil {
		return dst, f.err
	}

	return append(dst, f.packet...), nil
}

func (f *fakeEncoder) Close() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.closed = true
}

func (f *fakeEncoder) count() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return len(f.chunks)
}

func (f *fakeEncoder) isClosed() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return f.closed
}

func (f *fakeEncoder) setError(err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.err = err
}

// fakeStream returns a stream ready to consume chunks, with no capture child
// behind it.
//
// It sets the encoder directly rather than calling Start. Start would launch
// the source, which runs arecord: absent on a workstation, so the retry loop
// would spin and log for the length of every test that only wants to exercise
// consume. The two tests that are about the lifecycle call Start themselves
// and supply an inert child.
func fakeStream(encoder *fakeEncoder) *Stream {
	stream := NewStream()
	stream.newEncoder = func() (Encoder, error) { return encoder, nil }
	stream.encoder = encoder

	return stream
}

func TestStreamDeliversWhatTheEncoderReturns(t *testing.T) {
	encoder := &fakeEncoder{packet: []byte("opus-packet")}
	stream := fakeStream(encoder)

	go stream.consume(make([]byte, ChunkBytes))

	select {
	case frame, ok := <-stream.Frames():
		if !ok {
			t.Fatal("the frame channel closed before delivering a frame")
		}

		if string(frame) != "opus-packet" {
			t.Errorf("the frame carried %q, want the encoder's packet", frame)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no frame arrived within 5s")
	}

	stream.Stop()
}

// A frame the encoder rejects is dropped. The stream stays open, because the
// next frame may encode: an encoder that fails once is not an encoder that is
// gone, and the listener has nowhere else to get audio from.
func TestStreamDropsARejectedFrameAndKeepsGoing(t *testing.T) {
	encoder := &fakeEncoder{packet: []byte("opus-packet")}
	encoder.setError(errors.New("synthetic encode failure"))

	stream := fakeStream(encoder)

	stream.consume(make([]byte, ChunkBytes))

	select {
	case frame, ok := <-stream.Frames():
		if !ok {
			t.Fatal("the frame channel closed because one frame failed to encode")
		}
		t.Fatalf("a rejected frame reached the channel as %q", frame)
	case <-time.After(100 * time.Millisecond):
	}

	// The next frame encodes, and it must arrive.
	encoder.setError(nil)
	go stream.consume(make([]byte, ChunkBytes))

	select {
	case frame, ok := <-stream.Frames():
		if !ok {
			t.Fatal("the frame channel closed instead of delivering the next frame")
		}
		if string(frame) != "opus-packet" {
			t.Errorf("the frame carried %q, want the encoder's packet", frame)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no frame arrived after the encoder recovered")
	}

	stream.Stop()
}

// An encoder that cannot be built is permanent. The stream has to end, or the
// consumer blocks on a channel nothing will close and the manager goes on
// believing audio is being sent.
func TestStreamClosesFramesWhenTheEncoderCannotBeBuilt(t *testing.T) {
	stream := NewStream()
	stream.newEncoder = func() (Encoder, error) {
		return nil, errors.New("synthetic construction failure")
	}

	stream.Start()

	done := make(chan struct{})
	go func() {
		for range stream.Frames() { //nolint:revive // draining is the point
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the frame channel stayed open after the encoder failed to build")
	}

	// Stop must still be safe, and must not block on a goroutine that never ran.
	stream.Stop()
}

func TestStreamGivesEveryChunkToTheEncoder(t *testing.T) {
	encoder := &fakeEncoder{packet: []byte("p")}
	stream := fakeStream(encoder)

	// The channel holds four, so three fit without a reader.
	for range 3 {
		stream.consume(make([]byte, ChunkBytes))
	}

	if got := encoder.count(); got != 3 {
		t.Errorf("the encoder saw %d chunks, want 3", got)
	}

	stream.Stop()
}
