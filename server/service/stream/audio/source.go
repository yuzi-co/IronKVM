package audio

import (
	"bytes"
	"io"
	"os/exec"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// CaptureDevice names the card rather than its index. The index depends on
	// probe order and moves when the gadget is rebuilt; the name does not.
	CaptureDevice = "hw:UAC1Gadget,0"

	// ChunkBytes is 20 ms of 48 kHz stereo S16_LE. It is derived rather than
	// written out, because arecord and the encoder have to agree on it.
	ChunkBytes = SamplesPerFrame * Channels * 2

	// stderrDrainTimeout bounds the wait for the child's last stderr line
	// before we reap it. Short: the line is already written by then in every
	// case that matters, and this sits on the teardown path.
	stderrDrainTimeout = 200 * time.Millisecond

	// minRunDuration is the minimum time a child must run to count as healthy.
	// A child that emits a single chunk and exits is not a healthy restart; it
	// is a crash loop.
	minRunDuration = 5 * time.Second

	// quietAfterFailures is how many consecutive failures are worth a line
	// each. The source says so once more at that count and then writes nothing
	// until it recovers.
	//
	// Capture is never retired, so without this a source that cannot work
	// writes a line per attempt for as long as a viewer listens. A managed
	// host that plays nothing streams nothing, and arecord fails every time,
	// so that is the ordinary case rather than a rare one. The log it fills is
	// /tmp/nanokvm-server.log, which S99vidiag and the supervisor both read.
	quietAfterFailures = 5

	// stderrLimit caps what is kept of the child's stderr. ALSA states its
	// reason in one short line, so this is generous; the cap is there because
	// a child that complains once per period must not grow a buffer forever.
	stderrLimit = 512
)

// tailBuffer keeps the last stderrLimit bytes written to it and nothing else.
//
// arecord reports why it failed on stderr, and the reason is the whole
// diagnosis: "Device or resource busy" means another reader holds the card,
// "No such file or directory" means the gadget has no card, and
// "Sample format non available" means the gadget offers something else. With
// Stderr left unset the child writes to /dev/null and all of that is lost.
type tailBuffer struct {
	mutex sync.Mutex
	data  []byte
}

func (b *tailBuffer) Write(p []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.data = append(b.data, p...)
	if len(b.data) > stderrLimit {
		b.data = b.data[len(b.data)-stderrLimit:]
	}

	return len(p), nil
}

// lastLine returns the last line that has something in it. arecord prints its
// usage banner and its error on separate lines, and the error comes last.
func (b *tailBuffer) lastLine() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	lines := bytes.Split(b.data, []byte("\n"))
	for i := len(lines) - 1; i >= 0; i-- {
		if line := bytes.TrimSpace(lines[i]); len(line) > 0 {
			return string(line)
		}
	}

	return ""
}

// Source owns the arecord child process and hands its output out in chunks.
type Source struct {
	// newCmd builds the child. It is a field so that tests can supply a
	// command which does not need ALSA.
	newCmd func() *exec.Cmd

	minBackoff time.Duration
	maxBackoff time.Duration

	mutex   sync.Mutex
	cmd     *exec.Cmd
	stdout  io.ReadCloser
	stopped bool
	done    chan struct{}
}

func NewSource() *Source {
	return &Source{
		newCmd:     newArecord,
		minBackoff: 200 * time.Millisecond,
		maxBackoff: 5 * time.Second,
		done:       make(chan struct{}),
	}
}

// newArecord reads the gadget and writes raw samples to stdout.
//
// The period is pinned to 960 frames, which is 20 ms and the same size as a
// chunk. Left to ALSA, the period comes from the driver's default and decides
// how much delay sits in front of the encoder. The gadget advertises
// PERIOD_SIZE [32 1024], so 960 is inside its range.
func newArecord() *exec.Cmd {
	return exec.Command("arecord",
		"-D", CaptureDevice,
		"-f", "S16_LE",
		"-r", "48000",
		"-c", "2",
		"-t", "raw",
		"--period-size=960",
	)
}

// Run calls handle with each full chunk until Stop is called. The slice handed
// to handle is reused on the next read and is only valid until handle returns.
// Run blocks.
//
// A failing child is retried for as long as the caller listens, and capture is
// never retired. A host that streams no audio is the ordinary idle state of
// this device rather than a fault, and it becomes audio again the moment the
// host plays something. Retiring capture would make that moment produce
// nothing: StartAudioStream has one caller, and it fires on an ICE state
// change that a settled connection never produces again.
//
// What the retry must not do is cost anything while it waits. The backoff
// climbs to maxBackoff, and the log falls quiet after quietAfterFailures.
func (s *Source) Run(handle func([]byte)) {
	chunk := make([]byte, ChunkBytes)
	backoff := s.minBackoff

	var failures int

	for {
		if s.isStopped() {
			return
		}

		delivered, uptime := s.runOnce(chunk, handle)

		if delivered && uptime > minRunDuration {
			// The child produced audio and ran long enough, so the next failure
			// is a fresh one.
			if failures >= quietAfterFailures {
				log.Infof("audio capture recovered after %d failed attempts", failures)
			}

			backoff = s.minBackoff
			failures = 0
		} else {
			// This branch is the negation of the condition above, so every
			// arrival here is a child that failed or one that did not run
			// long enough to count.
			failures++

			switch {
			case failures < quietAfterFailures:
				log.Warnf("audio capture exited (uptime=%v, delivered=%v), attempt %d",
					uptime, delivered, failures)
			case failures == quietAfterFailures:
				log.Warnf("audio capture has failed %d times, and the last reason was "+
					"(uptime=%v, delivered=%v); it retries every %v from here, without "+
					"another line until it recovers",
					failures, uptime, delivered, s.maxBackoff)
			}
		}

		if s.isStopped() {
			return
		}

		select {
		case <-time.After(backoff):
		case <-s.done:
			return
		}

		if backoff *= 2; backoff > s.maxBackoff {
			backoff = s.maxBackoff
		}
	}
}

// runOnce starts one child and reads it to exhaustion. It reports whether the
// child delivered any audio and how long it ran.
func (s *Source) runOnce(chunk []byte, handle func([]byte)) (bool, time.Duration) {
	startTime := time.Now()

	cmd := s.newCmd()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorf("failed to open the audio capture pipe: %s", err)
		return false, time.Since(startTime)
	}

	// Keep the tail of stderr. A child that never delivers is the case that
	// needs explaining, and its reason is on stderr rather than in the exit
	// code.
	//
	// This takes a pipe and copies from it here, rather than handing os/exec a
	// writer. A writer makes Wait block until every holder of the write end
	// closes it, and that includes anything the child left behind, so a wedged
	// grandchild would hold Wait open for as long as it lived. Wait closes the
	// read end of a pipe as soon as it reaps the child, which ends the copy.
	stderr := &tailBuffer{}
	copied := make(chan struct{})

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Errorf("failed to open the audio capture error pipe: %s", err)
		return false, time.Since(startTime)
	}

	if err := cmd.Start(); err != nil {
		log.Errorf("failed to start audio capture: %s", err)
		return false, time.Since(startTime)
	}

	go func() {
		defer close(copied)

		_, _ = io.Copy(stderr, stderrPipe)
	}()

	// Store cmd and stdout together under one lock so Stop can close both
	// atomically, avoiding the window where one is recorded and the other is not.
	s.mutex.Lock()
	if s.stopped {
		s.mutex.Unlock()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		<-copied
		_ = stdout.Close()
		return false, time.Since(startTime)
	}
	s.cmd = cmd
	s.stdout = stdout
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		s.cmd = nil
		s.stdout = nil
		s.mutex.Unlock()
	}()

	var delivered bool

	for {
		if _, err := io.ReadFull(stdout, chunk); err != nil {
			break
		}

		delivered = true
		handle(chunk)
	}

	// Give the stderr copier a moment before reaping. cmd.Wait closes the read
	// end of the pipe, so whatever is still buffered is discarded and the
	// copier gets ErrFileClosed instead of the child's reason. arecord writes
	// that reason just before it exits, so on one core the copier can still be
	// behind us here - and the reason is the only diagnostic this feature has
	// on a device, because Available() reports an absent card in silence.
	select {
	case <-copied:
	case <-time.After(stderrDrainTimeout):
	}

	_ = cmd.Wait()

	// Wait reaped the child and closed the read end, so the copy above has
	// finished or is about to. Waiting for it is what makes the buffer safe to
	// read here.
	<-copied

	// A healthy child never gets here, so this costs nothing while audio works.
	if !delivered {
		if line := stderr.lastLine(); line != "" {
			log.Warnf("audio capture said: %s", line)
		}
	}

	return delivered, time.Since(startTime)
}

// Stop kills the child and stops the loop. It is safe to call more than once,
// and it is the only thing that unblocks a read while the host plays nothing.
func (s *Source) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.stopped {
		return
	}

	s.stopped = true

	// Close the done channel to wake the select in Run, and kill and close the
	// pipe under the same lock so there is no window where one is recorded and
	// the other is not.
	close(s.done)

	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}

	if s.stdout != nil {
		_ = s.stdout.Close()
	}
}

func (s *Source) isStopped() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.stopped
}
