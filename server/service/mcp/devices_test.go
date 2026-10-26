package mcpservice

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/service/audit"
)

// fakeHost is the board's buttons, power LED and drives, as the tools see them.
type fakeHost struct {
	pressed      []string
	ledConnected bool
	ledOn        bool
	ledErr       error
	ledReads     int
	inserted     []string
	ejected      []string
}

func (h *fakeHost) devices() Devices {
	return Devices{
		PressButton: func(kind string, d time.Duration) error {
			h.pressed = append(h.pressed, kind+" "+d.String())
			return nil
		},
		PowerLEDConnected: func() bool { return h.ledConnected },
		PowerLED: func() (bool, error) {
			h.ledReads++
			return h.ledOn, h.ledErr
		},
		ListMedia: func() ([]MediaDrive, []string, error) {
			return []MediaDrive{
				{ID: "disk"},
				{ID: "cdrom", File: "/data/win11.iso", ReadOnly: true},
			}, []string{"/data/win11.iso", "/data/rescue.img"}, nil
		},
		InsertMedia: func(drive, file string, ro bool) error {
			h.inserted = append(h.inserted, drive+" "+file+" ro="+map[bool]string{true: "1", false: "0"}[ro])
			return nil
		},
		EjectMedia: func(drive string) error {
			h.ejected = append(h.ejected, drive)
			return nil
		},
	}
}

func allowTiers(t *testing.T, tiers ...audit.Tier) {
	t.Helper()
	if _, err := audit.SaveTiers(audit.Tiers{MCP: tiers, Agent: []audit.Tier{audit.TierObserve}}); err != nil {
		t.Fatal(err)
	}
}

func isError(out string) bool {
	return strings.Contains(out, `"isError":true`) || strings.Contains(out, `"error":{`)
}

func TestReadToolsAreInTheObserveTier(t *testing.T) {
	isolateAudit(t)
	host := &fakeHost{}
	list := newSession(t, host.devices()).tools()

	// The default tiers, observe and input, offer the status and the list
	// but not what presses or changes a drive.
	for _, name := range []string{"cube_power_status", "cube_media_list"} {
		if !strings.Contains(list, `"name":"`+name+`"`) {
			t.Errorf("%s is not offered with the observe tier on", name)
		}
	}
	for _, name := range []string{"cube_power", "cube_media"} {
		if strings.Contains(list, `"name":"`+name+`"`) {
			t.Errorf("%s is offered with only observe and input on", name)
		}
	}
	if toolTier(powerStatusToolName) != audit.TierObserve || toolTier(mediaListToolName) != audit.TierObserve {
		t.Fatal("the read tools are not in the observe tier")
	}
}

func TestReadToolsAreRefusedWithObserveOff(t *testing.T) {
	isolateAudit(t)
	allowTiers(t, audit.TierInput)
	host := &fakeHost{ledConnected: true}
	s := newSession(t, host.devices())

	if out := s.call("cube_power_status", `{}`); !strings.Contains(out, "tier is off") {
		t.Fatalf("status answered with observe off: %s", out)
	}
	if host.ledReads != 0 {
		t.Fatal("the LED was read with observe off")
	}
}

func TestPowerStatusIsUnknownWithoutTheLED(t *testing.T) {
	isolateAudit(t)
	host := &fakeHost{ledConnected: false, ledOn: true}
	out := newSession(t, host.devices()).call("cube_power_status", `{}`)

	if isError(out) || !strings.Contains(out, `"state":"unknown"`) || !strings.Contains(out, `"ledConnected":false`) {
		t.Fatalf("status without the LED wired: %s", out)
	}
	if host.ledReads != 0 {
		t.Fatal("an LED the owner did not wire was read")
	}
}

func TestPowerStatusFollowsTheLED(t *testing.T) {
	for _, tc := range []struct {
		on    bool
		err   error
		state string
	}{
		{on: true, state: "on"},
		{on: false, state: "off"},
		{err: errors.New("gpio unreadable"), state: "unknown"},
	} {
		isolateAudit(t)
		host := &fakeHost{ledConnected: true, ledOn: tc.on, ledErr: tc.err}
		out := newSession(t, host.devices()).call("cube_power_status", `{}`)
		if isError(out) || !strings.Contains(out, `"state":"`+tc.state+`"`) || !strings.Contains(out, `"ledConnected":true`) {
			t.Fatalf("LED on=%v err=%v: %s", tc.on, tc.err, out)
		}
	}
}

func TestPowerPressShortAndLong(t *testing.T) {
	isolateAudit(t)
	allowTiers(t, audit.TierObserve, audit.TierPower)
	host := &fakeHost{}
	s := newSession(t, host.devices())

	for _, args := range []string{
		`{"button":"power"}`,
		`{"button":"power","press":"short"}`,
		`{"button":"power","press":"long"}`,
		`{"button":"reset","press":"short"}`,
	} {
		if out := s.call("cube_power", args); isError(out) {
			t.Fatalf("%s failed: %s", args, out)
		}
	}
	want := []string{"power 800ms", "power 800ms", "power 5s", "reset 800ms"}
	if strings.Join(host.pressed, ",") != strings.Join(want, ",") {
		t.Fatalf("pressed = %v, want %v", host.pressed, want)
	}
}

func TestPowerPressRefusesBadArguments(t *testing.T) {
	isolateAudit(t)
	allowTiers(t, audit.TierObserve, audit.TierPower)
	host := &fakeHost{}
	s := newSession(t, host.devices())

	for _, args := range []string{
		`{"button":"power","press":"medium"}`,
		`{"button":"power","press":"long","holdMs":300}`,
		`{"button":"hdd"}`,
	} {
		if out := s.call("cube_power", args); !isError(out) {
			t.Errorf("%s was accepted: %s", args, out)
		}
	}
	if len(host.pressed) != 0 {
		t.Fatalf("a refused call pressed: %v", host.pressed)
	}
}

func TestPowerPressIsRefusedWhileAnotherIsHeld(t *testing.T) {
	held := make(chan struct{})
	release := make(chan struct{})
	var presses int
	handler := powerHandler(Devices{PressButton: func(string, time.Duration) error {
		presses++
		close(held)
		<-release
		return nil
	}})

	done := make(chan error)
	go func() {
		_, _, err := handler(context.Background(), nil, PowerParams{Button: "power"})
		done <- err
	}()
	<-held
	if _, _, err := handler(context.Background(), nil, PowerParams{Button: "reset"}); !errors.Is(err, errPressBusy) {
		t.Fatalf("second press while one is held: err = %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if presses != 1 {
		t.Fatalf("presses = %d, want 1", presses)
	}
}

func TestPowerDescriptionWarnsOfTheLongPress(t *testing.T) {
	isolateAudit(t)
	allowTiers(t, audit.TierObserve, audit.TierPower)
	list := newSession(t, (&fakeHost{}).devices()).tools()
	if !strings.Contains(list, "forces the host off") {
		t.Fatalf("cube_power does not say a long press forces the host off: %s", list)
	}
}

func TestMediaListShowsDrivesAndImages(t *testing.T) {
	isolateAudit(t)
	out := newSession(t, (&fakeHost{}).devices()).call("cube_media_list", `{}`)

	for _, want := range []string{`"id":"cdrom"`, `"file":"/data/win11.iso"`, `"readOnly":true`, `"/data/rescue.img"`, `"id":"disk"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("list lacks %s: %s", want, out)
		}
	}
}

func TestMediaInsertAndEject(t *testing.T) {
	isolateAudit(t)
	allowTiers(t, audit.TierObserve, audit.TierMedia)
	host := &fakeHost{}
	s := newSession(t, host.devices())

	for _, args := range []string{
		`{"action":"insert","drive":"cdrom","file":"/data/win11.iso"}`,
		`{"action":"insert","drive":"disk","file":"/data/rescue.img"}`,
		`{"action":"insert","drive":"disk","file":"/data/rescue.img","readOnly":false}`,
		`{"action":"eject","drive":"disk"}`,
	} {
		if out := s.call("cube_media", args); isError(out) {
			t.Fatalf("%s failed: %s", args, out)
		}
	}
	want := []string{"cdrom /data/win11.iso ro=1", "disk /data/rescue.img ro=1", "disk /data/rescue.img ro=0"}
	if strings.Join(host.inserted, ",") != strings.Join(want, ",") {
		t.Fatalf("inserted = %v, want %v", host.inserted, want)
	}
	if strings.Join(host.ejected, ",") != "disk" {
		t.Fatalf("ejected = %v", host.ejected)
	}
}

func TestMediaRefusesAWritableCD(t *testing.T) {
	isolateAudit(t)
	allowTiers(t, audit.TierObserve, audit.TierMedia)
	host := &fakeHost{}
	s := newSession(t, host.devices())

	if out := s.call("cube_media", `{"action":"insert","drive":"cdrom","file":"/data/win11.iso","readOnly":false}`); !isError(out) {
		t.Fatalf("a writable CD was accepted: %s", out)
	}
	if out := s.call("cube_media", `{"action":"list"}`); !isError(out) {
		t.Fatalf("cube_media still lists; listing is cube_media_list, in the observe tier: %s", out)
	}
	if len(host.inserted) != 0 {
		t.Fatalf("inserted = %v", host.inserted)
	}
}

func TestDeviceCallsAreAudited(t *testing.T) {
	store := isolateAudit(t)
	allowTiers(t, audit.TierObserve, audit.TierPower, audit.TierMedia)
	host := &fakeHost{ledConnected: true, ledOn: true}
	s := newSession(t, host.devices())

	s.call("cube_power_status", `{}`)
	s.call("cube_media_list", `{}`)
	s.call("cube_power", `{"button":"power","press":"long"}`)
	s.call("cube_media", `{"action":"insert","drive":"disk","file":"/data/rescue.img","readOnly":false}`)

	got := events(t, store)
	if len(got) != 4 {
		t.Fatalf("got %d events, want 4: %+v", len(got), got)
	}
	// Newest first.
	want := []struct{ action, tier, target, detail string }{
		{"cube_media", "media", "/data/rescue.img", "insert disk rw"},
		{"cube_power", "power", "power", "long hold=5000ms"},
		{"cube_media_list", "observe", "", ""},
		{"cube_power_status", "observe", "", ""},
	}
	for i, w := range want {
		e := got[i]
		if e.Action != w.action || e.Tier != w.tier || e.Target != w.target || e.Detail != w.detail || e.Outcome != audit.OutcomeOK {
			t.Errorf("event %d = %+v, want %+v", i, e, w)
		}
	}
}
