package dashcam

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/pmp4"

	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/events"
	"NanoKVM-Server/service/stream/egress"
)

// fakeReader is the tap: the test pushes into the service itself.
type fakeReader struct {
	units  chan egress.Unit
	mu     sync.Mutex
	closed bool
}

func (r *fakeReader) Units() <-chan egress.Unit { return r.units }

func (r *fakeReader) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		r.closed = true
		close(r.units)
	}
}

type fixture struct {
	svc      *Service
	dir      string
	readers  []*fakeReader
	keyAsks  int
	records  []audit.Event
	mu       sync.Mutex
	now      time.Time
	settings Settings
}

// tick is the unit of the settings' times in these tests: a "second" is 10 ms,
// so a clip with 5 s after its trigger is written 50 ms after it.
const tick = 10 * time.Millisecond

func newFixture(t *testing.T, settings Settings) *fixture {
	t.Helper()
	f := &fixture{dir: t.TempDir(), now: time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC), settings: settings}
	f.svc = New(Deps{
		Load: func() (Settings, error) { return f.settings, nil },
		Save: func(s Settings) error { f.settings = s; return nil },
		Subscribe: func() Reader {
			r := &fakeReader{units: make(chan egress.Unit)}
			f.mu.Lock()
			f.readers = append(f.readers, r)
			f.mu.Unlock()
			return r
		},
		RequestKeyframe: func() {
			f.mu.Lock()
			f.keyAsks++
			f.mu.Unlock()
		},
		Bus: events.NewBus(),
		Record: func(e audit.Event) {
			f.mu.Lock()
			f.records = append(f.records, e)
			f.mu.Unlock()
		},
		Dir:    f.dir,
		Now:    f.clock,
		Second: tick,
	})
	f.svc.Start()
	t.Cleanup(f.svc.Close)
	return f
}

func (f *fixture) clock() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *fixture) advance(d time.Duration) {
	f.mu.Lock()
	f.now = f.now.Add(d)
	f.mu.Unlock()
}

// feed pushes samples as the reader would.
func (f *fixture) feed(samples []Sample) {
	f.svc.mu.Lock()
	gen := f.svc.gen
	f.svc.mu.Unlock()
	for _, s := range samples {
		f.svc.push(gen, s)
	}
}

func (f *fixture) clips(t *testing.T) []Clip {
	t.Helper()
	f.svc.writes.Wait()
	clips, err := listClips(f.dir)
	if err != nil {
		t.Fatal(err)
	}
	return clips
}

func enabled() Settings {
	s := Defaults()
	s.Enabled = true
	s.PostSeconds = 0
	return s
}

// readClip parses a clip back and answers its samples.
func readClip(t *testing.T, path string) []*pmp4.Sample {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var p pmp4.Presentation
	if err := p.Unmarshal(bytes.NewReader(data)); err != nil {
		t.Fatalf("%s does not parse: %s", path, err)
	}
	return p.Tracks[0].Samples
}

// Off, the dashcam holds no reader, so it keeps no encoder running and the
// signal is not watched for it; on, it holds one; off again, it lets go.
func TestOffKeepsNothingRunning(t *testing.T) {
	f := newFixture(t, Defaults())
	if f.svc.running() || len(f.readers) != 0 || SignalWanted() {
		t.Fatalf("an off dashcam subscribed %d times", len(f.readers))
	}
	if _, err := f.svc.Trigger(TriggerManual, true); !errors.Is(err, ErrDisabled) {
		t.Fatalf("a trigger while off: %v", err)
	}

	if err := f.svc.Configure(enabled()); err != nil {
		t.Fatal(err)
	}
	if !f.svc.running() || len(f.readers) != 1 || !SignalWanted() {
		t.Fatalf("on: running %t, %d readers", f.svc.running(), len(f.readers))
	}
	f.feed(h264Clip(30, 30, 30))

	off := enabled()
	off.Enabled = false
	if err := f.svc.Configure(off); err != nil {
		t.Fatal(err)
	}
	if f.svc.running() || !f.readers[0].closed || SignalWanted() {
		t.Fatal("off again, the reader is still held")
	}
	if f.svc.State().BufferedBytes != 0 {
		t.Fatal("off again, the ring still holds video")
	}
}

// With video off, as in failsafe mode, the dashcam does not subscribe.
func TestNoVideoKeepsNothingRunning(t *testing.T) {
	f := &fixture{dir: t.TempDir()}
	svc := New(Deps{
		Load:           func() (Settings, error) { return enabled(), nil },
		Subscribe:      func() Reader { t.Fatal("subscribed without video"); return nil },
		VideoAvailable: func() bool { return false },
		Dir:            f.dir,
	})
	svc.Start()
	defer svc.Close()
	if svc.running() || svc.State().Running {
		t.Fatal("running without video")
	}
}

func TestATriggerSavesTheRing(t *testing.T) {
	f := newFixture(t, enabled())
	// 3 s at 30 fps; the window is 60 ticks, 0.6 s, so the ring keeps two
	// GOPs of 10 at most.
	f.feed(h264Clip(90, 30, 10))

	if _, err := f.svc.Trigger(TriggerManual, true); err != nil {
		t.Fatal(err)
	}
	clips := f.clips(t)
	if len(clips) != 1 || !strings.HasSuffix(clips[0].Name, "-manual.mp4") {
		t.Fatalf("clips: %+v", clips)
	}
	samples := readClip(t, filepath.Join(f.dir, clips[0].Name))
	if len(samples) < 18 || len(samples) > 30 || samples[0].IsNonSyncSample {
		t.Fatalf("the clip has %d samples, the first a keyframe: %t", len(samples), !samples[0].IsNonSyncSample)
	}
	// A clip asked for by hand is in the log from its route.
	if len(f.records) != 0 {
		t.Fatalf("records: %+v", f.records)
	}
}

// A trigger while a clip collects joins it: one file, named for both, holding
// the stream after the trigger too.
func TestTriggersJoinOneClip(t *testing.T) {
	s := enabled()
	s.PostSeconds = 10 // 100 ms
	f := newFixture(t, s)
	f.feed(h264Clip(30, 30, 10))
	pre := f.svc.ring.Len()

	if _, err := f.svc.Trigger(TriggerReset, false); err != nil {
		t.Fatal(err)
	}
	p, err := f.svc.Trigger(TriggerSignalLost, false)
	if err != nil || !p.Joined {
		t.Fatalf("the second trigger: %+v, %v", p, err)
	}
	// The same trigger again is not named twice.
	if _, err := f.svc.Trigger(TriggerReset, false); err != nil {
		t.Fatal(err)
	}
	more := h264Clip(40, 30, 10)[30:]
	f.feed(more)

	clips := f.clips(t)
	if len(clips) != 1 || !strings.HasSuffix(clips[0].Name, "-reset+signal-lost.mp4") {
		t.Fatalf("clips: %+v", clips)
	}
	if got := len(readClip(t, filepath.Join(f.dir, clips[0].Name))); got != pre+len(more) {
		t.Fatalf("the clip has %d samples, want %d before and %d after", got, pre, len(more))
	}
	if len(f.records) != 1 || f.records[0].Action != "dashcam-save" || f.records[0].Outcome != audit.OutcomeOK ||
		f.records[0].Target != clips[0].Name || f.records[0].Actor != "dashcam" {
		t.Fatalf("records: %+v", f.records)
	}
}

// After a clip, the next waits minGap, and the events save at most
// autoPerHour in an hour; the button is not counted.
func TestTriggersAreRateLimited(t *testing.T) {
	f := newFixture(t, enabled())
	f.feed(h264Clip(30, 30, 10))

	if _, err := f.svc.Trigger(TriggerPower, false); err != nil {
		t.Fatal(err)
	}
	f.svc.writes.Wait()
	if _, err := f.svc.Trigger(TriggerManual, true); !errors.Is(err, ErrTooSoon) {
		t.Fatalf("a trigger right after a clip: %v", err)
	}

	for i := 1; i < autoPerHour; i++ {
		f.advance(minGap * tick)
		if _, err := f.svc.Trigger(TriggerPower, false); err != nil {
			t.Fatalf("clip %d: %v", i+1, err)
		}
		f.svc.writes.Wait()
	}
	f.advance(minGap * tick)
	if _, err := f.svc.Trigger(TriggerPower, false); !errors.Is(err, errTooMany) {
		t.Fatalf("clip %d in an hour: %v", autoPerHour+1, err)
	}
	if _, err := f.svc.Trigger(TriggerManual, true); err != nil {
		t.Fatalf("the button after the events' limit: %v", err)
	}
	f.svc.writes.Wait()

	f.advance(hour * tick)
	if _, err := f.svc.Trigger(TriggerPower, false); err != nil {
		t.Fatalf("an hour later: %v", err)
	}
}

// The events the settings name save a clip; the others do not.
func TestEventsSaveAClip(t *testing.T) {
	s := enabled()
	s.Triggers.Signal = false
	f := newFixture(t, s)
	f.feed(h264Clip(30, 30, 10))

	f.svc.onEvent(events.Event{Type: events.SignalLost})
	f.svc.onEvent(events.Event{Type: events.LoginFailure})
	if f.svc.State().Pending != nil || len(f.clips(t)) != 0 {
		t.Fatal("an event the settings leave off saved a clip")
	}

	f.svc.deps.Bus.Publish(events.Event{Type: events.ATXPress, Fields: map[string]string{"button": "reset"}})
	// The bus delivers on its own goroutine; wait for the clip by the state,
	// since a wait on the writes could race the trigger's start of one.
	deadline := time.Now().Add(2 * time.Second)
	for len(f.svc.State().LastClip) == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	clips := f.clips(t)
	if len(clips) != 1 || !strings.HasSuffix(clips[0].Name, "-reset.mp4") {
		t.Fatalf("clips: %+v", clips)
	}
	if got := clips[0].Triggers; len(got) != 1 || got[0] != TriggerReset {
		t.Fatalf("triggers read back: %v", got)
	}
}

// An empty ring has nothing to save by hand.
func TestNothingToSave(t *testing.T) {
	f := newFixture(t, enabled())
	if _, err := f.svc.Trigger(TriggerManual, true); !errors.Is(err, ErrEmpty) {
		t.Fatalf("a save of an empty ring: %v", err)
	}
}

// A resolution change inside a clip gives one file per resolution.
func TestAClipSplitsAtAResolutionChange(t *testing.T) {
	s := enabled()
	s.Seconds = MaxSeconds // 1.2 s
	f := newFixture(t, s)
	first := h264Clip(20, 30, 10)
	second := h264Clip(20, 30, 10)
	otherSPS := append([]byte(nil), second[0].NALUs[1]...)
	otherSPS[len(otherSPS)-1] ^= 1
	for i := range second {
		second[i].PTS += 700 * time.Millisecond
		if second[i].Key {
			second[i].NALUs = [][]byte{otherSPS, second[i].NALUs[2], second[i].NALUs[3]}
		}
	}
	f.feed(append(first, second...))
	if _, err := f.svc.Trigger(TriggerSignalRestored, false); err != nil {
		t.Fatal(err)
	}
	clips := f.clips(t)
	if len(clips) != 2 || !strings.HasSuffix(clips[0].Name, "-signal-restored-part2.mp4") ||
		!strings.HasSuffix(clips[1].Name, "-signal-restored-part1.mp4") {
		t.Fatalf("clips: %+v", clips)
	}
}

// The ring asks for a keyframe when its last one is keyframeEvery old, and no
// more often; an encoder that makes them often enough is never asked.
func TestKeyframeRequests(t *testing.T) {
	f := newFixture(t, enabled())
	// One keyframe, then 100 ticks of delta frames at a frame a tick.
	stream := gopStream(101, 100, 1000, 10)
	for i := range stream {
		stream[i].PTS = time.Duration(i) * tick
	}
	f.feed(stream)
	// Asked at 4, 8, ... 100 ticks.
	if f.keyAsks != 100/keyframeEvery {
		t.Fatalf("asked %d times, want %d", f.keyAsks, 100/keyframeEvery)
	}

	g := newFixture(t, enabled())
	stream = gopStream(101, 100, 2, 10)
	for i := range stream {
		stream[i].PTS = time.Duration(i) * tick
	}
	g.feed(stream)
	if g.keyAsks != 0 {
		t.Fatalf("an encoder with keyframes every 2 ticks was asked %d times", g.keyAsks)
	}
}

// The storage cap deletes the oldest clips and keeps the one just written.
func TestPruneKeepsTheStorageCap(t *testing.T) {
	dir := t.TempDir()
	names := []string{
		"20261027-100000-power.mp4",
		"20261027-110000-reset.mp4",
		"20261027-120000-manual.mp4",
		"20261027-130000-api.mp4",
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), make([]byte, 100), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Not a clip: never counted or deleted.
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), make([]byte, 1000), 0o644); err != nil {
		t.Fatal(err)
	}

	removed, err := prune(dir, 250, []string{names[3]})
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 2 || removed[0] != names[0] || removed[1] != names[1] {
		t.Fatalf("removed %v", removed)
	}

	// A clip just written larger than the cap stays.
	removed, err = prune(dir, 50, []string{names[3]})
	if err != nil || len(removed) != 1 || removed[0] != names[2] {
		t.Fatalf("removed %v, %v", removed, err)
	}
	clips, _ := listClips(dir)
	if len(clips) != 1 || clips[0].Name != names[3] {
		t.Fatalf("left %+v", clips)
	}
	if _, err := os.Stat(filepath.Join(dir, "notes.txt")); err != nil {
		t.Fatal("a file that is not a clip was deleted")
	}
}

func TestClipNames(t *testing.T) {
	for name, want := range map[string]bool{
		"20261027-120000-manual.mp4":                  true,
		"20261027-120000-reset+signal-lost-part2.mp4": true,
		"20261027-120000-api-2.mp4":                   true,
		"../20261027-120000-manual.mp4":               false,
		"20261027-120000-manual.mp4.tmp":              false,
		".20261027-120000-manual.mp4.tmp":             false,
		"20261027-120000-Manual.mp4":                  false,
		"20261027-120000-a/b.mp4":                     false,
		"notes.mp4":                                   false,
	} {
		if ValidName(name) != want {
			t.Errorf("ValidName(%q) = %t", name, !want)
		}
	}
}

func TestSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dashcam.json")
	s, err := Load(path)
	if err != nil || s != Defaults() || s.Enabled {
		t.Fatalf("a missing file loads %+v, %v", s, err)
	}

	// A key the file lacks keeps its default.
	if err := os.WriteFile(path, []byte(`{"enabled":true,"seconds":90}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err = Load(path)
	if err != nil || !s.Enabled || s.Seconds != 90 || s.PostSeconds != DefaultPostSeconds || !s.Triggers.ATX {
		t.Fatalf("loaded %+v, %v", s, err)
	}

	for _, bad := range []Settings{
		{Seconds: 29, StorageMB: 200},
		{Seconds: 121, StorageMB: 200},
		{Seconds: 60, PostSeconds: 31, StorageMB: 200},
		{Seconds: 60, PostSeconds: -1, StorageMB: 200},
		{Seconds: 60, MaxKbps: 100, StorageMB: 200},
		{Seconds: 60, StorageMB: 10},
	} {
		if bad.Validate() == nil {
			t.Errorf("%+v passed", bad)
		}
		if Save(path, bad) == nil {
			t.Errorf("%+v saved", bad)
		}
	}

	good := Defaults()
	good.Enabled, good.MaxKbps, good.PostSeconds = true, 1000, 0
	if err := Save(path, good); err != nil {
		t.Fatal(err)
	}
	if s, err := Load(path); err != nil || s != good {
		t.Fatalf("read back %+v, %v", s, err)
	}
}

// The ring's budget follows the bitrate cap, never past MaxRingBytes.
func TestBudget(t *testing.T) {
	svc := New(Deps{})
	s := Defaults()
	if got := svc.budget(s); got != MaxRingBytes {
		t.Fatalf("the stream's own bitrate: %d", got)
	}
	s.MaxKbps = 1000
	if got, want := svc.budget(s), 1000*1000/8*(DefaultSeconds+keyframeEvery); got != want {
		t.Fatalf("1000 kbit/s: %d, want %d", got, want)
	}
	s.MaxKbps = MaxKbps
	if got := svc.budget(s); got != MaxRingBytes {
		t.Fatalf("the highest cap: %d", got)
	}
}
