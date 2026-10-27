// Package hostlog receives the managed host's logs over UDP: syslog (RFC 3164
// and 5424) and netconsole. The lines go into a size-capped ring in memory and,
// if the owner asks, a size-capped log on /data. A page shows them live, and a
// line that matches a phrase the owner chose is published on the event bus.
// Both listeners are off by default, and only the sources the owner allows are
// heard.
package hostlog

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/events"
)

// Rate limits. A host in a panic loop can send thousands of lines a second;
// the board has one slow core and an SD card.
const (
	maxRate  = 500  // lines per second, sustained
	maxBurst = 1000 // lines
	// matchEvery is the least time between two phrase events. The notifier
	// has its own cooldown; this keeps a flood off the bus.
	matchEvery = time.Second

	readBuffer = 8192
	// reconcileEvery is how often the service checks that the listeners are
	// what the settings say. It is also how soon a USB link that came up is
	// bound.
	reconcileEvery = 10 * time.Second
	flushEvery     = 2 * time.Second
)

// Deps is everything the service reaches outside this package. The router
// wires the real ones; tests pass fakes.
type Deps struct {
	// Settings returns the current settings, defaults filled in.
	Settings func() config.HostLog
	// SetSettings saves the settings and applies them to the running server.
	SetSettings func(config.HostLog) error

	// DefaultSources are the addresses heard when the owner names none: the
	// host on the USB link, else the watchdog's ping host. Empty means any.
	DefaultSources func() []netip.Prefix
	// USBBoard is the board's address on the USB link, when the link is up.
	USBBoard func() (netip.Addr, bool)

	// Publish puts an event on the bus. Nil drops it.
	Publish func(events.Event)

	// DiskDir is where the optional log lives. Empty means DiskDir.
	DiskDir string

	Now func() time.Time
	// Listen opens a UDP socket. Nil means net.ListenPacket.
	Listen func(network, address string) (net.PacketConn, error)
}

type listener struct {
	proto string
	addr  string
	conn  net.PacketConn
}

// Service is the receiver.
type Service struct {
	deps Deps
	ring *Ring

	startOnce sync.Once

	mu        sync.Mutex
	disk      *Disk
	diskOn    bool
	listeners map[string]*listener
	errs      map[string]string
	waiting   map[string]bool
	sources   []netip.Prefix // the filter in force; empty means any
	patterns  []*regexp.Regexp
	patSrc    []string
	lastMatch time.Time

	// limit keeps a flood to maxRate lines a second, after a burst of
	// maxBurst. It is read with the time Deps.Now gives, so tests can move it.
	limit *rate.Limiter

	received    atomic.Uint64
	filtered    atomic.Uint64
	rateLimited atomic.Uint64
}

// New returns a service. Start runs it.
func New(deps Deps) *Service {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.Listen == nil {
		deps.Listen = net.ListenPacket
	}
	if deps.DiskDir == "" {
		deps.DiskDir = DiskDir
	}
	set := deps.Settings()
	return &Service{
		deps:      deps,
		ring:      NewRing(set.MemoryKB * 1024),
		listeners: map[string]*listener{},
		errs:      map[string]string{},
		waiting:   map[string]bool{},
		limit:     rate.NewLimiter(maxRate, maxBurst),
	}
}

// Start applies the settings and keeps them applied for the rest of uptime.
// Calling it again does nothing.
func (s *Service) Start() {
	s.startOnce.Do(func() {
		s.Reconcile()
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("hostlog panicked, no longer receiving: %v", r)
				}
			}()
			reconcile := time.NewTicker(reconcileEvery)
			flush := time.NewTicker(flushEvery)
			defer reconcile.Stop()
			defer flush.Stop()
			for {
				select {
				case <-reconcile.C:
					s.Reconcile()
				case <-flush.C:
					s.mu.Lock()
					d := s.disk
					s.mu.Unlock()
					if d != nil {
						if err := d.Flush(); err != nil {
							log.Warnf("hostlog: write the disk log: %s", err)
						}
					}
				}
			}
		}()
	})
}

// patternsOf compiles the phrases. A bad one is reported by name, never
// skipped silently.
func patternsOf(src []string) ([]*regexp.Regexp, error) {
	out := make([]*regexp.Regexp, 0, len(src))
	for _, p := range src {
		re, err := regexp.Compile("(?i)" + p)
		if err != nil {
			return nil, fmt.Errorf("the phrase %q is not a valid pattern", p)
		}
		out = append(out, re)
	}
	return out, nil
}

// parseSources reads the addresses and networks a filter lists.
func parseSources(list []string) ([]netip.Prefix, error) {
	out := make([]netip.Prefix, 0, len(list))
	for _, item := range list {
		if p, err := netip.ParsePrefix(item); err == nil {
			out = append(out, p.Masked())
			continue
		}
		a, err := netip.ParseAddr(item)
		if err != nil {
			return nil, fmt.Errorf("%q is not an address or network", item)
		}
		out = append(out, netip.PrefixFrom(a.Unmap(), a.Unmap().BitLen()))
	}
	return out, nil
}

// listenAddr is where a listener binds, or "" with a reason when it cannot yet.
func (s *Service) listenAddr(set config.HostLog, port int) (addr string, waiting bool) {
	switch set.Bind {
	case "usb":
		board, ok := s.deps.USBBoard()
		if !ok {
			return "", true
		}
		return net.JoinHostPort(board.String(), strconv.Itoa(port)), false
	default:
		return ":" + strconv.Itoa(port), false
	}
}

// Reconcile makes the listeners, the ring, the disk log and the filter what
// the settings say. It is safe to call at any time.
func (s *Service) Reconcile() {
	set := s.deps.Settings()

	s.ring.SetCap(set.MemoryKB * 1024)

	explicit, _ := parseSources(set.Sources)
	if len(explicit) == 0 {
		explicit = s.deps.DefaultSources()
	}
	pats, perr := patternsOf(set.Patterns)
	if perr != nil {
		log.Warnf("hostlog: %s", perr)
		pats = nil
	}

	s.mu.Lock()
	s.sources = explicit
	s.patterns, s.patSrc = pats, set.Patterns
	s.reconcileDiskLocked(set)
	s.mu.Unlock()

	for _, l := range []struct {
		proto string
		cfg   config.HostLogListener
	}{
		{ProtoSyslog, set.Syslog},
		{ProtoNetconsole, set.Netconsole},
	} {
		s.reconcileListener(set, l.proto, l.cfg)
	}
}

func (s *Service) reconcileDiskLocked(set config.HostLog) {
	switch {
	case set.Disk && !s.diskOn:
		s.disk = NewDisk(s.deps.DiskDir, int64(set.DiskKB)*1024)
		s.diskOn = true
		// The last run's lines come back first.
		if s.ring.hasNothing() {
			s.ring.Load(s.disk.ReadAll())
		}
	case set.Disk:
		s.disk.SetMax(int64(set.DiskKB) * 1024)
	case s.diskOn:
		_ = s.disk.Close()
		s.disk, s.diskOn = nil, false
	}
}

func (r *Ring) hasNothing() bool {
	n, _ := r.Stats()
	return n == 0
}

func (s *Service) reconcileListener(set config.HostLog, proto string, cfg config.HostLogListener) {
	want, waiting := "", false
	if cfg.Enabled {
		want, waiting = s.listenAddr(set, cfg.Port)
	}

	s.mu.Lock()
	cur := s.listeners[proto]
	if cur != nil && cur.addr == want {
		s.mu.Unlock()
		return
	}
	if cur != nil {
		delete(s.listeners, proto)
		_ = cur.conn.Close()
	}
	prevErr := s.errs[proto]
	delete(s.errs, proto)
	s.waiting[proto] = cfg.Enabled && waiting
	s.mu.Unlock()

	if want == "" {
		return
	}

	conn, err := s.deps.Listen("udp", want)
	if err != nil {
		// The same failure is retried every few seconds; say it once.
		if err.Error() != prevErr {
			log.Warnf("hostlog: listen for %s on %s: %s", proto, want, err)
		}
		s.mu.Lock()
		s.errs[proto] = err.Error()
		s.mu.Unlock()
		return
	}

	l := &listener{proto: proto, addr: want, conn: conn}
	s.mu.Lock()
	s.listeners[proto] = l
	s.mu.Unlock()
	log.Infof("hostlog: receiving %s on %s", proto, want)
	go s.serve(l)
}

func (s *Service) serve(l *listener) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("hostlog: the %s receiver panicked: %v", l.proto, r)
		}
	}()

	buf := make([]byte, readBuffer)
	for {
		n, from, err := l.conn.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Warnf("hostlog: read %s: %s", l.proto, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		var ip netip.Addr
		if ua, ok := from.(*net.UDPAddr); ok {
			ip, _ = netip.AddrFromSlice(ua.IP)
			ip = ip.Unmap()
		}
		s.Ingest(l.proto, ip, buf[:n])
	}
}

// Allowed reports whether a sender may be heard.
func (s *Service) allowed(ip netip.Addr) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.sources) == 0 {
		return true
	}
	for _, p := range s.sources {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}

// Ingest handles one datagram from a sender.
func (s *Service) Ingest(proto string, from netip.Addr, data []byte) {
	if !s.allowed(from) {
		s.filtered.Add(1)
		return
	}

	var msgs []parsed
	if proto == ProtoNetconsole {
		msgs = parseNetconsole(data)
	} else {
		if p := parseSyslog(data); p.msg != "" {
			msgs = []parsed{p}
		}
	}

	now := s.deps.Now()
	for _, p := range msgs {
		s.received.Add(1)
		if !s.limit.AllowN(now, 1) {
			s.rateLimited.Add(1)
			continue
		}
		e := s.ring.Add(Entry{
			Time:     now,
			Source:   from.String(),
			Proto:    proto,
			Facility: p.facility,
			Severity: p.severity,
			Host:     p.host,
			Tag:      p.tag,
			Msg:      p.msg,
		})

		s.mu.Lock()
		d := s.disk
		s.mu.Unlock()
		if d != nil {
			if err := d.Append(e); err != nil {
				log.Warnf("hostlog: write the disk log: %s", err)
			}
		}
		s.match(e, now)
	}
}

// match publishes a phrase event when the line holds a phrase.
func (s *Service) match(e Entry, now time.Time) {
	if s.deps.Publish == nil {
		return
	}

	s.mu.Lock()
	pats, src := s.patterns, s.patSrc
	s.mu.Unlock()

	for i, re := range pats {
		if !re.MatchString(e.Msg) {
			continue
		}
		s.mu.Lock()
		due := now.Sub(s.lastMatch) >= matchEvery
		if due {
			s.lastMatch = now
		}
		s.mu.Unlock()
		if !due {
			return
		}

		msg := e.Msg
		if len(msg) > 300 {
			msg = msg[:300] + "..."
		}
		s.deps.Publish(events.Event{
			Type:    events.HostLogMatch,
			Time:    now,
			Title:   "Host log: " + src[i],
			Message: msg,
			Fields:  map[string]string{"source": e.Source, "proto": e.Proto},
		})
		return
	}
}

// Ring exposes the ring to the handlers.
func (s *Service) Ring() *Ring { return s.ring }

// Clear forgets the log, in memory and on disk.
func (s *Service) Clear() error {
	s.ring.Clear()

	s.mu.Lock()
	d := s.disk
	s.mu.Unlock()
	if d == nil {
		d = NewDisk(s.deps.DiskDir, 0)
	}
	return d.Clear()
}

// ListenerState is what one listener is doing, for the page.
type ListenerState struct {
	// State is "off", "listening", "waiting" (for the USB link) or "error".
	State string `json:"state"`
	Addr  string `json:"addr,omitempty"`
	Error string `json:"error,omitempty"`
}

// State is what the page reads besides the settings.
type State struct {
	Syslog     ListenerState `json:"syslog"`
	Netconsole ListenerState `json:"netconsole"`
	// Sources is the filter in force; empty means any sender is heard.
	Sources []string `json:"sources"`
	// Received counts lines accepted from an allowed sender; Filtered counts
	// datagrams from a sender that is not; RateLimited counts lines dropped
	// for coming too fast.
	Received    uint64 `json:"received"`
	Filtered    uint64 `json:"filtered"`
	RateLimited uint64 `json:"rateLimited"`
	Buffered    int    `json:"buffered"`
	BufferBytes int    `json:"bufferBytes"`
	DiskBytes   int64  `json:"diskBytes"`
}

// State reports the listeners and the counters.
func (s *Service) State() State {
	set := s.deps.Settings()

	s.mu.Lock()
	st := State{
		Syslog:     s.listenerStateLocked(ProtoSyslog, set.Syslog.Enabled),
		Netconsole: s.listenerStateLocked(ProtoNetconsole, set.Netconsole.Enabled),
		Sources:    []string{},
	}
	for _, p := range s.sources {
		st.Sources = append(st.Sources, p.String())
	}
	d := s.disk
	s.mu.Unlock()

	st.Received = s.received.Load()
	st.Filtered = s.filtered.Load()
	st.RateLimited = s.rateLimited.Load()
	st.Buffered, st.BufferBytes = s.ring.Stats()
	if d != nil {
		st.DiskBytes = d.Size()
	}
	return st
}

func (s *Service) listenerStateLocked(proto string, enabled bool) ListenerState {
	if !enabled {
		return ListenerState{State: "off"}
	}
	if l := s.listeners[proto]; l != nil {
		return ListenerState{State: "listening", Addr: l.addr}
	}
	if msg := s.errs[proto]; msg != "" {
		return ListenerState{State: "error", Error: msg}
	}
	if s.waiting[proto] {
		return ListenerState{State: "waiting"}
	}
	return ListenerState{State: "off"}
}
