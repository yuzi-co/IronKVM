package hid

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// The watchdog asks about the descriptor every two seconds. It is parsed again
// only when its bytes change.
func TestTheDescriptorIsParsedOnlyWhenItChanges(t *testing.T) {
	parses := 0
	memo := &descriptorMemo{parse: func(raw []byte) bool {
		parses++
		return len(raw) > 1
	}}

	for range 5 {
		if !memo.touch([]byte{1, 2}) {
			t.Fatal("wrong answer for the first descriptor")
		}
	}
	if parses != 1 {
		t.Fatalf("parsed %d times for one descriptor, want 1", parses)
	}

	if memo.touch([]byte{1}) {
		t.Fatal("wrong answer after the descriptor changed")
	}
	if parses != 2 {
		t.Fatalf("parsed %d times after one change, want 2", parses)
	}
}

// The refresh reads configfs outside mouseMutex, so it runs alongside the
// pointer's own opens and closes. Run with -race.
func TestRefreshRacesWithThePointer(t *testing.T) {
	touch, extkeys, _ := touchScriptDescriptors(t)
	gadgetReportLength(t, "7\n")
	gadgetReportDesc(t, touch)

	node := filepath.Join(t.TempDir(), "hidg2")
	if err := os.WriteFile(node, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	h := &Hid{}
	device := h.absoluteMouseDevice(node)
	t.Cleanup(func() {
		h.mouseMutex.Lock()
		h.closeDeviceNoLock(device)
		h.mouseMutex.Unlock()
	})

	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		for range 200 {
			h.RefreshAbsoluteReportID()
		}
	}()
	go func() {
		defer workers.Done()
		for i := range 200 {
			h.mouseMutex.Lock()
			if err := h.openDeviceNoLock(device); err != nil {
				t.Error(err)
			}
			h.mouseMutex.Unlock()
			if i%20 == 0 {
				descriptor := touch
				if i%40 == 0 {
					descriptor = extkeys
				}
				if err := os.WriteFile(absoluteReportDescPath, descriptor, 0o644); err != nil {
					t.Error(err)
				}
			}
		}
	}()
	workers.Wait()

	// Settled, the handle's view matches the gadget's.
	if err := os.WriteFile(absoluteReportDescPath, touch, 0o644); err != nil {
		t.Fatal(err)
	}
	h.RefreshAbsoluteReportID()
	h.mouseMutex.Lock()
	if err := h.openDeviceNoLock(device); err != nil {
		t.Fatal(err)
	}
	gotTouch := h.absTouch
	h.mouseMutex.Unlock()
	if !gotTouch {
		t.Fatal("touch off after the gadget settled with touch")
	}
}
