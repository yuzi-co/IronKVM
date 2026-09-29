package vm

import (
	"errors"
	"net/netip"
	"os"
	"slices"
	"strings"
	"testing"

	"NanoKVM-Server/proto"
)

// board is a fake set of markers and a subnet file, with every rebuild and
// subnet write recorded.
type board struct {
	markers  map[string]bool
	subnet   string
	local    []netip.Prefix
	rebuilds [][]string
	written  []string
	fail     error
}

func newBoard(markers ...string) *board {
	b := &board{markers: map[string]bool{}}
	for _, marker := range markers {
		b.markers[marker] = true
	}

	return b
}

func (b *board) env() usbApplyEnv {
	return usbApplyEnv{
		present: func(marker string) bool { return b.markers[marker] },
		read: func(string) ([]byte, error) {
			if b.subnet == "" {
				return nil, os.ErrNotExist
			}
			return []byte(b.subnet + "\n"), nil
		},
		writeSubnet: func(s usbSubnet) error {
			b.written = append(b.written, s.prefix.String())
			return nil
		},
		local: func() []netip.Prefix { return b.local },
		rebuild: func(commands []string) error {
			b.rebuilds = append(b.rebuilds, commands)
			return b.fail
		},
	}
}

func applyReq(console, disk, audio bool, mode, subnet string) proto.ApplyVirtualDeviceReq {
	return proto.ApplyVirtualDeviceReq{
		Console: console,
		Disk:    disk,
		Audio:   audio,
		Network: proto.ApplyUSBNetworkRequest{Mode: mode, Subnet: subnet},
	}
}

func count(commands []string, command string) int {
	n := 0
	for _, c := range commands {
		if c == command {
			n++
		}
	}

	return n
}

// The reason the endpoint exists: several changes, one rebuild.
func TestApplyRebuildsOnceForSeveralChanges(t *testing.T) {
	b := newBoard(virtualConsole, virtualDisk, virtualAudio)

	// Console and disk off, network on as NCM: three changes.
	if failure := applyUSBSelection(applyReq(false, false, true, usbNetworkNCM, ""), b.env()); failure != nil {
		t.Fatalf("refused: %+v", failure)
	}

	if len(b.rebuilds) != 1 {
		t.Fatalf("rebuilt %d times, want 1", len(b.rebuilds))
	}

	commands := b.rebuilds[0]
	if count(commands, usbGadgetStop) != 1 || count(commands, usbGadgetStart) != 1 {
		t.Errorf("want one stop and one start, got %v", commands)
	}
	if commands[0] != usbGadgetStop || commands[len(commands)-1] != usbGadgetStart {
		t.Errorf("want stop first and start last, got %v", commands)
	}

	for _, want := range []string{"rm -f /boot/usb.acm", "rm -f /boot/usb.disk0", "touch " + virtualNetworkNCM} {
		if !slices.Contains(commands, want) {
			t.Errorf("missing %q in %v", want, commands)
		}
	}

	for _, c := range commands {
		if strings.Contains(c, "usb.uac") {
			t.Errorf("audio did not change, yet %q runs", c)
		}
	}
}

func TestApplyWithNoChangeDoesNotRebuild(t *testing.T) {
	b := newBoard(virtualConsole, virtualDisk, virtualAudio)
	b.subnet = "172.31.255.0/30"

	if failure := applyUSBSelection(applyReq(true, true, true, usbNetworkOff, "172.31.255.0/30"), b.env()); failure != nil {
		t.Fatalf("refused: %+v", failure)
	}

	if len(b.rebuilds) != 0 || len(b.written) != 0 {
		t.Errorf("rebuilds %v, writes %v, want neither", b.rebuilds, b.written)
	}
}

// A new subnet rebuilds only while the link is on, as SetUSBNetwork decides.
func TestApplySubnetChangeRebuildsOnlyWithTheLinkOn(t *testing.T) {
	off := newBoard(virtualDisk)
	if failure := applyUSBSelection(applyReq(false, true, false, usbNetworkOff, "192.168.77.0/30"), off.env()); failure != nil {
		t.Fatalf("refused: %+v", failure)
	}
	if len(off.rebuilds) != 0 || !slices.Equal(off.written, []string{"192.168.77.0/30"}) {
		t.Errorf("link off: rebuilds %v, writes %v", off.rebuilds, off.written)
	}

	on := newBoard(virtualDisk, virtualNetworkNCM)
	if failure := applyUSBSelection(applyReq(false, true, false, usbNetworkNCM, "192.168.77.0/30"), on.env()); failure != nil {
		t.Fatalf("refused: %+v", failure)
	}
	if len(on.rebuilds) != 1 || !slices.Equal(on.rebuilds[0], []string{usbGadgetStop, usbGadgetStart}) {
		t.Errorf("link on: rebuilds %v, want one bare stop and start", on.rebuilds)
	}
}

// The budget is checked against the whole set, not one switch at a time: the
// console and the network never fit together beside HID.
func TestApplyRefusesASetOverTheBudget(t *testing.T) {
	b := newBoard(virtualDisk)

	failure := applyUSBSelection(applyReq(true, true, false, usbNetworkNCM, ""), b.env())
	if failure == nil || failure.code != -4 {
		t.Fatalf("got %+v, want a -4 refusal", failure)
	}

	// Eight inbound of six: dropping the console or the network fits, dropping
	// the disk alone does not, so the disk is not offered.
	want := "these devices need 8 inbound USB endpoints, 6 available: turn off console or network"
	if failure.message != want {
		t.Errorf("refusal %q, want %q", failure.message, want)
	}

	if len(b.rebuilds) != 0 || len(b.written) != 0 {
		t.Errorf("a refusal touched the board: rebuilds %v, writes %v", b.rebuilds, b.written)
	}
}

// Swapping the console for the network in one apply fits, although neither
// order of two single switches would let the second one on first.
func TestApplySwapsTheConsoleForTheNetworkInOneStep(t *testing.T) {
	b := newBoard(virtualConsole, virtualDisk, virtualAudio)

	if failure := applyUSBSelection(applyReq(false, true, true, usbNetworkECM, ""), b.env()); failure != nil {
		t.Fatalf("refused: %+v", failure)
	}

	if len(b.rebuilds) != 1 || !slices.Contains(b.rebuilds[0], "touch "+virtualNetworkECM) {
		t.Errorf("rebuilds %v, want one that turns ECM on", b.rebuilds)
	}
}

// A board over its budget can still give things up, as a single switch can.
func TestApplyAllowsSheddingOnABoardOverItsBudget(t *testing.T) {
	b := newBoard(virtualConsole, virtualDisk, virtualNetworkNCM, virtualAudio)

	if failure := applyUSBSelection(applyReq(true, false, true, usbNetworkNCM, ""), b.env()); failure != nil {
		t.Fatalf("refused shedding the disk: %+v", failure)
	}

	// Still over, and adding the disk back is refused.
	b = newBoard(virtualConsole, virtualNetworkNCM)
	if failure := applyUSBSelection(applyReq(true, true, false, usbNetworkNCM, ""), b.env()); failure == nil {
		t.Error("added the disk to a board already over its budget")
	}
}

// With HID switched off its endpoints are free, and the check knows it.
func TestApplyChargesNothingForHidWhenItIsOff(t *testing.T) {
	b := newBoard(disableHid)

	if failure := applyUSBSelection(applyReq(true, true, true, usbNetworkNCM, ""), b.env()); failure != nil {
		t.Fatalf("refused everything with HID off: %+v", failure)
	}
}

func TestApplyValidatesTheSubnetBeforeTouchingAnything(t *testing.T) {
	b := newBoard()
	b.local = []netip.Prefix{netip.MustParsePrefix("192.168.1.0/24")}

	for _, subnet := range []string{"8.8.8.0/30", "172.31.255.1/30", "192.168.1.0/30", "nonsense"} {
		failure := applyUSBSelection(applyReq(false, false, false, usbNetworkNCM, subnet), b.env())
		if failure == nil || failure.code != -2 {
			t.Errorf("%s: got %+v, want -2", subnet, failure)
		}
	}

	if len(b.rebuilds) != 0 || len(b.written) != 0 {
		t.Errorf("an invalid subnet touched the board: rebuilds %v, writes %v", b.rebuilds, b.written)
	}
}

func TestApplyKeepsRndisOnlyWhereItRuns(t *testing.T) {
	fresh := newBoard()
	if failure := applyUSBSelection(applyReq(false, false, false, usbNetworkRNDIS, ""), fresh.env()); failure == nil || failure.code != -2 {
		t.Errorf("turned RNDIS on: %+v", failure)
	}

	old := newBoard(virtualNetworkRNDIS)
	if failure := applyUSBSelection(applyReq(false, true, false, usbNetworkRNDIS, ""), old.env()); failure != nil {
		t.Fatalf("refused keeping RNDIS: %+v", failure)
	}
	for _, c := range old.rebuilds[0] {
		if strings.Contains(c, "usb.rndis0") || strings.Contains(c, "usb.ncm") {
			t.Errorf("kept link was touched: %q", c)
		}
	}
}

func TestApplyNetworkOffClearsEveryMarker(t *testing.T) {
	b := newBoard(virtualNetworkECM)

	if failure := applyUSBSelection(applyReq(false, false, false, usbNetworkOff, ""), b.env()); failure != nil {
		t.Fatalf("refused: %+v", failure)
	}

	for _, marker := range []string{virtualNetworkNCM, virtualNetworkECM, virtualNetworkRNDIS} {
		if !slices.Contains(b.rebuilds[0], "rm -f "+marker) {
			t.Errorf("%s not removed: %v", marker, b.rebuilds[0])
		}
	}
}

func TestApplyReportsAFailedRebuild(t *testing.T) {
	b := newBoard()
	b.fail = errors.New("exit status 1")

	failure := applyUSBSelection(applyReq(false, true, false, usbNetworkOff, ""), b.env())
	if failure == nil || failure.code != -3 {
		t.Errorf("got %+v, want -3", failure)
	}
}

// A failed marker step skips the rest of the steps but not start, so the
// host gets its keyboard back. Nothing is undone.
func TestRunRebuildAlwaysStartsAfterAStop(t *testing.T) {
	var ran []string
	run := func(command string) error {
		ran = append(ran, command)
		if command == "touch /boot/usb.disk0" {
			return errors.New("read-only file system")
		}
		return nil
	}

	commands := []string{usbGadgetStop, "rm -f /boot/usb.acm", "touch /boot/usb.disk0", "touch /boot/usb.uac", usbGadgetStart}
	if err := runRebuild(commands, run); err == nil {
		t.Error("a failed step was not reported")
	}

	want := []string{usbGadgetStop, "rm -f /boot/usb.acm", "touch /boot/usb.disk0", usbGadgetStart}
	if !slices.Equal(ran, want) {
		t.Errorf("ran %v, want %v", ran, want)
	}
}

func TestRunRebuildStopsWhenStopFails(t *testing.T) {
	var ran []string
	run := func(command string) error {
		ran = append(ran, command)
		if command == usbGadgetStop {
			return errors.New("exit status 1")
		}
		return nil
	}

	if err := runRebuild([]string{usbGadgetStop, "touch /boot/usb.uac", usbGadgetStart}, run); err == nil {
		t.Error("a failed stop was not reported")
	}

	if !slices.Equal(ran, []string{usbGadgetStop}) {
		t.Errorf("ran %v after a failed stop", ran)
	}
}

// markerSteps must strip exactly the stop and start, so the per-device lists
// stay the single source of what turning a function on or off means.
func TestMarkerStepsKeepsEverythingButStopAndStart(t *testing.T) {
	for _, device := range []string{"console", "disk", "audio", "network"} {
		_, mount, unmount, _ := commandsFor(device)
		for _, list := range [][]string{mount, unmount} {
			steps := markerSteps(list)
			if len(steps) != len(list)-2 {
				t.Errorf("%s: %v became %v", device, list, steps)
			}
		}
	}
}
