// Package audio reads what the managed host plays into the UAC1 USB gadget and
// hands it out as 20 ms frames of Opus.
package audio

import (
	"bytes"
	"os"
	"os/exec"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// cardName is the id the UAC1 gadget registers under.
const cardName = "UAC1Gadget"

// cardsPath is a variable so tests can point it at a fixture.
var cardsPath = "/proc/asound/cards"

// stopTimeout bounds how long Stop waits for the capture goroutine.
//
// The wait is bounded in theory already, because SIGKILL always reaps a
// process. It is not bounded for a child wedged in an uninterruptible read.
// This board has a documented case of that exact shape: a read of
// /proc/cvitek/vb blocks forever in D state and survives every signal.
// Tearing the USB gadget down underneath a blocked snd_pcm_readi is the same
// class of event, and the settings switch can do it while a viewer listens.
//
// A goroutine left behind on a wedged child costs one stack, and it finishes
// by itself when the child finally dies. A Stop that never returns wedges the
// websocket handler that called it, so the bounded wait is the better trade.
const stopTimeout = 2 * time.Second

// warnCardsOnce keeps a failing read of the card list to one log line.
// Available runs once per WebRTC connection, so an unguarded log repeats with
// every viewer.
var warnCardsOnce sync.Once

// hasArecord is a function variable so tests can override it. In production,
// it checks once per process via sync.Once. Tests may assign a different
// function to control the arecord availability independently of the card check.
var hasArecord = func() func() bool {
	var (
		once    sync.Once
		present bool
	)

	return func() bool {
		once.Do(func() {
			if _, err := exec.LookPath("arecord"); err != nil {
				log.Warnf("audio is off: arecord is not installed: %s", err)
				return
			}

			present = true
		})

		return present
	}
}()

// Available reports whether audio can be captured right now.
//
// The card appears and disappears while the process runs, because the settings
// page rebuilds the USB gadget, so this must not be cached.
func Available() bool {
	if !hasArecord() {
		return false
	}

	cards, err := os.ReadFile(cardsPath)
	if err != nil {
		// A silent false here hides the whole feature. The file is always
		// present on this kernel, so a read that fails is a fault worth one
		// line in the log.
		warnCardsOnce.Do(func() {
			log.Warnf("audio is off: cannot read %s: %s", cardsPath, err)
		})

		return false
	}

	return bytes.Contains(cards, []byte(cardName))
}

// Stream turns the capture device into Opus packets.
type Stream struct {
	source *Source
	frames chan []byte

	// newEncoder is a field so a test can supply an encoder that needs no
	// libopus. The production value is the cgo wrapper, or the stub on a build
	// without the device libraries.
	newEncoder func() (Encoder, error)

	// encoder and packet are touched only by the source goroutine, between
	// Start launching it and Run returning.
	encoder Encoder
	packet  []byte

	// encodeFailures counts consecutive failed frames, so the log can fall
	// quiet when failure is the steady state.
	encodeFailures int

	mutex     sync.Mutex
	started   bool
	closeOnce sync.Once
	done      chan struct{}
}

func NewStream() *Stream {
	return &Stream{
		source:     NewSource(),
		newEncoder: newOpusEncoder,
		// Four frames of slack. A consumer further behind than 80 ms is not
		// going to catch up, and buffering only adds delay.
		frames: make(chan []byte, 4),
		packet: make([]byte, 0, maxPacketBytes),
		done:   make(chan struct{}),
	}
}

// Start begins capture. Frames arrive on Frames until Stop.
func (s *Stream) Start() {
	s.mutex.Lock()
	if s.started {
		s.mutex.Unlock()
		return
	}
	s.started = true
	s.mutex.Unlock()

	encoder, err := s.newEncoder()
	if err != nil {
		// This is permanent, and it is deliberately unlike the source's retry
		// loop. A host that plays nothing is idle and may start at any moment,
		// so capture retries it forever. An encoder that cannot be built will
		// not build itself later, so the stream ends and the manager learns
		// that audio stopped.
		//
		// Nothing was launched, so this path closes done itself. The goroutine
		// below is the only other closer and it never runs.
		log.Errorf("audio is off: %s", err)
		close(s.done)
		s.closeFrames()

		return
	}

	s.encoder = encoder

	go func() {
		defer close(s.done)
		defer s.encoder.Close()

		s.source.Run(s.consume)

		// Run returns when Stop killed the child. Closing here is what lets a
		// consumer that is draining Frames finish.
		s.closeFrames()
	}()
}

// closeFrames ends the channel exactly once, whichever path gets there first.
func (s *Stream) closeFrames() {
	s.closeOnce.Do(func() {
		close(s.frames)
	})
}

// Stop kills the child and closes Frames.
//
// Killing the child is what ends the read. While the host plays nothing,
// arecord blocks, so nothing in the read path notices that the last viewer has
// gone.
//
// Calling s.source.Stop() before reading started under the mutex is critical:
// it ensures that a Start call racing with Stop will spawn a goroutine whose
// Run returns immediately because the source is already stopped, so consume is
// never reached after the channel closes.
//
// The wait on done is what makes closing frames safe: it means the source
// goroutine, and therefore consume, has finished. A stream that was never
// started has no such goroutine to wait for.
//
// The wait is bounded by stopTimeout. On a timeout Stop returns without
// closing the channel, because consume may still be running and a send on a
// closed channel panics. Nothing is lost by that: the goroutine started by
// Start closes the channel itself once Run returns, so a consumer of a
// timed-out stream unblocks when the wedged child finally dies.
func (s *Stream) Stop() {
	s.source.Stop()

	s.mutex.Lock()
	started := s.started
	s.mutex.Unlock()

	if started {
		select {
		case <-s.done:
		case <-time.After(stopTimeout):
			log.Warnf("audio capture did not stop in %v; leaving it behind", stopTimeout)
			return
		}
	}

	s.closeFrames()
}

func (s *Stream) Frames() <-chan []byte {
	return s.frames
}

// consume encodes one capture chunk and offers the packet. It never blocks: a
// consumer that is behind loses 20 ms rather than stalling capture.
//
// A drop here is not the same as the per-client drop in Client.enqueueAudio,
// and it is the worse of the two. The client drop happens after packetization,
// so the receiver sees a sequence gap and a timestamp jump and treats it as
// loss, which its jitter buffer is built for. A drop here never reaches the
// packetizer, so the RTP timestamp does not advance across the gap: the
// receiver hears a stream with the silence cut out of it, time-compressed and
// drifting further from the host with every drop, and nothing in the protocol
// reports that it happened. This channel therefore has to keep up, and the
// only consumer is the send loop, which does not block.
func (s *Stream) consume(chunk []byte) {
	packet, err := s.encoder.Encode(chunk, s.packet[:0])
	if err != nil {
		s.reportEncodeFailure(err)
		return
	}

	// Keep the grown buffer for the next frame.
	s.packet = packet

	if s.encodeFailures >= quietAfterFailures {
		log.Infof("audio encode recovered after %d failed frames", s.encodeFailures)
	}
	s.encodeFailures = 0

	// The buffer is reused, so the channel gets a copy.
	frame := make([]byte, len(packet))
	copy(frame, packet)

	select {
	case s.frames <- frame:
	default:
	}
}

// reportEncodeFailure logs the first few failures and then goes quiet, the way
// Source does and for the same reason: the log is /tmp/nanokvm-server.log,
// which S99vidiag and the supervisor both read, and a frame arrives fifty
// times a second.
//
// A failed frame is dropped rather than fatal. One bad 20 ms is not a reason
// to end a stream that may encode the next one.
func (s *Stream) reportEncodeFailure(err error) {
	s.encodeFailures++

	switch {
	case s.encodeFailures < quietAfterFailures:
		log.Warnf("audio encode failed: %s (frame %d)", err, s.encodeFailures)
	case s.encodeFailures == quietAfterFailures:
		log.Warnf("audio encode has failed %d times, and the last reason was %s; "+
			"it stays quiet until a frame encodes again", s.encodeFailures, err)
	}
}
