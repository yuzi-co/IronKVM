package ws

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hid"
	"NanoKVM-Server/service/inputcontrol"

	"github.com/gorilla/websocket"
)

// fakeTimers stands in for time.AfterFunc. Fire runs the newest timer's
// function the way the runtime would, whether or not it was stopped since, so
// the tests can reproduce a timer that fired while it was being re-armed.
type fakeTimers struct {
	mutex  sync.Mutex
	timers []*fakeTimer
}

type fakeTimer struct {
	f       func()
	stopped bool
}

func (t *fakeTimer) Stop() bool {
	was := !t.stopped
	t.stopped = true
	return was
}

func (f *fakeTimers) new(_ time.Duration, fn func()) holdTimer {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	timer := &fakeTimer{f: fn}
	f.timers = append(f.timers, timer)
	return timer
}

func (f *fakeTimers) count() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return len(f.timers)
}

func (f *fakeTimers) fire(i int) {
	f.mutex.Lock()
	timer := f.timers[i]
	// A timer that has run is no longer pending.
	timer.stopped = true
	f.mutex.Unlock()
	timer.f()
}

func (f *fakeTimers) live() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	n := 0
	for _, timer := range f.timers {
		if !timer.stopped {
			n++
		}
	}
	return n
}

func newTestWatch() (*holdWatch, *fakeTimers, *[]holdKind) {
	timers := &fakeTimers{}
	var released []holdKind
	w := newHoldWatch(time.Second, func(kind holdKind) { released = append(released, kind) })
	w.newTimer = timers.new
	return w, timers, &released
}

func TestHoldWatchIsOffUntilTheFirstKeepalive(t *testing.T) {
	w, timers, released := newTestWatch()

	w.Input(holdKeyboard, true)
	if timers.count() != 0 {
		t.Fatalf("a client that never sent a keepalive armed %d timers", timers.count())
	}

	w.Keepalive()
	if timers.live() != 1 {
		t.Fatalf("the first keepalive with a key held armed %d timers, want 1", timers.live())
	}
	timers.fire(0)
	if len(*released) != 1 || (*released)[0] != holdKeyboard {
		t.Fatalf("released %v, want the keyboard", *released)
	}
}

func TestHoldWatchArmsNothingWhileIdle(t *testing.T) {
	w, timers, released := newTestWatch()

	w.Keepalive()
	w.Keepalive()
	w.Input(holdKeyboard, false)
	w.Input(holdRelativeMouse, false)
	if timers.count() != 0 || len(*released) != 0 {
		t.Fatalf("idle watch armed %d timers and released %v", timers.count(), *released)
	}
}

func TestHoldWatchKeepaliveRearmsAndReleaseDisarms(t *testing.T) {
	w, timers, released := newTestWatch()
	w.Keepalive()

	w.Input(holdKeyboard, true)
	w.Keepalive()
	w.Keepalive()
	if timers.live() != 1 {
		t.Fatalf("%d timers live after re-arming, want 1", timers.live())
	}

	// The earlier timers ran after they were replaced. A timer that was
	// already on its way when the keepalive re-armed it must not release.
	timers.fire(0)
	timers.fire(1)
	if len(*released) != 0 {
		t.Fatalf("a superseded timer released %v", *released)
	}

	w.Input(holdKeyboard, false)
	if timers.live() != 0 {
		t.Fatalf("%d timers live after the key-up", timers.live())
	}
	timers.fire(2)
	if len(*released) != 0 {
		t.Fatalf("a timer disarmed by the key-up released %v", *released)
	}
}

func TestHoldWatchReleasesEachKindOnItsOwnDeadline(t *testing.T) {
	w, timers, released := newTestWatch()
	w.Keepalive()

	w.Input(holdKeyboard, true)       // timer 0
	w.Input(holdAbsoluteMouse, true)  // timer 1
	w.Input(holdRelativeMouse, false) // nothing
	if timers.live() != 2 {
		t.Fatalf("%d timers live, want one per held kind", timers.live())
	}

	timers.fire(1)
	if len(*released) != 1 || (*released)[0] != holdAbsoluteMouse {
		t.Fatalf("released %v, want the absolute mouse alone", *released)
	}
	timers.fire(0)
	if len(*released) != 2 || (*released)[1] != holdKeyboard {
		t.Fatalf("released %v, want the keyboard next", *released)
	}

	// Released once. The same timer running twice does not release again,
	// and a keepalive after the release re-arms nothing: the server no
	// longer holds anything, whatever the browser believes.
	timers.fire(0)
	w.Keepalive()
	if len(*released) != 2 || timers.live() != 0 {
		t.Fatalf("released %v with %d timers live after the release", *released, timers.live())
	}
}

func TestHoldWatchCloseStopsEverything(t *testing.T) {
	w, timers, released := newTestWatch()
	w.Keepalive()
	w.Input(holdKeyboard, true)
	w.Input(holdRelativeMouse, true)

	w.Close()
	if timers.live() != 0 {
		t.Fatalf("%d timers live after close", timers.live())
	}
	timers.fire(0)
	timers.fire(1)
	w.Input(holdKeyboard, true)
	w.Keepalive()
	if len(*released) != 0 || timers.count() != 2 {
		t.Fatalf("closed watch released %v and armed %d timers", *released, timers.count())
	}
}

func TestAbsoluteReleaseKeepsThePosition(t *testing.T) {
	got := absoluteReleaseAt([]byte{0x01, 0x34, 0x12, 0x78, 0x56, 0x02})
	want := []byte{0x00, 0x34, 0x12, 0x78, 0x56, 0x00}
	if !bytes.Equal(got, want) {
		t.Fatalf("release = % x, want % x", got, want)
	}
	if got := absoluteReleaseAt(nil); !bytes.Equal(got, make([]byte, hid.AbsoluteMouseReportLen)) {
		t.Fatalf("release without a position = % x", got)
	}
}

// TestStalledClientKeyIsReleased drives a real socket: a key goes down, the
// client keeps it alive for a while, then goes silent without closing. The
// key-up has to reach the keyboard queue within the timeout, long before the
// heartbeat timeout would notice anything.
func TestStalledClientKeyIsReleased(t *testing.T) {
	control := controlmode.NewManager(filepath.Join(t.TempDir(), "mode"), controlmode.ModeMCP)
	manual := inputcontrol.NewManualSession(control, &inputcontrol.Coordinator{})
	defer manual.Close()

	keyboard := make(chan hid.QueuedReport, 16)
	serverDone := make(chan struct{})
	upgrade := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade websocket: %v", err)
			return
		}
		client := &Client{
			ws:               conn,
			hid:              hid.GetHid(),
			manual:           manual,
			keyboard:         keyboard,
			mouse:            make(chan hid.QueuedReport, 16),
			heartbeatTimeout: time.Minute,
		}
		client.holds = newHoldWatch(100*time.Millisecond, client.releaseHeld)
		_ = client.Read()
		client.inputMutex.Lock()
		client.inputClosed = true
		client.inputMutex.Unlock()
		client.holds.Close()
		close(serverDone)
	}))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = conn.Close()
		<-serverDone
	}()

	send := func(data ...byte) {
		if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
			t.Fatal(err)
		}
	}
	take := func() hid.QueuedReport {
		select {
		case report := <-keyboard:
			report.Complete(true)
			return report
		case <-time.After(2 * time.Second):
			t.Fatal("no keyboard report")
			return hid.QueuedReport{}
		}
	}

	send(KeepaliveEvent)
	send(KeyboardEvent, 0x02, 0, 0, 0, 0, 0, 0, 0) // left shift down
	if got := take(); got.Data[0] != 0x02 {
		t.Fatalf("first report = % x, want shift down", got.Data)
	}

	// Kept alive for three timeouts: nothing is released meanwhile.
	start := time.Now()
	for time.Since(start) < 300*time.Millisecond {
		send(KeepaliveEvent)
		time.Sleep(20 * time.Millisecond)
	}
	select {
	case report := <-keyboard:
		t.Fatalf("a key kept alive was released: % x", report.Data)
	default:
	}

	silent := time.Now()
	got := take()
	if !bytes.Equal(got.Data, make([]byte, 8)) {
		t.Fatalf("report after silence = % x, want all keys up", got.Data)
	}
	if elapsed := time.Since(silent); elapsed > time.Second {
		t.Fatalf("release took %s", elapsed)
	}

	// Context for the manual session: the release went through a
	// reservation, so a held lease does not linger.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := manual.Reserve(ctx, inputcontrol.ManualKeyboard, false, nil); err != nil {
		t.Fatalf("reserve after release: %v", err)
	}
}
