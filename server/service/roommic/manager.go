package roommic

import (
	"errors"
	"slices"
	"sync"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/service/stream/audio"

	log "github.com/sirupsen/logrus"
)

var (
	// ErrUnavailable means this kernel has no onboard microphone card.
	ErrUnavailable = errors.New("the room microphone is not available on this device")
	// ErrNotAllowed means an administrator has not allowed the microphone.
	ErrNotAllowed = errors.New("the room microphone is not allowed on this device")
)

// Format is the microphone as the hardware gives it: 48 kHz, one channel.
// 32 kbit/s is plenty for one channel of a room.
var Format = audio.Format{
	Device:   Device,
	Channels: 1,
	Bitrate:  32000,
	Name:     "room microphone",
}

// Status is what every viewer is told.
type Status struct {
	// Available is false on a kernel without the onboard card.
	Available bool `json:"available"`
	// Allowed is the administrator's setting.
	Allowed bool `json:"allowed"`
	// Gain is the PGA setting the next capture starts with.
	Gain int `json:"gain"`
	// Live is true while the microphone is open, whoever opened it.
	Live bool `json:"live"`
	// ListenerCount is how many accounts are listening.
	ListenerCount int `json:"listenerCount"`
	// Listeners names the accounts listening, each once. Only administrators
	// are told who listens: ForViewer leaves it out for everyone else.
	Listeners []string `json:"listeners,omitempty"`
}

// ForViewer returns the status as a viewer with the given role may see it.
// The listeners' names go to administrators only; everyone else keeps the
// live flag and the count.
func (s Status) ForViewer(role authn.Role) Status {
	if role != authn.RoleAdmin {
		s.Listeners = nil
	}

	return s
}

// Manager counts the viewers who have the microphone on, opens it for the
// first and closes it after the last.
type Manager struct {
	hub       *audio.Hub
	available func() bool

	mutex     sync.Mutex
	settings  Settings
	loaded    bool
	listeners map[*Listener]struct{}
	watchers  map[uint64]func()
	nextWatch uint64
}

// Shared is the one microphone every transport opens through.
var Shared = NewManager(Available, func() audio.Capture {
	filter := NewHighPass(audio.SampleRate, HighPassCutoff)
	return audio.NewStreamFor(Format, filter.Process)
})

// NewManager builds a manager around a card check and a capture factory, so a
// test can supply both. The manager sets the gain before each capture starts.
func NewManager(available func() bool, newCapture func() audio.Capture) *Manager {
	m := &Manager{
		available: available,
		listeners: make(map[*Listener]struct{}),
		watchers:  make(map[uint64]func()),
	}

	m.hub = audio.NewHubWith(func() audio.Capture {
		return &gainCapture{Capture: newCapture(), gain: m.Settings().Gain}
	}, m.captureAllowed)

	return m
}

// captureAllowed is the hub's last check before it starts arecord.
func (m *Manager) captureAllowed() bool {
	return m.available() && m.Settings().Allowed
}

// Settings returns the administrator's settings, reading the file the first
// time.
func (m *Manager) Settings() Settings {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.settingsLocked()
}

func (m *Manager) settingsLocked() Settings {
	if !m.loaded {
		settings, err := loadSettings()
		if err != nil {
			log.Warnf("room microphone: cannot read %s, keeping it off: %s", SettingsFile, err)
		}
		m.settings = settings
		m.loaded = true
	}

	return m.settings
}

// Status reports the device's state for a viewer.
func (m *Manager) Status() Status {
	available := m.available()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	settings := m.settingsLocked()
	names := make([]string, 0, len(m.listeners))
	for listener := range m.listeners {
		if !slices.Contains(names, listener.user) {
			names = append(names, listener.user)
		}
	}
	slices.Sort(names)

	return Status{
		Available:     available,
		Allowed:       settings.Allowed,
		Gain:          settings.Gain,
		Live:          len(m.listeners) > 0,
		ListenerCount: len(names),
		Listeners:     names,
	}
}

// Live reports whether the microphone is open.
func (m *Manager) Live() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return len(m.listeners) > 0
}

// SetSettings saves an administrator's change and applies it: a gain change
// reaches a live microphone at once, and disallowing closes it for everyone.
func (m *Manager) SetSettings(settings Settings, by string) error {
	if err := settings.validate(); err != nil {
		return err
	}

	m.mutex.Lock()
	previous := m.settingsLocked()
	if err := saveSettings(settings); err != nil {
		m.mutex.Unlock()
		return err
	}
	m.settings = settings
	live := len(m.listeners) > 0
	m.mutex.Unlock()

	if previous.Allowed != settings.Allowed {
		if settings.Allowed {
			log.Infof("room microphone: %s allowed it on this device", by)
		} else {
			log.Infof("room microphone: %s disallowed it on this device", by)
		}
	}

	if previous.Gain != settings.Gain {
		log.Infof("room microphone: %s set the gain to %d", by, settings.Gain)
		if live && settings.Allowed {
			if err := applyGain(settings.Gain); err != nil {
				log.Warnf("room microphone: cannot set the gain: %s", err)
			}
		}
	}

	if !settings.Allowed {
		m.CloseAll("it was disallowed by " + by)
	}

	m.notify()

	return nil
}

// Open starts listening for one viewer, opening the microphone if nobody else
// has it on. user and via are for the log: who, and over which transport.
func (m *Manager) Open(user string, via string) (*Listener, error) {
	if !m.available() {
		return nil, ErrUnavailable
	}
	if !m.Settings().Allowed {
		return nil, ErrNotAllowed
	}

	sub := m.hub.Subscribe()
	if sub == nil {
		return nil, ErrNotAllowed
	}

	listener := &Listener{manager: m, sub: sub, user: user, via: via}

	m.mutex.Lock()
	m.listeners[listener] = struct{}{}
	count := len(m.listeners)
	m.mutex.Unlock()

	if count == 1 {
		log.Infof("room microphone opened: %s switched it on (%s)", user, via)
	} else {
		log.Infof("room microphone: %s is listening too (%s), %d listeners", user, via, count)
	}

	m.notify()

	return listener, nil
}

// CloseAll ends every listener, which closes the microphone.
func (m *Manager) CloseAll(reason string) {
	m.mutex.Lock()
	listeners := make([]*Listener, 0, len(m.listeners))
	for listener := range m.listeners {
		listeners = append(listeners, listener)
	}
	m.mutex.Unlock()

	for _, listener := range listeners {
		listener.close(reason)
	}
}

// Watch calls fn after every change of Status that viewers see: the microphone
// opening or closing, a listener joining or leaving, a settings change. fn
// runs on the goroutine that made the change and must not block for long. The
// returned function stops the calls.
func (m *Manager) Watch(fn func()) func() {
	m.mutex.Lock()
	id := m.nextWatch
	m.nextWatch++
	m.watchers[id] = fn
	m.mutex.Unlock()

	return func() {
		m.mutex.Lock()
		delete(m.watchers, id)
		m.mutex.Unlock()
	}
}

func (m *Manager) notify() {
	m.mutex.Lock()
	watchers := make([]func(), 0, len(m.watchers))
	for _, fn := range m.watchers {
		watchers = append(watchers, fn)
	}
	m.mutex.Unlock()

	for _, fn := range watchers {
		fn()
	}
}

// Listener is one viewer's hold on the microphone.
type Listener struct {
	manager *Manager
	sub     *audio.Subscription
	user    string
	via     string
	once    sync.Once
}

// Frames delivers the microphone's Opus frames. It closes when the listener
// closes, or when capture ends by itself; the transport then calls Close.
func (l *Listener) Frames() <-chan audio.Frame {
	return l.sub.Frames()
}

// Close stops listening, and closes the microphone if this was the last
// listener. It is safe to call more than once.
func (l *Listener) Close() {
	l.close("")
}

func (l *Listener) close(reason string) {
	l.once.Do(func() {
		// The subscription goes first, so the microphone is closed by the
		// time the log and the viewers hear that it is.
		l.sub.Close()

		m := l.manager
		m.mutex.Lock()
		delete(m.listeners, l)
		count := len(m.listeners)
		m.mutex.Unlock()

		switch {
		case count == 0 && reason != "":
			log.Infof("room microphone closed: %s; %s was listening (%s)", reason, l.user, l.via)
		case count == 0:
			log.Infof("room microphone closed: %s, the last listener, switched it off (%s)", l.user, l.via)
		default:
			log.Infof("room microphone: %s stopped listening (%s), %d listeners left", l.user, l.via, count)
		}

		m.notify()
	})
}

// gainCapture sets the PGA before the capture opens the device. The codec
// boots at gain 0, where the room is at the ADC's floor.
type gainCapture struct {
	audio.Capture
	gain int
}

func (g *gainCapture) Start() {
	if err := applyGain(g.gain); err != nil {
		log.Warnf("room microphone: cannot set the gain to %d: %s", g.gain, err)
	}

	g.Capture.Start()
}

// SetStateHandler passes the hub's state handler through, so the wrapper does
// not hide the capture's state reports.
func (g *gainCapture) SetStateHandler(fn func(audio.State)) {
	if reporter, ok := g.Capture.(interface{ SetStateHandler(func(audio.State)) }); ok {
		reporter.SetStateHandler(fn)
	}
}
