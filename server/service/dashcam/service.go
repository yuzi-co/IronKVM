// Package dashcam keeps the last minute or two of the video in memory and
// saves it as an MP4 file on /data when something happens to the host: a
// press of the power or reset button, the HDMI signal going or coming back, a
// press of the button on the page, or a call of the save route
// (ironkvm-dist#119). It catches what nobody was watching: the crash screen
// before a reset, the firmware screens after it.
//
// It does no media work of its own. It is one more reader of the tap the RTSP
// server and the SRT listener read (service/stream/egress), so it takes the
// H.264 or H.265 the hardware encoder already makes, at the settings the
// viewers have, and never changes them. While it is on, that reader keeps the
// encoder running with nobody watching; while it is off, it holds nothing and
// keeps nothing running.
//
// The access units go into a ring in memory (ring.go) that always starts on a
// keyframe and is bounded by time and by bytes. On a trigger the ring is
// copied, the stream after the trigger is collected for a few seconds more,
// and the whole is written once as a regular MP4 file without encoding
// anything again (mp4.go). Nothing is written to the card until then.
package dashcam

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/events"
	"NanoKVM-Server/service/stream/egress"
)

// The triggers, as clip names and the audit log spell them.
const (
	TriggerPower          = "power"
	TriggerReset          = "reset"
	TriggerWatchdog       = "watchdog"
	TriggerSignalLost     = "signal-lost"
	TriggerSignalRestored = "signal-restored"
	TriggerFailsafe       = "failsafe"
	// TriggerManual is the button on the page, TriggerAPI the save route
	// called with an API key.
	TriggerManual = "manual"
	TriggerAPI    = "api"
)

// MaxRingBytes bounds the ring whatever the settings. At the default 3000
// kbit/s it holds about 43 s of a busy picture; a desktop that changes little
// costs a fraction of its bitrate, and holds the full window.
const MaxRingBytes = 16 << 20

// maxPostBytes bounds what is collected after a trigger, beside the ring.
const maxPostBytes = MaxRingBytes / 2

// The times below are in units of Deps.Second, which is a second on the board
// and much less in a test.
const (
	// keyframeEvery is how long the ring goes without a keyframe before the
	// encoder is asked for one. The ring can only start a clip on a
	// keyframe, so it holds the window plus up to this much. Slot A's encoder
	// makes one every 2 s at most by itself and is never asked; slot B's
	// makes them on request only, and is asked this often.
	keyframeEvery = 4
	// minGap is the least time between the end of one clip and the start
	// of the next. A trigger while a clip is collecting joins that clip.
	minGap = 10
	// maxExtend is how far triggers that join a clip can push its end out,
	// beyond the time after the first one.
	maxExtend = 30
	// autoPerHour is how many clips the events may save in an hour, so a
	// signal that comes and goes does not rotate every older clip out of the
	// storage cap. The button and the save route are not counted.
	autoPerHour = 12
	hour        = 3600
)

var (
	ErrDisabled = errors.New("the dashcam is off")
	ErrEmpty    = errors.New("the dashcam holds no video yet")
	ErrTooSoon  = errors.New("a clip was saved a moment ago; try again in a few seconds")
	errTooMany  = errors.New("too many clips in the last hour")
)

// Reader is the tap seen from the dashcam. *egress.Reader is the one the
// server runs with.
type Reader interface {
	Units() <-chan egress.Unit
	Close()
}

// Deps is everything the service reaches outside this package.
type Deps struct {
	// Load reads the settings at start, Save writes them on a change.
	Load func() (Settings, error)
	Save func(Settings) error
	// Subscribe joins the tap, which starts the encoder if nobody watches.
	Subscribe func() Reader
	// RequestKeyframe asks the encoder for a keyframe.
	RequestKeyframe func()
	// VideoAvailable reports whether the board captures at all: a server in
	// failsafe mode does not.
	VideoAvailable func() bool
	Bus            *events.Bus
	// Record adds an event to the audit log.
	Record func(audit.Event)
	Dir    string
	Now    func() time.Time
	// Second is the unit of the settings' times. Zero means time.Second.
	Second time.Duration
}

// pending is a clip that collects the stream after its trigger.
type pending struct {
	at       time.Time
	triggers []string
	// manual is a clip a person asked for, whose request the route already
	// recorded in the audit log.
	manual bool

	pre       []Sample
	post      []Sample
	postBytes int
	postDone  bool

	deadline    time.Time
	maxDeadline time.Time
	timer       *time.Timer
}

// Service is the dashcam.
type Service struct {
	deps Deps

	mu       sync.Mutex
	settings Settings
	ring     *Ring
	reader   Reader
	gen      uint64
	pending  *pending

	// lastClosed is when the last clip stopped collecting.
	lastClosed time.Time
	autoTimes  []time.Time

	// lastKey and lastAsk are the tap's clock at the last keyframe and the
	// last request for one.
	lastKey time.Duration
	lastAsk time.Duration
	hasKey  bool
	hasAsk  bool

	lastClip  []string
	lastError string

	writeMu sync.Mutex
	writes  sync.WaitGroup

	startOnce   sync.Once
	unsubscribe func()
}

// signalWanted is true while a running dashcam saves on the signal events, so
// the notifier watches the signal for it (notify.Deps.SignalWanted).
var signalWanted atomic.Bool

// SignalWanted reports whether the dashcam needs the signal events.
func SignalWanted() bool {
	return signalWanted.Load()
}

// New returns a service that is off. Start loads the settings and applies
// them.
func New(deps Deps) *Service {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.Second <= 0 {
		deps.Second = time.Second
	}
	if deps.Dir == "" {
		deps.Dir = Dir
	}
	if deps.Record == nil {
		deps.Record = func(audit.Event) {}
	}
	if deps.VideoAvailable == nil {
		deps.VideoAvailable = func() bool { return true }
	}
	if deps.RequestKeyframe == nil {
		deps.RequestKeyframe = func() {}
	}
	s := &Service{deps: deps, settings: Defaults()}
	s.ring = NewRing(s.window(s.settings), s.budget(s.settings))
	return s
}

// Start loads the settings, applies them, and listens for the events that
// save a clip. Calling it again does nothing.
func (s *Service) Start() {
	s.startOnce.Do(func() {
		removeTemporary(s.deps.Dir)

		settings := Defaults()
		if s.deps.Load != nil {
			loaded, err := s.deps.Load()
			if err != nil {
				log.Warnf("dashcam: settings: %s; the dashcam stays off", err)
			} else {
				settings = loaded
			}
		}
		s.apply(settings)

		if s.deps.Bus != nil {
			s.unsubscribe = s.deps.Bus.Subscribe(s.onEvent)
		}
	})
}

// Close stops the service: the reader, the events and a clip that collects.
// A clip already being written is finished first.
func (s *Service) Close() {
	if s.unsubscribe != nil {
		s.unsubscribe()
	}
	s.mu.Lock()
	s.stopLocked()
	p := s.pending
	s.mu.Unlock()
	if p != nil && p.timer.Stop() {
		s.finish(p)
	}
	s.writes.Wait()
}

func (s *Service) window(settings Settings) time.Duration {
	return time.Duration(settings.Seconds) * s.deps.Second
}

// budget is the ring's bound in bytes: the cap's bitrate over the window and
// the GOP beyond it, and never more than MaxRingBytes.
func (s *Service) budget(settings Settings) int {
	if settings.MaxKbps <= 0 {
		return MaxRingBytes
	}
	seconds := settings.Seconds + keyframeEvery
	return min(MaxRingBytes, settings.MaxKbps*1000/8*seconds)
}

// Settings answers the settings in force.
func (s *Service) Settings() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.settings
}

// Configure saves the settings and applies them.
func (s *Service) Configure(settings Settings) error {
	if err := settings.Validate(); err != nil {
		return err
	}
	if s.deps.Save != nil {
		if err := s.deps.Save(settings); err != nil {
			return err
		}
	}
	s.apply(settings)
	return nil
}

func (s *Service) apply(settings Settings) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings = settings
	run := settings.Enabled && s.deps.Subscribe != nil && s.deps.VideoAvailable()
	switch {
	case run && s.reader == nil:
		s.gen++
		s.hasKey, s.hasAsk = false, false
		s.reader = s.deps.Subscribe()
		go s.run(s.reader, s.gen)
		log.Infof("dashcam: on, keeping %d s of video", settings.Seconds)
	case !run && s.reader != nil:
		s.stopLocked()
		log.Info("dashcam: off")
	}
	s.ring.SetBounds(s.window(settings), s.budget(settings))
	signalWanted.Store(run && settings.Triggers.Signal)
}

// stopLocked lets go of the tap and of the video the ring holds. A clip that
// collects keeps what it has.
func (s *Service) stopLocked() {
	if s.reader == nil {
		return
	}
	s.reader.Close()
	s.reader = nil
	s.gen++
	s.ring.Reset()
	signalWanted.Store(false)
}

// running reports whether the service holds a reader, for the tests.
func (s *Service) running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reader != nil
}

func (s *Service) run(reader Reader, gen uint64) {
	for unit := range reader.Units() {
		if unit.Audio {
			continue
		}
		if !s.push(gen, Sample{Codec: unit.Codec, Key: unit.KeyFrame, NALUs: unit.NALUs, PTS: unit.PTS}) {
			return
		}
	}
}

// push adds a unit to the ring, and to a clip that collects. It answers false
// once the reader it came from is no longer the service's.
func (s *Service) push(gen uint64, sample Sample) bool {
	s.mu.Lock()
	if gen != s.gen {
		s.mu.Unlock()
		return false
	}

	s.ring.Push(sample)
	if p := s.pending; p != nil {
		p.add(sample)
	}

	// A ring that waits for a keyframe, or one whose keyframe is getting
	// old, asks for one, no more often than keyframeEvery.
	every := keyframeEvery * s.deps.Second
	if sample.Key {
		s.lastKey, s.hasKey = sample.PTS, true
	}
	ask := (!s.hasKey || s.ring.Waiting() || sample.PTS-s.lastKey >= every) &&
		(!s.hasAsk || sample.PTS-s.lastAsk >= every || sample.PTS < s.lastAsk)
	if ask {
		s.lastAsk, s.hasAsk = sample.PTS, true
	}
	s.mu.Unlock()

	if ask {
		s.deps.RequestKeyframe()
	}
	return true
}

// add collects a sample after the trigger. It stops at maxPostBytes, and at
// a clock that went back, which is a new stream.
func (p *pending) add(sample Sample) {
	if p.postDone {
		return
	}
	var last *Sample
	if n := len(p.post); n > 0 {
		last = &p.post[n-1]
	} else if n := len(p.pre); n > 0 {
		last = &p.pre[n-1]
	}
	switch {
	case last == nil && !sample.Key:
		// Nothing before it to decode it from.
		return
	case last != nil && sample.PTS <= last.PTS:
		p.postDone = true
		return
	}
	cost := sample.cost()
	if p.postBytes+cost > maxPostBytes {
		p.postDone = true
		return
	}
	p.post = append(p.post, sample)
	p.postBytes += cost
}

// Pending says what a trigger did: when the clip it joined started, and when
// it will be written.
type Pending struct {
	Started  time.Time `json:"started"`
	Deadline time.Time `json:"deadline"`
	Triggers []string  `json:"triggers"`
	// Joined is true when the trigger joined a clip that was collecting.
	Joined bool `json:"joined"`
}

// Trigger saves the ring, and the stream for PostSeconds after it. manual
// marks a clip a person asked for. A trigger while a clip collects joins it,
// and pushes its end out by PostSeconds, to at most maxExtend more.
func (s *Service) Trigger(kind string, manual bool) (Pending, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.reader == nil {
		return Pending{}, ErrDisabled
	}
	now := s.deps.Now()
	post := time.Duration(s.settings.PostSeconds) * s.deps.Second

	if p := s.pending; p != nil {
		if !slices.Contains(p.triggers, kind) {
			p.triggers = append(p.triggers, kind)
		}
		p.manual = p.manual || manual
		if deadline := minTime(now.Add(post), p.maxDeadline); deadline.After(p.deadline) {
			p.deadline = deadline
			p.timer.Reset(deadline.Sub(now))
		}
		return p.state(true), nil
	}

	if !s.lastClosed.IsZero() && now.Sub(s.lastClosed) < minGap*s.deps.Second {
		return Pending{}, ErrTooSoon
	}
	if s.ring.Waiting() && (manual || post == 0) {
		return Pending{}, ErrEmpty
	}
	if !manual {
		kept := s.autoTimes[:0]
		for _, at := range s.autoTimes {
			if now.Sub(at) < hour*s.deps.Second {
				kept = append(kept, at)
			}
		}
		s.autoTimes = kept
		if len(s.autoTimes) >= autoPerHour {
			return Pending{}, errTooMany
		}
		s.autoTimes = append(s.autoTimes, now)
	}

	p := &pending{
		at:          now,
		triggers:    []string{kind},
		manual:      manual,
		pre:         s.ring.Snapshot(),
		deadline:    now.Add(post),
		maxDeadline: now.Add(post + maxExtend*s.deps.Second),
	}
	s.pending = p
	s.writes.Add(1)
	p.timer = time.AfterFunc(post, func() { s.finish(p) })
	log.Infof("dashcam: %s: saving %s before and %s after", kind, s.ring.Span().Round(time.Second), post)
	return p.state(false), nil
}

func (p *pending) state(joined bool) Pending {
	return Pending{
		Started:  p.at,
		Deadline: p.deadline,
		Triggers: append([]string(nil), p.triggers...),
		Joined:   joined,
	}
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// finish writes a clip that has collected enough, then keeps the storage cap.
func (s *Service) finish(p *pending) {
	defer s.writes.Done()

	s.mu.Lock()
	if s.pending == p {
		s.pending = nil
	}
	s.lastClosed = s.deps.Now()
	limit := int64(s.settings.StorageMB) << 20
	samples := append(p.pre, p.post...)
	triggers := p.triggers
	p.pre, p.post = nil, nil
	s.mu.Unlock()

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	base := clipBase(p.at, triggers)
	var names []string
	err := ErrEmpty
	if len(samples) > 0 {
		names, err = writeClip(s.deps.Dir, base, samples)
	}
	detail := strings.Join(triggers, ", ")
	if err == nil {
		var removed []string
		removed, err = prune(s.deps.Dir, limit, names)
		if len(removed) > 0 {
			log.Infof("dashcam: the storage cap deleted %s", strings.Join(removed, ", "))
			detail += "; deleted to keep the storage cap: " + strings.Join(removed, ", ")
		}
	}

	s.mu.Lock()
	if err != nil {
		s.lastError = err.Error()
	} else {
		s.lastClip, s.lastError = names, ""
	}
	s.mu.Unlock()

	if err != nil {
		log.Errorf("dashcam: saving %s: %s", base, err)
	} else {
		log.Infof("dashcam: saved %s", strings.Join(names, ", "))
	}

	// A clip a person asked for is in the log already, from the route.
	if p.manual && err == nil {
		return
	}
	e := audit.Event{
		Actor:    "dashcam",
		Via:      "event",
		Category: audit.CategoryMedia,
		Action:   "dashcam-save",
		Target:   strings.Join(names, ", "),
		Outcome:  audit.OutcomeOK,
		Detail:   detail,
	}
	if err != nil {
		e.Outcome, e.Target = audit.OutcomeFail, base+".mp4"
		e.Detail = fmt.Sprintf("%s: %s", detail, err)
	}
	s.deps.Record(e)
}

// onEvent saves a clip on the events the settings name.
func (s *Service) onEvent(ev events.Event) {
	s.mu.Lock()
	triggers := s.settings.Triggers
	s.mu.Unlock()

	var kind string
	switch ev.Type {
	case events.ATXPress:
		if !triggers.ATX {
			return
		}
		kind = TriggerPower
		if ev.Fields["button"] == "reset" {
			kind = TriggerReset
		}
	case events.WatchdogAction:
		if !triggers.ATX {
			return
		}
		kind = TriggerWatchdog
	case events.SignalLost, events.SignalRestored:
		if !triggers.Signal {
			return
		}
		kind = TriggerSignalLost
		if ev.Type == events.SignalRestored {
			kind = TriggerSignalRestored
		}
	case events.FailsafeEntered:
		if !triggers.Failsafe {
			return
		}
		kind = TriggerFailsafe
	default:
		return
	}

	if _, err := s.Trigger(kind, false); err != nil && !errors.Is(err, ErrDisabled) {
		log.Infof("dashcam: %s: no clip: %s", kind, err)
	}
}

// State is what the page shows.
type State struct {
	Enabled bool `json:"enabled"`
	// Running is false while the dashcam is on but the board captures no
	// video (failsafe mode).
	Running bool `json:"running"`
	Seconds int  `json:"seconds"`
	// Buffered is how many seconds the ring holds now, and BufferedBytes
	// their size; BudgetBytes is the ring's bound.
	Buffered      float64  `json:"buffered"`
	BufferedBytes int      `json:"bufferedBytes"`
	BudgetBytes   int      `json:"budgetBytes"`
	Pending       *Pending `json:"pending,omitempty"`
	LastClip      []string `json:"lastClip,omitempty"`
	LastError     string   `json:"lastError,omitempty"`
}

// State answers the state.
func (s *Service) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := State{
		Enabled:       s.settings.Enabled,
		Running:       s.reader != nil,
		Seconds:       s.settings.Seconds,
		Buffered:      s.ring.Span().Seconds() * float64(time.Second) / float64(s.deps.Second),
		BufferedBytes: s.ring.Bytes(),
		BudgetBytes:   s.budget(s.settings),
		LastClip:      s.lastClip,
		LastError:     s.lastError,
	}
	if s.pending != nil {
		p := s.pending.state(false)
		st.Pending = &p
	}
	return st
}
