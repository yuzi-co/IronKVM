package roommic

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"NanoKVM-Server/service/stream/audio"
)

// fakeCapture stands in for arecord and the encoder.
type fakeCapture struct {
	frames chan []byte
	once   sync.Once
}

func (f *fakeCapture) Start()                {}
func (f *fakeCapture) Frames() <-chan []byte { return f.frames }
func (f *fakeCapture) Stop()                 { f.once.Do(func() { close(f.frames) }) }

// rig is a manager over fake captures that records what it did.
type rig struct {
	t        *testing.T
	mutex    sync.Mutex
	started  int
	gains    []int
	captures []*fakeCapture
	card     bool
	manager  *Manager
}

func newRig(t *testing.T) *rig {
	t.Helper()

	original := SettingsFile
	SettingsFile = filepath.Join(t.TempDir(), "room-mic")
	t.Cleanup(func() { SettingsFile = original })

	r := &rig{t: t, card: true}

	originalGain := applyGain
	applyGain = func(gain int) error {
		r.mutex.Lock()
		defer r.mutex.Unlock()
		r.gains = append(r.gains, gain)
		return nil
	}
	t.Cleanup(func() { applyGain = originalGain })

	r.manager = NewManager(func() bool {
		r.mutex.Lock()
		defer r.mutex.Unlock()
		return r.card
	}, func() audio.Capture {
		r.mutex.Lock()
		defer r.mutex.Unlock()
		r.started++
		capture := &fakeCapture{frames: make(chan []byte)}
		r.captures = append(r.captures, capture)
		return capture
	})

	return r
}

func (r *rig) allow(allowed bool) {
	r.t.Helper()
	settings := r.manager.Settings()
	settings.Allowed = allowed
	if err := r.manager.SetSettings(settings, "admin"); err != nil {
		r.t.Fatal(err)
	}
}

func (r *rig) open(user string) *Listener {
	r.t.Helper()
	listener, err := r.manager.Open(user, "test")
	if err != nil {
		r.t.Fatalf("open for %s: %s", user, err)
	}
	return listener
}

func (r *rig) captureCount() int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.started
}

// stopped reports whether the newest capture has been stopped.
func (r *rig) stopped() bool {
	r.mutex.Lock()
	capture := r.captures[len(r.captures)-1]
	r.mutex.Unlock()

	select {
	case _, ok := <-capture.frames:
		return !ok
	default:
		return false
	}
}

func TestTheMicrophoneIsOffByDefault(t *testing.T) {
	r := newRig(t)

	if settings := r.manager.Settings(); settings.Allowed || settings.Gain != DefaultGain {
		t.Fatalf("defaults are %+v, want not allowed at gain %d", settings, DefaultGain)
	}
	if _, err := r.manager.Open("alice", "test"); !errors.Is(err, ErrNotAllowed) {
		t.Fatalf("opening without the setting gave %v, want ErrNotAllowed", err)
	}
	if r.captureCount() != 0 {
		t.Fatal("a capture started although the microphone is not allowed")
	}
}

func TestWithoutTheCardNothingOpens(t *testing.T) {
	r := newRig(t)
	r.allow(true)
	r.card = false

	if _, err := r.manager.Open("alice", "test"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("opening without the card gave %v, want ErrUnavailable", err)
	}
	if r.captureCount() != 0 {
		t.Fatal("a capture started on a kernel without the card")
	}
}

// The first listener opens the microphone, the others share it, and it closes
// with the last.
func TestTheMicrophoneOpensForTheFirstAndClosesAfterTheLast(t *testing.T) {
	r := newRig(t)
	r.allow(true)

	alice := r.open("alice")
	if r.captureCount() != 1 || !r.manager.Live() {
		t.Fatalf("after one listener: %d captures, live %v", r.captureCount(), r.manager.Live())
	}

	bob := r.open("bob")
	if r.captureCount() != 1 {
		t.Fatalf("a second listener started a second capture: %d", r.captureCount())
	}

	if got := r.manager.Status().Listeners; !slices.Equal(got, []string{"alice", "bob"}) {
		t.Fatalf("listeners %v", got)
	}

	alice.Close()
	if !r.manager.Live() || r.stopped() {
		t.Fatal("the microphone closed while bob still listens")
	}

	bob.Close()
	if r.manager.Live() || !r.stopped() {
		t.Fatal("the microphone is still open after the last listener left")
	}

	// Closing twice changes nothing.
	bob.Close()
	if r.manager.Live() {
		t.Fatal("a second close reopened it")
	}

	// And it opens afresh for the next one.
	r.open("carol").Close()
	if r.captureCount() != 2 {
		t.Fatalf("the next listener should start a new capture, %d so far", r.captureCount())
	}
}

// One account in two tabs is listed once.
func TestAListenerIsNamedOnce(t *testing.T) {
	r := newRig(t)
	r.allow(true)

	r.open("alice")
	r.open("alice")

	if got := r.manager.Status().Listeners; !slices.Equal(got, []string{"alice"}) {
		t.Fatalf("listeners %v", got)
	}
}

func TestTheGainIsSetBeforeEachCapture(t *testing.T) {
	r := newRig(t)
	r.allow(true)

	r.open("alice").Close()
	if !slices.Equal(r.gains, []int{DefaultGain}) {
		t.Fatalf("gains applied %v, want [%d]", r.gains, DefaultGain)
	}

	// A change while nobody listens waits for the next capture.
	settings := r.manager.Settings()
	settings.Gain = 10
	if err := r.manager.SetSettings(settings, "admin"); err != nil {
		t.Fatal(err)
	}
	if len(r.gains) != 1 {
		t.Fatalf("the gain was applied with the microphone closed: %v", r.gains)
	}

	listener := r.open("alice")
	// A change while it is open applies at once.
	settings.Gain = 12
	if err := r.manager.SetSettings(settings, "admin"); err != nil {
		t.Fatal(err)
	}
	listener.Close()

	if !slices.Equal(r.gains, []int{DefaultGain, 10, 12}) {
		t.Fatalf("gains applied %v, want [%d 10 12]", r.gains, DefaultGain)
	}
}

// Disallowing closes the microphone for everyone at once.
func TestDisallowingClosesItForEveryone(t *testing.T) {
	r := newRig(t)
	r.allow(true)

	alice := r.open("alice")
	r.open("bob")

	r.allow(false)

	if r.manager.Live() || !r.stopped() {
		t.Fatal("the microphone stayed open after it was disallowed")
	}
	if _, ok := <-alice.Frames(); ok {
		t.Fatal("a listener's frames stayed open")
	}
	if _, err := r.manager.Open("alice", "test"); !errors.Is(err, ErrNotAllowed) {
		t.Fatalf("opening after it was disallowed gave %v", err)
	}
}

func TestWatchersHearEveryChange(t *testing.T) {
	r := newRig(t)
	r.allow(true)

	var seen []bool
	cancel := r.manager.Watch(func() { seen = append(seen, r.manager.Live()) })

	listener := r.open("alice")
	listener.Close()
	cancel()
	r.open("bob")

	if !slices.Equal(seen, []bool{true, false}) {
		t.Fatalf("watcher saw %v, want [true false]", seen)
	}
}

func TestSettingsArePersisted(t *testing.T) {
	r := newRig(t)

	if err := r.manager.SetSettings(Settings{Allowed: true, Gain: 6}, "admin"); err != nil {
		t.Fatal(err)
	}

	// A new manager, as after a restart, reads them back.
	fresh := NewManager(func() bool { return true }, func() audio.Capture { return &fakeCapture{frames: make(chan []byte)} })
	if got := fresh.Settings(); got != (Settings{Allowed: true, Gain: 6}) {
		t.Fatalf("read back %+v", got)
	}

	if err := r.manager.SetSettings(Settings{Allowed: true, Gain: 25}, "admin"); err == nil {
		t.Fatal("a gain above 24 was accepted")
	}
	if err := r.manager.SetSettings(Settings{Allowed: true, Gain: -1}, "admin"); err == nil {
		t.Fatal("a negative gain was accepted")
	}
}

func TestAnUnreadableFileKeepsItOff(t *testing.T) {
	r := newRig(t)

	if err := os.WriteFile(SettingsFile, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if r.manager.Settings().Allowed {
		t.Fatal("a corrupt settings file allowed the microphone")
	}
}
