package metrics

import (
	"context"
	"sync"
	"time"

	"NanoKVM-Server/service/extensions/netbird"
	"NanoKVM-Server/service/extensions/tailscale"
)

const helpVpn = "Whether an installed VPN is connected. Its CLI is asked once a minute in the background."

// vpnProvider is one VPN add-on. installed and running are a stat and a pid
// file, cheap enough for every scrape. connected runs the CLI, so only the
// background refresher calls it.
type vpnProvider struct {
	name      string
	installed func() bool
	running   func() bool
	connected func(context.Context) bool
}

var vpnProviders = []vpnProvider{
	{name: "tailscale", installed: tailscale.Installed, running: tailscale.Running, connected: tailscale.Connected},
	{name: "netbird", installed: netbird.Installed, running: netbird.Running, connected: netbird.Connected},
}

const (
	vpnRefreshInterval = time.Minute
	// vpnCheckTimeout bounds one CLI call. A CLI that does not answer in time
	// counts as not connected.
	vpnCheckTimeout = 5 * time.Second
)

// vpnState keeps the last answer of each provider's CLI and the goroutine
// that renews them. A scrape only reads the answers: it never runs a CLI, so
// a slow or wedged CLI cannot slow a scrape down.
//
// The refresher starts on the first scrape, so a board nobody scrapes never
// runs the CLIs for this.
type vpnState struct {
	interval time.Duration

	mu      sync.Mutex
	answers map[string]bool

	start sync.Once
	kick  chan struct{}
	stop  context.CancelFunc
	done  chan struct{}
}

func newVpnState(interval time.Duration) *vpnState {
	return &vpnState{
		interval: interval,
		answers:  map[string]bool{},
		kick:     make(chan struct{}, 1),
	}
}

var vpn = newVpnState(vpnRefreshInterval)

// StopVpnRefresher ends the background refresher, cancelling a CLI call in
// flight, and waits for it. It is safe to call when the refresher never
// started, and more than once.
func StopVpnRefresher() {
	vpn.shutdown()
}

// collectVpn writes a series for each installed provider only. A daemon that
// is not running is not connected, which needs no CLI; it also forgets the
// kept answer, so a daemon that starts is not served an old one. A running
// daemon with no answer yet is left out until the refresher has asked it, and
// the refresher is told to ask now rather than at its next tick.
func collectVpn(w *Writer) {
	vpn.ensureStarted()

	vpn.mu.Lock()
	defer vpn.mu.Unlock()

	for _, p := range vpnProviders {
		if !p.installed() {
			delete(vpn.answers, p.name)
			continue
		}

		if !p.running() {
			delete(vpn.answers, p.name)
			w.Gauge("ironkvm_vpn_connected", helpVpn, 0, L("provider", p.name))
			continue
		}

		connected, ok := vpn.answers[p.name]
		if !ok {
			vpn.wake()
			continue
		}
		w.Gauge("ironkvm_vpn_connected", helpVpn, boolValue(connected), L("provider", p.name))
	}
}

func (s *vpnState) ensureStarted() {
	s.start.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		s.stop = cancel
		s.done = make(chan struct{})
		go s.loop(ctx)
	})
}

// wake asks the refresher for a pass now. It never blocks: one pending
// request is as good as several.
func (s *vpnState) wake() {
	select {
	case s.kick <- struct{}{}:
	default:
	}
}

func (s *vpnState) shutdown() {
	// Once more, so a later scrape cannot start a new refresher.
	s.start.Do(func() {})
	if s.stop == nil {
		return
	}
	s.stop()
	<-s.done
}

func (s *vpnState) loop(ctx context.Context) {
	defer close(s.done)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		s.refresh(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-s.kick:
		}
	}
}

// refresh asks the CLI of every running provider, one at a time, outside the
// lock, so a scrape meanwhile serves the previous answers.
func (s *vpnState) refresh(ctx context.Context) {
	for _, p := range vpnProviders {
		if ctx.Err() != nil {
			return
		}
		if !p.installed() || !p.running() {
			s.forget(p.name)
			continue
		}

		checkCtx, cancel := context.WithTimeout(ctx, vpnCheckTimeout)
		connected := p.connected(checkCtx)
		cancel()
		if ctx.Err() != nil {
			return
		}

		s.mu.Lock()
		// The daemon may have stopped while its CLI was asked; a scrape
		// that saw it stopped has forgotten the answer, so do not bring
		// the old one back.
		if p.running() {
			s.answers[p.name] = connected
		} else {
			delete(s.answers, p.name)
		}
		s.mu.Unlock()
	}
}

func (s *vpnState) forget(name string) {
	s.mu.Lock()
	delete(s.answers, name)
	s.mu.Unlock()
}
