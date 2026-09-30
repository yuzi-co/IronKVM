//go:build linux

package vpn

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
)

// fakeSteps records the steps it is asked for. Start makes the daemon run;
// State answers from states in turn and then repeats the last one.
type fakeSteps struct {
	t      *testing.T
	calls  []string
	states []proto.VpnState
	upErr  error
	downEr error
}

func (f *fakeSteps) steps() Steps {
	return Steps{
		Start: func() error {
			f.calls = append(f.calls, "start")
			fakeRunning(f.t, addon.NetBird, 4242)
			return nil
		},
		Stop: func() error { f.calls = append(f.calls, "stop"); return nil },
		Up:   func() error { f.calls = append(f.calls, "up"); return f.upErr },
		Down: func() error { f.calls = append(f.calls, "down"); return f.downEr },
		State: func() (proto.VpnState, error) {
			st := f.states[0]
			if len(f.states) > 1 {
				f.states = f.states[1:]
			}
			return st, nil
		},
	}
}

// actions leaves out the state queries.
func (f *fakeSteps) actions() string {
	return strings.Join(f.calls, ",")
}

func quickSettle(t *testing.T) {
	t.Helper()
	savedWait, savedPoll := settleWait, settlePoll
	savedBootWait, savedBootPoll := bootWait, bootPoll
	t.Cleanup(func() {
		settleWait, settlePoll = savedWait, savedPoll
		bootWait, bootPoll = savedBootWait, savedBootPoll
	})
	settleWait, settlePoll = 50*time.Millisecond, time.Millisecond
	bootWait, bootPoll = 50*time.Millisecond, time.Millisecond
}

func TestConnectStartsThenUps(t *testing.T) {
	scratchNetBird(t, "")
	quickSettle(t)
	f := &fakeSteps{t: t, states: []proto.VpnState{proto.VpnStopped}}
	if err := Connect(addon.NetBird, f.steps()); err != nil {
		t.Fatal(err)
	}
	if got := f.actions(); got != "start,up" {
		t.Fatalf("got %s", got)
	}
}

func TestConnectWaitsForAStartingDaemon(t *testing.T) {
	scratchNetBird(t, "")
	quickSettle(t)
	f := &fakeSteps{t: t, states: []proto.VpnState{proto.VpnNotRunning, proto.VpnNotRunning, proto.VpnStopped}}
	if err := Connect(addon.NetBird, f.steps()); err != nil {
		t.Fatal(err)
	}
	if got := f.actions(); got != "start,up" {
		t.Fatalf("got %s", got)
	}
}

func TestConnectLeavesAConnectedDaemonAlone(t *testing.T) {
	scratchNetBird(t, "")
	quickSettle(t)
	fakeRunning(t, addon.NetBird, 4242)
	f := &fakeSteps{t: t, states: []proto.VpnState{proto.VpnRunning}}
	if err := Connect(addon.NetBird, f.steps()); err != nil {
		t.Fatal(err)
	}
	if got := f.actions(); got != "" {
		t.Fatalf("got %s", got)
	}
}

func TestConnectWithoutLoginOnlyStarts(t *testing.T) {
	scratchNetBird(t, "")
	quickSettle(t)
	f := &fakeSteps{t: t, states: []proto.VpnState{proto.VpnNotLogin}}
	if err := Connect(addon.NetBird, f.steps()); err != nil {
		t.Fatal(err)
	}
	if got := f.actions(); got != "start" {
		t.Fatalf("got %s", got)
	}
}

func TestConnectUpFailure(t *testing.T) {
	scratchNetBird(t, "")
	quickSettle(t)
	fakeRunning(t, addon.NetBird, 4242)

	// An up refused because the daemon reconnected on its own is success.
	f := &fakeSteps{t: t, upErr: errors.New("up already in progress"),
		states: []proto.VpnState{proto.VpnStopped, proto.VpnRunning}}
	if err := Connect(addon.NetBird, f.steps()); err != nil {
		t.Fatalf("got %v", err)
	}

	f = &fakeSteps{t: t, upErr: errors.New("management unreachable"),
		states: []proto.VpnState{proto.VpnStopped}}
	if err := Connect(addon.NetBird, f.steps()); err == nil || err.Error() != "management unreachable" {
		t.Fatalf("got %v", err)
	}
}

func TestDisconnectDownsThenStops(t *testing.T) {
	scratchNetBird(t, "")
	fakeRunning(t, addon.NetBird, 4242)
	f := &fakeSteps{t: t}
	if err := Disconnect(addon.NetBird, f.steps()); err != nil {
		t.Fatal(err)
	}
	if got := f.actions(); got != "down,stop" {
		t.Fatalf("got %s", got)
	}

	// A failed down still stops the daemon.
	f = &fakeSteps{t: t, downEr: errors.New("no daemon socket")}
	if err := Disconnect(addon.NetBird, f.steps()); err != nil {
		t.Fatal(err)
	}
	if got := f.actions(); got != "down,stop" {
		t.Fatalf("got %s", got)
	}
}

func TestDisconnectOfAStoppedDaemonOnlyStops(t *testing.T) {
	scratchNetBird(t, "")
	f := &fakeSteps{t: t}
	if err := Disconnect(addon.NetBird, f.steps()); err != nil {
		t.Fatal(err)
	}
	if got := f.actions(); got != "stop" {
		t.Fatalf("got %s", got)
	}
}

// bootScratch is a board up for uptime seconds with NetBird running, and
// connecting at boot when enabled.
func bootScratch(t *testing.T, uptime string, enabled bool) {
	t.Helper()
	scratchNetBird(t, "")
	quickSettle(t)
	fakeRunning(t, addon.NetBird, 4242)
	if err := os.WriteFile(filepath.Join(addon.ProcDir, "uptime"), []byte(uptime+" 100.00\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if enabled {
		if err := addon.SetBoot(addon.NetBird, true); err != nil {
			t.Fatal(err)
		}
	}
}

func TestConnectAtBootBringsAStoppedDaemonUp(t *testing.T) {
	bootScratch(t, "60.5", true)
	f := &fakeSteps{t: t, states: []proto.VpnState{proto.VpnNotRunning, proto.VpnStopped, proto.VpnStopped}}
	ConnectAtBoot(addon.NetBird, f.steps())
	if got := f.actions(); got != "up" {
		t.Fatalf("got %s", got)
	}
}

func TestConnectAtBootWaitsOutAMomentaryIdle(t *testing.T) {
	bootScratch(t, "60.5", true)
	f := &fakeSteps{t: t, states: []proto.VpnState{proto.VpnStopped, proto.VpnRunning}}
	ConnectAtBoot(addon.NetBird, f.steps())
	if got := f.actions(); got != "" {
		t.Fatalf("got %s", got)
	}
}

func TestConnectAtBootLeavesTheDaemonAlone(t *testing.T) {
	for name, tc := range map[string]struct {
		uptime  string
		enabled bool
		state   proto.VpnState
	}{
		"boot off":        {"60.5", false, proto.VpnStopped},
		"long after boot": {"3600.0", true, proto.VpnStopped},
		"not logged in":   {"60.5", true, proto.VpnNotLogin},
	} {
		t.Run(name, func(t *testing.T) {
			bootScratch(t, tc.uptime, tc.enabled)
			f := &fakeSteps{t: t, states: []proto.VpnState{tc.state}}
			ConnectAtBoot(addon.NetBird, f.steps())
			if got := f.actions(); got != "" {
				t.Fatalf("got %s", got)
			}
		})
	}
}
