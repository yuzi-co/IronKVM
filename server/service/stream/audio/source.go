package audio

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
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
	// writes a line per attempt for as long as a viewer listens. The log it
	// fills is /tmp/nanokvm-server.log, which S98vidiag and the supervisor
	// both read. A host that plays nothing is not counted as a failure: see
	// hostIdleReason.
	quietAfterFailures = 5

	// hostIdleReason is what arecord says when the host plays nothing to the
	// gadget. It does not block waiting for sound: measured on a board on
	// 2026-09-30, with the host's stream0 at "Playback Status: Stop", arecord
	// exits after about 0.4 s with "arecord: pcm_read:2285: read error: I/O
	// error", status 1, no samples and nothing in dmesg. That is this
	// device's ordinary idle state, not a fault, so it is told apart from one.
	hostIdleReason = "read error: I/O error"

	// stderrLimit caps what is kept of the child's stderr. ALSA states its
	// reason in one short line, so this is generous; the cap is there because
	// a child that complains once per period must not grow a buffer forever.
	stderrLimit = 512
)

// State is what capture is doing, as far as a viewer needs to know.
//
// The zero value means nothing is known yet. A Frame carries a State only when
// it is a notice rather than audio, so StateUnknown there means "this is audio".
type State uint8

const (
	StateUnknown State = iota
	// StatePlaying means the host plays and audio is being delivered.
	StatePlaying
	// StateIdle means the host plays nothing to the gadget, so there is no
	// audio to deliver. It turns into StatePlaying by itself once the host
	// plays something.
	StateIdle
	// StateFailing means capture failed for a reason other than an idle host.
	StateFailing
)

func (s State) String() string {
	switch s {
	case StatePlaying:
		return "playing"
	case StateIdle:
		return "idle"
	case StateFailing:
		return "failing"
	default:
		return "unknown"
	}
}

// isHostIdle reports whether a child's run is what an idle host looks like:
// it exited well short of a healthy run, delivered nothing, and said the read
// failed with an I/O error.
func isHostIdle(delivered bool, uptime time.Duration, reason string) bool {
	return !delivered && uptime < minRunDuration && strings.Contains(reason, hostIdleReason)
}

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

	// chunkBytes is 20 ms of the format the child delivers.
	chunkBytes int

	minBackoff time.Duration
	maxBackoff time.Duration

	// name is Format.Name: empty for the host's audio, otherwise the prefix of
	// every capture line. A named source has no host that may be idle, so an
	// I/O error from it is a failure like any other.
	name string

	// onState, when set, hears every change of state. It is set before Run
	// and called from Run's goroutine only.
	onState func(State)
	// state is the last state reported. Only Run's goroutine touches it.
	state State

	mutex   sync.Mutex
	cmd     *exec.Cmd
	stdout  io.ReadCloser
	stopped bool
	done    chan struct{}
}

func NewSource() *Source {
	return NewSourceFor(HostFormat)
}

// NewSourceFor reads one capture format.
func NewSourceFor(format Format) *Source {
	return &Source{
		newCmd:     func() *exec.Cmd { return arecordFor(format) },
		chunkBytes: format.ChunkBytes(),
		minBackoff: 200 * time.Millisecond,
		maxBackoff: 5 * time.Second,
		name:       format.Name,
		done:       make(chan struct{}),
	}
}

// newArecord reads the gadget and writes raw samples to stdout.
//
// The period is pinned to SamplesPerFrame, which is 20 ms and the same size
// as a chunk. Left to ALSA, the period comes from the driver's default and
// decides how much delay sits in front of the encoder. The gadget advertises
// PERIOD_SIZE [32 1024], so 960 is inside its range.
//
// The rate, channel count and period are built from the package constants
// rather than written out as literals, so this is actually the single set of
// constants the comment on those constants claims it is. Two literal sets
// that happen to agree today can drift silently: change SampleRate alone and
// arecord would still capture 48 kHz while the encoder assumed a different
// rate, and nothing on this path would catch it — io.ReadFull still fills the
// (now wrong) chunk size, opus_encode still succeeds, and the only symptom is
// audio at the wrong speed.
func newArecord() *exec.Cmd {
	return arecordFor(HostFormat)
}

// arecordFor reads one format's device the same way.
func arecordFor(format Format) *exec.Cmd {
	return exec.Command("arecord",
		"-D", format.Device,
		"-f", "S16_LE",
		"-r", strconv.Itoa(SampleRate),
		"-c", strconv.Itoa(format.Channels),
		"-t", "raw",
		"--period-size="+strconv.Itoa(SamplesPerFrame),
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
// climbs to maxBackoff, which also bounds how long a host that starts playing
// waits for its sound.
//
// An idle host makes arecord exit at once rather than wait for sound (see
// hostIdleReason), so that is not a failure. The loop says so once, at info,
// retries at the same cadence without another line, and says so once more
// when sound arrives. A real failure keeps its warnings, which fall quiet
// after quietAfterFailures.
func (s *Source) Run(handle func([]byte)) {
	chunkBytes := s.chunkBytes
	if chunkBytes == 0 {
		chunkBytes = ChunkBytes
	}
	chunk := make([]byte, chunkBytes)
	backoff := s.minBackoff

	var failures int
	var idle bool

	deliver := func(chunk []byte) {
		if idle {
			idle = false
			log.Infof("audio capture: the host is playing audio to the KVM again")
		}
		s.report(StatePlaying)
		handle(chunk)
	}

	for {
		if s.isStopped() {
			return
		}

		delivered, uptime, reason := s.runOnce(chunk, deliver)

		// A child killed by Stop is neither idle nor failing.
		if s.isStopped() {
			return
		}

		switch {
		case s.name == "" && isHostIdle(delivered, uptime, reason):
			// Not a failure, so it neither counts toward the quiet limit nor
			// warns. One line when it starts explains the silence.
			if !idle {
				idle = true
				log.Infof("audio capture: the host is not playing audio to the KVM; waiting for it")
			}

			failures = 0
			s.report(StateIdle)

		case delivered && uptime > minRunDuration:
			// The child produced audio and ran long enough, so the next failure
			// is a fresh one.
			if failures >= quietAfterFailures {
				if s.name != "" {
					log.Infof("%s: capture recovered after %d failed attempts", s.name, failures)
				} else {
					log.Infof("audio capture recovered after %d failed attempts", failures)
				}
			}

			backoff = s.minBackoff
			failures = 0

		default:
			// Every arrival here is a child that failed for a reason other
			// than an idle host, or one that did not run long enough to count.
			failures++
			idle = false
			s.report(StateFailing)

			if s.name != "" {
				s.logNamedFailure(failures, uptime, delivered, reason)
				break
			}

			// The child's own reason rides on these lines rather than on one
			// of its own, so it falls quiet with them. The reason is the only
			// diagnostic this feature has, so it is carried, not dropped.
			said := ""
			if reason != "" {
				said = fmt.Sprintf(", and it said: %s", reason)
			}

			switch {
			case failures < quietAfterFailures:
				log.Warnf("audio capture exited (uptime=%v, delivered=%v), attempt %d%s",
					uptime, delivered, failures, said)
			case failures == quietAfterFailures:
				log.Warnf("audio capture has failed %d times, and the last reason was "+
					"(uptime=%v, delivered=%v)%s; it retries every %v from here, without "+
					"another line until it recovers",
					failures, uptime, delivered, said, s.maxBackoff)
			}
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

// logNamedFailure writes a named source's failure under its own prefix, with
// the child's reason first. It falls quiet after quietAfterFailures like the
// host's lines do.
func (s *Source) logNamedFailure(failures int, uptime time.Duration, delivered bool, reason string) {
	if reason == "" {
		reason = "arecord exited without saying why"
	}

	switch {
	case failures < quietAfterFailures:
		log.Warnf("%s: capture failed: %s (uptime=%v, delivered=%v, attempt %d)",
			s.name, reason, uptime, delivered, failures)
	case failures == quietAfterFailures:
		log.Warnf("%s: capture failed: %s (uptime=%v, delivered=%v); it has failed %d times "+
			"and retries every %v from here, without another line until it recovers",
			s.name, reason, uptime, delivered, failures, s.maxBackoff)
	}
}

// report passes a change of state to onState. A repeat is dropped, so calling
// it for every chunk costs a comparison.
func (s *Source) report(state State) {
	if state == s.state {
		return
	}

	s.state = state
	if s.onState != nil {
		s.onState(state)
	}
}

// runOnce starts one child and reads it to exhaustion. It reports whether the
// child delivered any audio, how long it ran, and the last thing it said.
//
// It reports the reason rather than logging it. Whether a line is worth writing
// depends on how many attempts have failed already, and that is only known in
// Run. Logging here wrote the same sentence every retry for as long as the
// board was up.
func (s *Source) runOnce(chunk []byte, handle func([]byte)) (bool, time.Duration, string) {
	startTime := time.Now()

	cmd := s.newCmd()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return false, time.Since(startTime), fmt.Sprintf("the capture pipe would not open: %s", err)
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
		return false, time.Since(startTime), fmt.Sprintf("the capture error pipe would not open: %s", err)
	}

	// Reported rather than logged, for the same reason as the child's own
	// complaint: a capture command that cannot start cannot start on every
	// retry either, and one line per attempt for the life of the board is not
	// more informative than one line.
	if err := cmd.Start(); err != nil {
		return false, time.Since(startTime), fmt.Sprintf("the capture command would not start: %s", err)
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

		// Stopping is not a failure and has no reason to report.
		return false, time.Since(startTime), ""
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
	var reason string
	if !delivered {
		reason = stderr.lastLine()
	}

	return delivered, time.Since(startTime), reason
}

// Stop kills the child and stops the loop. It is safe to call more than once.
// It is the only thing that ends a read while the host plays, and it cuts short
// the wait between attempts while the host plays nothing.
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
