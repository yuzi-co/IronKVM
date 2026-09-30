package metrics

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fakeVpn is a provider whose CLI calls are counted. The refresher runs on
// its own goroutine, hence the atomics.
type fakeVpn struct {
	installed, running, connected atomic.Bool
	calls                         atomic.Int32
	// block, when set, makes the CLI wait for its context and report on
	// entered that it was called.
	block   bool
	entered chan struct{}
}

func newFakeVpn(installed, running, connected bool) *fakeVpn {
	f := &fakeVpn{entered: make(chan struct{}, 1)}
	f.installed.Store(installed)
	f.running.Store(running)
	f.connected.Store(connected)
	return f
}

func (f *fakeVpn) provider(name string) vpnProvider {
	return vpnProvider{
		name:      name,
		installed: f.installed.Load,
		running:   f.running.Load,
		connected: func(ctx context.Context) bool {
			f.calls.Add(1)
			if f.block {
				f.entered <- struct{}{}
				<-ctx.Done()
				return false
			}
			return f.connected.Load()
		},
	}
}

// stubVpn gives the test its own state. With started false the refresher can
// never start, and the test calls refresh itself.
func stubVpn(t *testing.T, started bool, interval time.Duration, providers ...vpnProvider) *vpnState {
	t.Helper()
	setVar(t, &vpnProviders, providers)
	state := newVpnState(interval)
	if !started {
		state.start.Do(func() {})
	}
	setVar(t, &vpn, state)
	// Registered after setVar, so it runs before the originals come back.
	t.Cleanup(state.shutdown)
	return state
}

func TestVpnOnlyInstalledProvidersAreWritten(t *testing.T) {
	ts := newFakeVpn(true, true, true)
	nb := newFakeVpn(false, false, false)
	state := stubVpn(t, false, time.Hour, ts.provider("tailscale"), nb.provider("netbird"))
	state.refresh(context.Background())

	want := `# HELP ironkvm_vpn_connected Whether an installed VPN is connected. Its CLI is asked once a minute in the background.
# TYPE ironkvm_vpn_connected gauge
ironkvm_vpn_connected{provider="tailscale"} 1
`
	assertText(t, render(t, collectVpn), want)
}

func TestVpnScrapeNeverRunsTheCli(t *testing.T) {
	ts := newFakeVpn(true, true, true)
	nb := newFakeVpn(true, false, true)
	stubVpn(t, false, time.Hour, ts.provider("tailscale"), nb.provider("netbird"))

	// Running but not asked yet: left out rather than reported as 0. Not
	// running: 0, which needs no CLI.
	want := `# HELP ironkvm_vpn_connected Whether an installed VPN is connected. Its CLI is asked once a minute in the background.
# TYPE ironkvm_vpn_connected gauge
ironkvm_vpn_connected{provider="netbird"} 0
`
	for range 3 {
		assertText(t, render(t, collectVpn), want)
	}
	if n := ts.calls.Load() + nb.calls.Load(); n != 0 {
		t.Fatalf("a scrape ran a CLI %d times", n)
	}
}

func TestVpnRefreshSkipsAStoppedDaemon(t *testing.T) {
	ts := newFakeVpn(true, false, true)
	nb := newFakeVpn(true, true, false)
	state := stubVpn(t, false, time.Hour, ts.provider("tailscale"), nb.provider("netbird"))
	state.refresh(context.Background())

	want := `# HELP ironkvm_vpn_connected Whether an installed VPN is connected. Its CLI is asked once a minute in the background.
# TYPE ironkvm_vpn_connected gauge
ironkvm_vpn_connected{provider="tailscale"} 0
ironkvm_vpn_connected{provider="netbird"} 0
`
	assertText(t, render(t, collectVpn), want)
	if n := ts.calls.Load(); n != 0 {
		t.Fatalf("the CLI of a stopped daemon was run %d times", n)
	}
	if n := nb.calls.Load(); n != 1 {
		t.Fatalf("the CLI of a running daemon was run %d times, want 1", n)
	}
}

func TestVpnScrapeServesTheKeptAnswer(t *testing.T) {
	ts := newFakeVpn(true, true, true)
	state := stubVpn(t, false, time.Hour, ts.provider("tailscale"))
	state.refresh(context.Background())

	ts.connected.Store(false)
	if got := render(t, collectVpn); !contains(got, `ironkvm_vpn_connected{provider="tailscale"} 1`) {
		t.Fatalf("the kept answer was not served:\n%s", got)
	}

	state.refresh(context.Background())
	if got := render(t, collectVpn); !contains(got, `ironkvm_vpn_connected{provider="tailscale"} 0`) {
		t.Fatalf("the renewed answer was not served:\n%s", got)
	}
}

func TestVpnDaemonThatRestartsIsNotServedTheOldAnswer(t *testing.T) {
	ts := newFakeVpn(true, true, true)
	state := stubVpn(t, false, time.Hour, ts.provider("tailscale"))
	state.refresh(context.Background())

	ts.running.Store(false)
	render(t, collectVpn)
	ts.running.Store(true)
	if got := render(t, collectVpn); strings.Contains(got, "ironkvm_vpn_connected") {
		t.Fatalf("a restarted daemon was served the old answer:\n%s", got)
	}
}

// waitFor polls cond for up to two seconds.
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestVpnRefresherStartsOnFirstScrapeAndIsWokenForANewDaemon(t *testing.T) {
	ts := newFakeVpn(true, true, true)
	stubVpn(t, true, time.Hour, ts.provider("tailscale"))

	render(t, collectVpn)
	waitFor(t, "the first answer", func() bool {
		return contains(render(t, collectVpn), `ironkvm_vpn_connected{provider="tailscale"} 1`)
	})

	// The interval is an hour, so only a wake-up can bring a new answer.
	ts.running.Store(false)
	render(t, collectVpn)
	ts.running.Store(true)
	ts.connected.Store(false)
	render(t, collectVpn)
	waitFor(t, "the answer after a wake-up", func() bool {
		return contains(render(t, collectVpn), `ironkvm_vpn_connected{provider="tailscale"} 0`)
	})
}

func TestVpnStopEndsACliInFlight(t *testing.T) {
	ts := newFakeVpn(true, true, true)
	ts.block = true
	state := stubVpn(t, true, time.Hour, ts.provider("tailscale"))

	render(t, collectVpn)
	select {
	case <-ts.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("the refresher never ran the CLI")
	}

	// A scrape meanwhile does not wait for the CLI.
	scraped := make(chan struct{})
	go func() {
		var out strings.Builder
		collectVpn(NewWriter(&out))
		close(scraped)
	}()
	select {
	case <-scraped:
	case <-time.After(time.Second):
		t.Fatal("a scrape waited for the CLI")
	}

	stopped := make(chan struct{})
	go func() {
		StopVpnRefresher()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(2 * time.Second):
		t.Fatal("stopping did not end the CLI call")
	}

	// Stopped for good: a later scrape does not start it again.
	render(t, collectVpn)
	if state.stop == nil {
		t.Fatal("the refresher had not started")
	}
	select {
	case <-ts.entered:
		t.Fatal("a scrape after stop started the refresher again")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestVpnStopWithoutStartReturns(t *testing.T) {
	stubVpn(t, true, time.Hour)
	StopVpnRefresher()
	StopVpnRefresher()
}

func contains(text, line string) bool {
	return strings.Contains(text, line+"\n")
}
