package hid

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	linkConfigured = usbLink{State: udcStateConfigured, Speed: udcSpeedHigh}
	linkGone       = usbLink{State: udcStateDetached, Speed: "UNKNOWN"}
	linkDefault    = usbLink{State: "default", Speed: "UNKNOWN"}
	linkAddressed  = usbLink{State: "addressed", Speed: udcSpeedHigh}
	linkSuspended  = usbLink{State: "suspended", Speed: udcSpeedHigh}
)

// reenumerationsDuring returns how many re-enumerations run counted.
func reenumerationsDuring(run func()) uint64 {
	before := USBReenumerations()
	run()
	return USBReenumerations() - before
}

// A port reset read change by change: one re-enumeration, not one per state.
func TestTrackCountsAResetOnce(t *testing.T) {
	freezeClock(t, epoch)
	w := &usbWatchdog{}

	got := reenumerationsDuring(func() {
		w.track(linkConfigured, epoch, false)
		w.track(linkGone, epoch.Add(time.Second), true)
		w.track(linkDefault, epoch.Add(time.Second+100*time.Millisecond), true)
		w.track(linkAddressed, epoch.Add(time.Second+200*time.Millisecond), true)
		w.track(linkConfigured, epoch.Add(time.Second+300*time.Millisecond), true)
		// The next sample reads configured again. That is not a second reset.
		w.track(linkConfigured, epoch.Add(2*time.Second), false)
	})
	if got != 1 {
		t.Fatalf("a reset counted %d times, want 1", got)
	}
}

// The notification comes from a work item, so a whole reset can wake the
// reader once, and the read already says configured.
func TestTrackCountsAResetSeenAsOneWakeup(t *testing.T) {
	freezeClock(t, epoch)
	w := &usbWatchdog{}

	got := reenumerationsDuring(func() {
		w.track(linkConfigured, epoch, false)
		w.track(linkConfigured, epoch.Add(time.Second), true)
	})
	if got != 1 {
		t.Fatalf("a collapsed reset counted %d times, want 1", got)
	}
}

// Samples alone never count configured after configured: that is the link
// staying up.
func TestTrackDoesNotCountASteadyLink(t *testing.T) {
	freezeClock(t, epoch)
	w := &usbWatchdog{}

	got := reenumerationsDuring(func() {
		for i := 0; i < 5; i++ {
			w.track(linkConfigured, epoch.Add(time.Duration(i)*usbPollInterval), false)
		}
	})
	if got != 0 {
		t.Fatalf("a steady link counted %d re-enumerations", got)
	}
}

// The first enumeration after the server starts is not an "again".
func TestTrackDoesNotCountTheFirstEnumeration(t *testing.T) {
	freezeClock(t, epoch)
	w := &usbWatchdog{}

	got := reenumerationsDuring(func() {
		w.track(linkGone, epoch, false)
		w.track(linkDefault, epoch.Add(time.Second), true)
		w.track(linkConfigured, epoch.Add(2*time.Second), true)
	})
	if got != 0 {
		t.Fatalf("the first enumeration counted %d times", got)
	}
}

// A host that sleeps and wakes suspends and resumes the bus. Nothing was
// enumerated again.
func TestTrackDoesNotCountSuspendAndResume(t *testing.T) {
	freezeClock(t, epoch)
	w := &usbWatchdog{}

	got := reenumerationsDuring(func() {
		w.track(linkConfigured, epoch, false)
		w.track(linkSuspended, epoch.Add(time.Second), true)
		w.track(linkConfigured, epoch.Add(time.Minute), true)
	})
	if got != 0 {
		t.Fatalf("a suspend and resume counted %d times", got)
	}
}

// The server's own rebind leaves configured and comes back inside the settle
// window. That is a recovery, already counted as one.
func TestTrackDoesNotCountTheServersOwnRebind(t *testing.T) {
	freezeClock(t, epoch)
	w := &usbWatchdog{}

	got := reenumerationsDuring(func() {
		w.track(linkConfigured, epoch, false)
		NoteUSBGadgetMutated()
		w.track(linkGone, epoch, true)
		w.track(linkConfigured, epoch, true)
	})
	if got != 0 {
		t.Fatalf("a rebind of our own counted %d times", got)
	}
}

// poll drains the events before it samples, so a reset between two samples is
// counted even though both samples read configured.
func TestPollCountsAResetBetweenSamples(t *testing.T) {
	freezeClock(t, epoch)
	fakeUDC(t, map[string]string{"state": "configured", "current_speed": "high-speed"})
	for len(usbLinkEvents) > 0 {
		<-usbLinkEvents
	}

	w := &usbWatchdog{}
	w.poll()

	got := reenumerationsDuring(func() {
		sendUSBLinkEvent(usbLinkEvent{at: epoch, link: linkGone})
		sendUSBLinkEvent(usbLinkEvent{at: epoch, link: linkDefault})
		sendUSBLinkEvent(usbLinkEvent{at: epoch, link: linkConfigured})
		w.poll()
	})
	if got != 1 {
		t.Fatalf("poll counted %d re-enumerations, want 1", got)
	}
	if !w.faultSince.IsZero() {
		t.Fatal("a reset the link came back from started the fault timer")
	}
}

// An enumeration that stops half way is a fault now (#70): after the debounce
// the watchdog rebinds.
func TestStuckEnumerationStartsTheFaultTimer(t *testing.T) {
	freezeClock(t, epoch)
	w := &usbWatchdog{sawHealthy: true}

	w.observe(linkAddressed)
	if w.faultSince.IsZero() {
		t.Fatal("an enumeration in progress did not start the fault timer")
	}
	if got := w.decide(epoch.Add(usbFaultDebounceSeen-time.Second), time.Time{}); got != usbActionNone {
		t.Fatalf("acted on an enumeration %s old: %v", usbFaultDebounceSeen-time.Second, got)
	}
	if got := w.decide(epoch.Add(usbFaultDebounceSeen), time.Time{}); got != usbActionRebind {
		t.Fatalf("a stuck enumeration asked for %v, want a rebind", got)
	}
}

// fakeScript stands in for /etc/init.d/S00aagadget.
func fakeScript(t *testing.T, mode os.FileMode) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "S00aagadget")
	if mode != 0 {
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, mode); err != nil {
			t.Fatal(err)
		}
	}
	previous := gadgetScriptStat
	gadgetScriptStat = func(name string) (fs.FileInfo, error) {
		if name == MainlineGadgetScript {
			return os.Stat(path)
		}
		return previous(name)
	}
	t.Cleanup(func() { gadgetScriptStat = previous })
}

// The watchdog's rebind must reach a script that runs on both slots (#70).
func TestGadgetScriptPerSlot(t *testing.T) {
	cases := []struct {
		name string
		mode os.FileMode
		want string
	}{
		{"vendor slot, no S00aagadget", 0, USBDevScript},
		{"mainline slot", 0o755, MainlineGadgetScript},
		{"S00aagadget present but not executable", 0o644, USBDevScript},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fakeScript(t, tc.mode)
			if got := gadgetScript(); got != tc.want {
				t.Fatalf("gadget script %s, want %s", got, tc.want)
			}
		})
	}
}
