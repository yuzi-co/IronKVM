// Package notify sends the events on the server's bus to the owner: a generic
// webhook, ntfy, and a Telegram bot, with a JPEG snapshot where the channel
// can carry one. It is off by default, rate limits each event type, and keeps
// its tokens in a 0600 file that it never logs.
package notify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/service/events"
)

// Deps is everything the service reaches outside this package. The router
// wires the real ones; tests pass fakes.
type Deps struct {
	Store *Store
	Bus   *events.Bus

	// Snapshot returns a JPEG of the screen.
	Snapshot func(ctx context.Context) ([]byte, error)
	// Signal reports whether the HDMI signal is present. ok is false when
	// that cannot be told, as with capture switched off.
	Signal func() (present, ok bool)
	// Client returns the HTTP client for the outside world, which honours
	// the configured proxy. Nil means a plain client.
	Client func() *http.Client
	// Device names this board in a message.
	Device func() string

	Now func() time.Time
	// SignalInterval is how often the signal is read. Zero means 5s.
	SignalInterval time.Duration
}

// recentMax is how many events the page remembers.
const recentMax = 20

// Entry is one event the service handled, for the page.
type Entry struct {
	Time       time.Time `json:"time"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Suppressed bool      `json:"suppressed,omitempty"`
	// Channels says what each channel answered: "ok" or the reason it failed.
	Channels map[string]string `json:"channels,omitempty"`
}

type limit struct {
	last       time.Time
	suppressed int
}

// Service is the notifier.
type Service struct {
	deps Deps

	startOnce sync.Once

	mu     sync.Mutex
	limits map[string]*limit
	logins []time.Time
	recent []Entry
	// lastResult is the last answer of each channel.
	lastResult map[string]ChannelResult
}

// ChannelResult is the last attempt on one channel.
type ChannelResult struct {
	Time  time.Time `json:"time"`
	OK    bool      `json:"ok"`
	Error string    `json:"error,omitempty"`
}

// New returns a service. Start runs it.
func New(deps Deps) *Service {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.Bus == nil {
		deps.Bus = events.Default
	}
	if deps.Client == nil {
		deps.Client = func() *http.Client { return &http.Client{Timeout: sendTimeout} }
	}
	if deps.Device == nil {
		deps.Device = func() string { return "IronKVM" }
	}
	if deps.SignalInterval <= 0 {
		deps.SignalInterval = 5 * time.Second
	}
	return &Service{
		deps:       deps,
		limits:     map[string]*limit{},
		lastResult: map[string]ChannelResult{},
	}
}

// Start subscribes to the bus and watches the HDMI signal for the rest of
// uptime. Calling it again does nothing.
func (s *Service) Start() {
	s.startOnce.Do(func() { s.run() })
}

// run is Start's body, split out so tests can stop it.
func (s *Service) run() (stop func()) {
	cancel := s.deps.Bus.Subscribe(s.handle)
	done := make(chan struct{})
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("notify: the signal monitor panicked and stopped: %v", r)
			}
		}()
		s.watchSignal(done)
	}()
	return func() {
		close(done)
		cancel()
	}
}

// handle takes one event off the bus.
func (s *Service) handle(ev events.Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("notify: handling %s panicked: %v", ev.Type, r)
		}
	}()

	set, err := s.deps.Store.Load()
	if err != nil {
		log.Warnf("notify: settings: %s", err)
		return
	}
	if !set.Enabled {
		return
	}

	if ev.Type == events.LoginFailure {
		var raised bool
		if ev, raised = s.countLogin(set, ev); !raised {
			return
		}
	}

	rule, ok := set.Events[ev.Type]
	if !ok || !rule.Enabled {
		return
	}

	now := s.deps.Now()
	suppressed, allowed := s.allow(ev.Type, now, time.Duration(rule.CooldownSeconds)*time.Second)
	if !allowed {
		s.remember(Entry{Time: now, Type: ev.Type, Title: ev.Title, Suppressed: true})
		return
	}
	if suppressed > 0 {
		ev = withField(ev, "suppressed", strconv.Itoa(suppressed))
	}

	s.deliver(set, ev, true)
}

func withField(ev events.Event, k, v string) events.Event {
	fields := make(map[string]string, len(ev.Fields)+1)
	for fk, fv := range ev.Fields {
		fields[fk] = fv
	}
	fields[k] = v
	ev.Fields = fields
	return ev
}

// allow applies the cooldown to one event type. It returns how many events the
// cooldown swallowed since the last one sent.
func (s *Service) allow(t string, now time.Time, cooldown time.Duration) (suppressed int, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	l := s.limits[t]
	if l == nil {
		l = &limit{}
		s.limits[t] = l
	}
	if !l.last.IsZero() && now.Sub(l.last) < cooldown {
		l.suppressed++
		return 0, false
	}
	suppressed = l.suppressed
	l.last, l.suppressed = now, 0
	return suppressed, true
}

// countLogin folds one failed login into the window and reports whether the
// run reached the threshold, with the event to raise.
func (s *Service) countLogin(set Settings, ev events.Event) (events.Event, bool) {
	now := s.deps.Now()
	window := time.Duration(set.LoginWindowSeconds) * time.Second

	s.mu.Lock()
	defer s.mu.Unlock()

	kept := s.logins[:0]
	for _, t := range s.logins {
		if now.Sub(t) < window {
			kept = append(kept, t)
		}
	}
	s.logins = append(kept, now)
	if len(s.logins) < set.LoginThreshold {
		return ev, false
	}
	n := len(s.logins)
	s.logins = nil

	raised := events.Event{
		Type:       events.LoginFailures,
		Time:       now,
		Title:      fmt.Sprintf("%d failed logins in %s", n, window.Round(time.Second)),
		Message:    "Someone is trying to log in to the web UI or an API.",
		Fields:     map[string]string{"count": strconv.Itoa(n)},
		NoSnapshot: true,
	}
	if src := ev.Fields["source"]; src != "" {
		raised.Fields["lastSource"] = src
	}
	return raised, true
}

// remember adds an entry to the page's list.
func (s *Service) remember(e Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.recent = append([]Entry{e}, s.recent...)
	if len(s.recent) > recentMax {
		s.recent = s.recent[:recentMax]
	}
}

// deliver sends the event on every channel that is switched on. snapshot says
// whether this event may carry a picture; a test passes what it wants.
func (s *Service) deliver(set Settings, ev events.Event, snapshot bool) map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 4*sendTimeout)
	defer cancel()

	var jpeg []byte
	if snapshot && set.Snapshot && !ev.NoSnapshot {
		jpeg = ev.JPEG
		if len(jpeg) == 0 && s.deps.Snapshot != nil {
			shotCtx, shotCancel := context.WithTimeout(ctx, 5*time.Second)
			shot, err := s.deps.Snapshot(shotCtx)
			shotCancel()
			if err != nil {
				log.Warnf("notify: no snapshot for %s: %s", ev.Type, err)
			} else {
				jpeg = shot
			}
		}
	}

	m := message{ev: ev, device: s.deps.Device(), jpeg: jpeg}
	m.ev.Title = m.device + ": " + ev.Title

	results := map[string]string{}
	client := s.deps.Client()
	send := func(name string, on bool, fn func() error) {
		if !on {
			return
		}
		err := fn()
		s.setResult(name, err)
		if err != nil {
			log.Warnf("notify: %s for %s: %s", name, ev.Type, err)
			results[name] = err.Error()
			return
		}
		results[name] = "ok"
	}
	send(ChannelWebhook, set.Webhook.Enabled, func() error {
		return sendWebhook(ctx, client, set.Webhook, m)
	})
	send(ChannelNtfy, set.Ntfy.Enabled, func() error {
		return sendNtfy(ctx, client, set.Ntfy, m)
	})
	send(ChannelTelegram, set.Telegram.Enabled, func() error {
		return sendTelegram(ctx, client, set.Telegram, m)
	})

	s.remember(Entry{Time: s.deps.Now(), Type: ev.Type, Title: ev.Title, Channels: results})
	return results
}

func (s *Service) setResult(name string, err error) {
	r := ChannelResult{Time: s.deps.Now(), OK: err == nil}
	if err != nil {
		r.Error = err.Error()
	}
	s.mu.Lock()
	s.lastResult[name] = r
	s.mu.Unlock()
}

// ErrNoChannel is what Test returns when the channel asked for is not set up.
var ErrNoChannel = errors.New("that channel is not set up")

// Test sends a test message now, ignoring the master switch, the event rules
// and the cooldowns, on one channel or, when channel is empty, on every
// channel that is set up. It returns what each channel answered.
func (s *Service) Test(channel string) (map[string]string, error) {
	set, err := s.deps.Store.Load()
	if err != nil {
		return nil, err
	}
	switch channel {
	case "":
		set.Webhook.Enabled = set.Webhook.URL != ""
		set.Ntfy.Enabled = set.Ntfy.Topic != ""
		set.Telegram.Enabled = set.Telegram.BotToken != "" && set.Telegram.ChatID != ""
	case ChannelWebhook:
		set.Webhook.Enabled = set.Webhook.URL != ""
		set.Ntfy.Enabled, set.Telegram.Enabled = false, false
	case ChannelNtfy:
		set.Ntfy.Enabled = set.Ntfy.Topic != ""
		set.Webhook.Enabled, set.Telegram.Enabled = false, false
	case ChannelTelegram:
		set.Telegram.Enabled = set.Telegram.BotToken != "" && set.Telegram.ChatID != ""
		set.Webhook.Enabled, set.Ntfy.Enabled = false, false
	default:
		return nil, ErrNoChannel
	}
	if !set.Webhook.Enabled && !set.Ntfy.Enabled && !set.Telegram.Enabled {
		return nil, ErrNoChannel
	}

	ev := events.Event{
		Type:    "test",
		Time:    s.deps.Now(),
		Title:   "Test notification",
		Message: "If you can read this, the channel works.",
	}
	return s.deliver(set, ev, true), nil
}

// State is what the page shows besides the settings.
type State struct {
	Channels map[string]ChannelResult `json:"channels"`
	Recent   []Entry                  `json:"recent"`
}

// State returns the last answer of each channel and the recent events.
func (s *Service) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := State{
		Channels: make(map[string]ChannelResult, len(s.lastResult)),
		Recent:   append([]Entry{}, s.recent...),
	}
	for k, v := range s.lastResult {
		st.Channels[k] = v
	}
	return st
}

// watchSignal reads the HDMI signal and publishes a change once it has held
// for two reads, so a flicker during a mode change says nothing. The first
// reading is the baseline: it raises nothing.
func (s *Service) watchSignal(done <-chan struct{}) {
	ticker := time.NewTicker(s.deps.SignalInterval)
	defer ticker.Stop()

	var (
		known, state bool
		pending      bool
		havePending  bool
	)
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
		}

		if s.deps.Signal == nil {
			continue
		}
		set, err := s.deps.Store.Load()
		if err != nil || !set.Enabled || !(set.Events[events.SignalLost].Enabled || set.Events[events.SignalRestored].Enabled) {
			known, havePending = false, false
			continue
		}
		present, ok := s.deps.Signal()
		if !ok {
			known, havePending = false, false
			continue
		}
		if !known {
			known, state, havePending = true, present, false
			continue
		}
		if present == state {
			havePending = false
			continue
		}
		if !havePending || pending != present {
			pending, havePending = present, true
			continue
		}

		state, havePending = present, false
		if present {
			s.deps.Bus.Publish(events.Event{
				Type:  events.SignalRestored,
				Title: "HDMI signal restored",
			})
		} else {
			s.deps.Bus.Publish(events.Event{
				Type:       events.SignalLost,
				Title:      "HDMI signal lost",
				Message:    "The host stopped sending video.",
				NoSnapshot: true,
			})
		}
	}
}
