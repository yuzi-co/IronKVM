# HID Consumer and System Control Keys Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The web UI sends media keys (Consumer Control) and Sleep, Wake and Power Down (System Control) to the managed host through the existing absolute pointer HID function.

**Architecture:** `hid.GS2` in `S03usbdev` gets report IDs: 1 for the pointer, 2 for Consumer Control, 3 for System Control, and `report_length` 7. The server reads that `report_length` from configfs when it opens `/dev/hidg2` and puts ID 1 in front of pointer reports only when it reads 7, so every existing writer keeps sending 6 bytes and a mismatched gadget and server never break the mouse. A new `POST /api/hid/key` route presses and releases one extended key. The UI shows the new buttons only when `GET /api/hid/mode` says `extendedKeys: true`.

**Tech Stack:** POSIX sh (init scripts and tool suites), Go 1.25 with gin, React + TypeScript + antd + lucide-react.

**Spec:** `docs/superpowers/specs/2026-09-26-hid-consumer-system-keys-design.md`

## Global Constraints

- Branch: `feat/hid-extended-keys` (cut from `fork/integration`). Commit there only. Do not push, do not merge.
- Commit messages carry no `Co-Authored-By` and no `Claude-Session` trailers.
- No em dashes anywhere: code, comments, commit messages, docs, UI strings. Use a colon, comma or full stop.
- Do not change `hid.GS0` (keyboard) or `hid.GS1` (relative mouse), in either script.
- Do not change `kvmapp/system/init.d/S03usbhid`. Hid-only mode keeps its 6-byte pointer with no report IDs.
- Report IDs: pointer `1`, Consumer Control `2`, System Control `3`.
- `hid.GS2` `report_length` in `S03usbdev`: `7`. `AbsoluteMouseReportLen` stays `6`.
- configfs path the server reads: `/sys/kernel/config/usb_gadget/g0/functions/hid.GS2/report_length`.
- Route: `POST /api/hid/key`, body `{"page": "consumer" | "system", "usage": <int>}`, in the `CheckToken` group. Consumer usage `1` to `0x3FF`. System usage `0x81` to `0xB7`. Press, wait 50 ms, release. The release is written even when the request context ends early.
- `GET /api/hid/mode` gains `extendedKeys` (bool, JSON name `extendedKeys`).
- Go tests run in Docker from the repository root (the usb report test reads `kvmapp/`):
  `MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W):/repo" -v nanokvm-gomod:/go/pkg/mod -w /repo/server -e CGO_ENABLED=0 golang:1.25 go test -tags novision ./service/hid/ ./router/ ./proto/`
- Shell suites: `sh tools/run-tests.sh --container usbdev`. Exit 0 = pass, 1 = fail, 2 = could not run (not a pass).
- Nothing in this plan touches the board. Deploy is Task 5 and needs the operator's word first.

## Review Focus

- A server built from this branch running on a gadget that still says `report_length` 6 (old scripts, or hid-only mode): the absolute mouse must keep working with 6-byte reports and no ID. Pinned in Task 2.
- The jiggler, MCP and the release path call `WriteAbsoluteMouseReport` with 6 bytes: they must reach the wire as 7 bytes with ID 1 when IDs are on. Pinned in Task 2 (the prefix is in the shared write path, tested through `WriteAbsoluteMouseReport`).
- An extended key report must never get the pointer's ID 1 in front of it. Pinned in Task 3.
- A request that ends between press and release must not leave the key held on the host. Pinned in Task 3 (`pressExtendedKey` test with a cancelled context).
- A `POST /api/hid/key` on a gadget without IDs must write nothing to `/dev/hidg2`. Pinned in Task 3.

---

## File Structure

| File | Responsibility |
| ---- | -------------- |
| `kvmapp/system/init.d/S03usbdev` | `hid.GS2` descriptor with report IDs, `report_length` 7 |
| `tools/usbdev/test-usb-descriptors.sh` | Expected `hid.GS2` bytes and length for normal mode |
| `server/service/hid/usb_report_test.go` | Per-script report lengths; descriptor carries the three IDs |
| `server/service/hid/hid.go` | Report ID constants, configfs read, prefixing in the write path, extended key writes |
| `server/service/hid/extended_keys.go` (new) | Report building, validation, press-and-release, HTTP handler |
| `server/service/hid/extended_keys_test.go` (new) | Tests for the above |
| `server/service/hid/report_id_test.go` (new) | Tests for the configfs read and the prefix |
| `server/service/hid/status.go` | `GetHidMode` reports `extendedKeys` |
| `server/proto/hid.go` | `GetHidModeRsp.ExtendedKeys`, `SendHidKeyReq` |
| `server/router/hid.go` | Registers `POST /api/hid/key` |
| `web/src/api/hid.ts` | `sendKey` |
| `web/src/hooks/useExtendedKeys.ts` (new) | Reads `extendedKeys` once from `/api/hid/mode` |
| `web/src/pages/desktop/menu/keyboard/media-keys.tsx` (new) | Media keys row |
| `web/src/pages/desktop/menu/keyboard/index.tsx` | Mounts the row |
| `web/src/pages/desktop/menu/power/host-power.tsx` (new) | Sleep, Wake, Power Down group |
| `web/src/pages/desktop/menu/power/index.tsx` | Mounts the group |
| `web/src/i18n/locales/en.ts` | New strings |

---

### Task 1: Gadget descriptor with report IDs

**Files:**
- Modify: `kvmapp/system/init.d/S03usbdev:884-885`
- Modify: `tools/usbdev/test-usb-descriptors.sh:93` and `:288`
- Modify: `server/service/hid/hid.go:56-71` (constants only)
- Modify: `server/service/hid/usb_report_test.go`

**Interfaces:**
- Produces: Go constants `AbsoluteMouseReportLenWithID = AbsoluteMouseReportLen + 1`, `AbsolutePointerReportID byte = 1`, `ConsumerReportID byte = 2`, `SystemReportID byte = 3`, all in `server/service/hid/hid.go`.

- [ ] **Step 1: Update the shell suite's expectations (failing test)**

In `tools/usbdev/test-usb-descriptors.sh`, replace line 93:

```sh
ABSOLUTE_MOUSE_DESC=05010902a1010901a10005091901290515002501950575018102950175038101050109300931150026ff7f350046ff7f751095028102050109381581257f35004500750895018106c0c0
```

with:

```sh
# Three report IDs on one interface: 1 is the pointer, 2 Consumer Control, 3
# System Control. The keyboard and the relative mouse claim the boot protocol
# and must never carry an ID; this interface does not, so it carries the keys.
ABSOLUTE_MOUSE_DESC=05010902a10185010901a10005091901290515002501950575018102950175038101050109300931150026ff7f350046ff7f751095028102050109381581257f35004500750895018106c0c0050c0901a1018502150026ff0319002aff03751095018100c005010980a1018503150026ff00190029ff750895018100c0
```

Leave `HID_ONLY_ABSOLUTE_DESC` (line 94) unchanged.

Replace line 288:

```sh
hid_is hid.GS2 "" 2 6 "$ABSOLUTE_MOUSE_DESC"  "absolute pointer"
```

with:

```sh
hid_is hid.GS2 "" 2 7 "$ABSOLUTE_MOUSE_DESC"  "absolute pointer"
```

Leave line 659 (`hid_is hid.GS2 "" 2 6 "$HID_ONLY_ABSOLUTE_DESC" ...`) unchanged.

- [ ] **Step 2: Update the Go report-length test (failing test)**

In `server/service/hid/usb_report_test.go`, replace the `hidFunctions` map and `TestShellAndGoAgreeOnEveryReportLength` so the expected length is per script:

```go
// hidFunctions maps each script to the report length every configfs instance
// must have. The absolute pointer differs between the two: S03usbdev gives it
// report IDs, so the host reads one ID byte and the report, and S03usbhid, the
// mode for hosts that are picky about the gadget, keeps the plain report.
var hidFunctions = map[string]map[string]int{
	"../../../kvmapp/system/init.d/S03usbdev": {
		"hid.GS0": KeyboardReportLen,
		"hid.GS1": RelativeMouseReportLen,
		"hid.GS2": AbsoluteMouseReportLenWithID,
	},
	"../../../kvmapp/system/init.d/S03usbhid": {
		"hid.GS0": KeyboardReportLen,
		"hid.GS1": RelativeMouseReportLen,
		"hid.GS2": AbsoluteMouseReportLen,
	},
}
```

and in `TestShellAndGoAgreeOnEveryReportLength` change the loop body to use `want := hidFunctions[path]`:

```go
func TestShellAndGoAgreeOnEveryReportLength(t *testing.T) {
	for _, path := range initScripts {
		script := readScript(t, path)
		found := shellReportLengths(t, script)
		wantAll := hidFunctions[path]

		if len(found) != len(wantAll) {
			t.Fatalf("%s writes %d report lengths, want %d: %v", path, len(found), len(wantAll), found)
		}

		for function, want := range wantAll {
			got, ok := found[function]
			if !ok {
				t.Errorf("%s sets no report_length for %s", path, function)
				continue
			}
			if got != want {
				t.Errorf("%s gives %s a report_length of %d, but this package sends %d",
					path, function, got, want)
			}
		}
	}
}
```

Read the existing function body first and keep any lines it has that this replacement does not show.

Add these tests and a helper at the end of the file:

```go
// absoluteDescriptor decodes the escaped bytes the script echoes into
// hid.GS2/report_desc.
func absoluteDescriptor(t *testing.T, path string, script string) []byte {
	t.Helper()

	pattern := regexp.MustCompile(`echo -ne (\S+) > functions/hid\.GS2/report_desc`)
	match := pattern.FindStringSubmatch(script)
	if match == nil {
		t.Fatalf("%s: no report_desc line for hid.GS2", path)
	}

	descriptor, err := decodeHexEscapes(strings.ReplaceAll(match[1], `\\`, `\`))
	if err != nil {
		t.Fatalf("%s: %s", path, err)
	}
	return descriptor
}

// The server puts AbsolutePointerReportID in front of every pointer report and
// sends the keys under the other two IDs. A descriptor that declared different
// numbers would have the host read clicks as volume keys.
func TestTheNormalModePointerDeclaresTheThreeReportIDs(t *testing.T) {
	path := "../../../kvmapp/system/init.d/S03usbdev"
	descriptor := absoluteDescriptor(t, path, readScript(t, path))

	// 85 NN = Report ID (NN), each right after its application collection.
	for _, want := range [][]byte{
		{0x09, 0x02, 0xa1, 0x01, 0x85, AbsolutePointerReportID},
		{0x09, 0x01, 0xa1, 0x01, 0x85, ConsumerReportID},
		{0x09, 0x80, 0xa1, 0x01, 0x85, SystemReportID},
	} {
		if !containsBytes(descriptor, want) {
			t.Errorf("%s: the absolute descriptor lacks % x", path, want)
		}
	}
}

// Hid-only mode keeps the pointer without report IDs. The server reads
// report_length to decide whether to send an ID, and this is the case where it
// must not.
func TestTheHidOnlyPointerHasNoReportIDs(t *testing.T) {
	path := "../../../kvmapp/system/init.d/S03usbhid"
	descriptor := absoluteDescriptor(t, path, readScript(t, path))

	if containsBytes(descriptor, []byte{0x85}) {
		t.Errorf("%s: the hid-only absolute descriptor declares a report ID", path)
	}
}
```

- [ ] **Step 3: Add the constants in `hid.go` so the test compiles**

In `server/service/hid/hid.go`, after `AbsoluteMouseReportLen = 6` inside the same const block, add:

```go
	// AbsoluteMouseReportLenWithID is what the host reads from the absolute
	// pointer in normal mode: one report ID byte, then the report.
	AbsoluteMouseReportLenWithID = AbsoluteMouseReportLen + 1
```

After that const block, add:

```go
// Report IDs on the absolute pointer. Only S03usbdev declares them; the
// keyboard and the relative mouse claim the boot protocol, where a firmware
// reads reports without an ID, so they never carry one.
const (
	AbsolutePointerReportID byte = 1
	ConsumerReportID        byte = 2
	SystemReportID          byte = 3
)
```

- [ ] **Step 4: Run both tests to verify they fail**

Run: `sh tools/run-tests.sh --container usbdev`
Expected: exit 1, with the "absolute pointer" cases failing on the descriptor and on `report_length`.

Run the Go command from Global Constraints, restricted: `... go test -tags novision ./service/hid/ -run 'ReportLength|ReportIDs'`
Expected: FAIL. `S03usbdev gives hid.GS2 a report_length of 6, but this package sends 7` and `lacks 85 01`-style messages.

- [ ] **Step 5: Change the descriptor in `S03usbdev`**

In `kvmapp/system/init.d/S03usbdev`, replace line 884:

```sh
        echo 6 > functions/hid.GS2/report_length
```

with:

```sh
        # One report ID byte, then the largest report (the pointer's six). The
        # keys share this interface under IDs 2 and 3 rather than taking an
        # endpoint of their own, and the keyboard and relative mouse stay
        # without IDs because a firmware reading them in boot protocol expects
        # none. The server reads this number to decide whether to send an ID.
        echo 7 > functions/hid.GS2/report_length
```

Replace line 885 (the `hid.GS2/report_desc` line) with:

```sh
        echo -ne \\x05\\x01\\x09\\x02\\xa1\\x01\\x85\\x01\\x09\\x01\\xa1\\x00\\x05\\x09\\x19\\x01\\x29\\x05\\x15\\x00\\x25\\x01\\x95\\x05\\x75\\x01\\x81\\x02\\x95\\x01\\x75\\x03\\x81\\x01\\x05\\x01\\x09\\x30\\x09\\x31\\x15\\x00\\x26\\xff\\x7f\\x35\\x00\\x46\\xff\\x7f\\x75\\x10\\x95\\x02\\x81\\x02\\x05\\x01\\x09\\x38\\x15\\x81\\x25\\x7f\\x35\\x00\\x45\\x00\\x75\\x08\\x95\\x01\\x81\\x06\\xc0\\xc0\\x05\\x0c\\x09\\x01\\xa1\\x01\\x85\\x02\\x15\\x00\\x26\\xff\\x03\\x19\\x00\\x2a\\xff\\x03\\x75\\x10\\x95\\x01\\x81\\x00\\xc0\\x05\\x01\\x09\\x80\\xa1\\x01\\x85\\x03\\x15\\x00\\x26\\xff\\x00\\x19\\x00\\x29\\xff\\x75\\x08\\x95\\x01\\x81\\x00\\xc0 > functions/hid.GS2/report_desc
```

Keep the existing 8-space indentation of those lines. Do not touch `S03usbhid`.

- [ ] **Step 6: Run both tests to verify they pass**

Run: `sh tools/run-tests.sh --container usbdev`
Expected: exit 0.

Run: `... go test -tags novision ./service/hid/`
Expected: `ok`. The server still sends 6-byte pointer reports; nothing reads `report_length` yet.

- [ ] **Step 7: Commit**

```bash
git add kvmapp/system/init.d/S03usbdev tools/usbdev/test-usb-descriptors.sh server/service/hid/hid.go server/service/hid/usb_report_test.go
git commit -m "S03usbdev: give the absolute pointer report IDs for Consumer and System Control"
```

---

### Task 2: The server prefixes pointer reports when the gadget has IDs

**Files:**
- Modify: `server/service/hid/hid.go` (struct `Hid`, struct `hidDevice`, `absoluteMouseDevice`, `openDeviceNoLock`, `writeHIDLocked`)
- Create: `server/service/hid/report_id_test.go`

**Interfaces:**
- Consumes: `AbsoluteMouseReportLenWithID`, `AbsolutePointerReportID` from Task 1.
- Produces:
  - `var absoluteReportLengthPath string` (package var, tests override it).
  - `func readAbsoluteReportID() byte`: `AbsolutePointerReportID` when the file holds `7`, else `0`.
  - `hidDevice.idPrefix bool`; `Hid.absReportID byte`; `Hid.absBuf [AbsoluteMouseReportLenWithID]byte`.
  - `func (h *Hid) extendedKeyDevice() hidDevice`: the `/dev/hidg2` handle, lock and health, with `idPrefix` false.

- [ ] **Step 1: Write the failing tests**

Create `server/service/hid/report_id_test.go`:

```go
package hid

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// recordWrites captures every report that reaches the wire.
func recordWrites(t *testing.T) *[][]byte {
	t.Helper()

	restore := writeReport
	t.Cleanup(func() { writeReport = restore })

	var got [][]byte
	writeReport = func(_ string, _ *os.File, data []byte, _ time.Duration) error {
		got = append(got, append([]byte(nil), data...))
		return nil
	}
	return &got
}

// gadgetReportLength points the configfs read at a scratch file holding value,
// or at a missing file when value is empty.
func gadgetReportLength(t *testing.T, value string) {
	t.Helper()

	restore := absoluteReportLengthPath
	t.Cleanup(func() { absoluteReportLengthPath = restore })

	path := filepath.Join(t.TempDir(), "report_length")
	if value != "" {
		if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	absoluteReportLengthPath = path
}

func TestReadAbsoluteReportID(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  byte
	}{
		{"7\n", AbsolutePointerReportID},
		{"7", AbsolutePointerReportID},
		{"6\n", 0},
		{"", 0},
		{"garbage\n", 0},
	} {
		gadgetReportLength(t, tc.value)
		if got := readAbsoluteReportID(); got != tc.want {
			t.Errorf("report_length %q: got ID %d, want %d", tc.value, got, tc.want)
		}
	}
}

func TestOpeningThePointerReadsTheReportID(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  byte
	}{
		{"7\n", AbsolutePointerReportID},
		{"6\n", 0},
		{"", 0},
	} {
		gadgetReportLength(t, tc.value)
		h := &Hid{absReportID: 0xee}

		node := filepath.Join(t.TempDir(), "hidg2")
		if err := os.WriteFile(node, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := h.openDeviceNoLock(h.absoluteMouseDevice(node)); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { h.closeDeviceNoLock(h.absoluteMouseDevice(node)) })

		if h.absReportID != tc.want {
			t.Errorf("report_length %q: absReportID %d, want %d", tc.value, h.absReportID, tc.want)
		}
	}
}

func TestOpeningTheKeyDeviceAlsoReadsTheReportID(t *testing.T) {
	gadgetReportLength(t, "7\n")
	h := &Hid{}

	node := filepath.Join(t.TempDir(), "hidg2")
	if err := os.WriteFile(node, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	device := h.extendedKeyDevice()
	device.path = node
	if err := h.openDeviceNoLock(device); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { h.closeDeviceNoLock(device) })

	if h.absReportID != AbsolutePointerReportID {
		t.Fatalf("absReportID %d after opening through the key device, want %d", h.absReportID, AbsolutePointerReportID)
	}
}

func TestPointerReportsGetTheIDWhenTheGadgetHasIDs(t *testing.T) {
	got := recordWrites(t)
	h := &Hid{absReportID: AbsolutePointerReportID}
	openDevice(t, h.absoluteMouseDevice(HID2))

	report := []byte{0x01, 0x34, 0x12, 0x78, 0x56, 0xff}
	if err := h.WriteAbsoluteMouseReport(report); err != nil {
		t.Fatal(err)
	}

	want := []byte{AbsolutePointerReportID, 0x01, 0x34, 0x12, 0x78, 0x56, 0xff}
	if len(*got) != 1 || !bytes.Equal((*got)[0], want) {
		t.Fatalf("wrote %x, want one report %x", *got, want)
	}
}

func TestPointerReportsStayPlainWithoutIDs(t *testing.T) {
	got := recordWrites(t)
	h := &Hid{}
	openDevice(t, h.absoluteMouseDevice(HID2))

	report := []byte{0x01, 0x34, 0x12, 0x78, 0x56, 0xff}
	if err := h.WriteAbsoluteMouseReport(report); err != nil {
		t.Fatal(err)
	}

	if len(*got) != 1 || !bytes.Equal((*got)[0], report) {
		t.Fatalf("wrote %x, want one report %x", *got, report)
	}
}

func TestThePrefixDoesNotAllocate(t *testing.T) {
	recordWritesNoCopy := writeReport
	t.Cleanup(func() { writeReport = recordWritesNoCopy })
	writeReport = func(string, *os.File, []byte, time.Duration) error { return nil }

	h := &Hid{absReportID: AbsolutePointerReportID}
	openDevice(t, h.absoluteMouseDevice(HID2))
	report := []byte{0, 0, 0, 0, 0, 0}

	allocs := testing.AllocsPerRun(200, func() {
		_ = h.WriteAbsoluteMouseReport(report)
	})
	if allocs != 0 {
		t.Fatalf("a prefixed pointer report allocates %.0f times, want 0", allocs)
	}
}
```

If `TestThePrefixDoesNotAllocate` fails on the unmodified tree with a non-zero count, the write path already allocates for another reason. Then change the test to measure the count before your change and assert that the prefix adds none, and say so in the report.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `... go test -tags novision ./service/hid/ -run 'ReportID|Prefix|PointerReports|KeyDevice'`
Expected: build failure, `undefined: absoluteReportLengthPath`, `readAbsoluteReportID`, `absReportID`, `extendedKeyDevice`.

- [ ] **Step 3: Implement**

In `server/service/hid/hid.go`:

Add to `type Hid struct`, after `absHealth endpointHealth`:

```go
	// absReportID is the ID put in front of every absolute pointer report, or
	// 0 for none. It is read from configfs each time /dev/hidg2 is opened,
	// because the gadget and this binary are deployed separately, and a
	// pointer that sends the wrong shape does not move: f_hid cuts a 7-byte
	// write to 6, and a host expecting an ID misreads a 6-byte report.
	// Guarded by mouseMutex.
	absReportID byte
	// absBuf holds one prefixed pointer report, so the prefix costs no
	// allocation on the hottest path in the server. Guarded by mouseMutex.
	absBuf [AbsoluteMouseReportLenWithID]byte
```

Add to `type hidDevice struct`, after `health *endpointHealth`:

```go
	// idPrefix asks writeHID to put Hid.absReportID in front of the report.
	// Only the pointer sets it; the keys on the same endpoint carry their own
	// ID in the report.
	idPrefix bool
```

Replace `absoluteMouseDevice` and add `extendedKeyDevice` below it:

```go
func (h *Hid) absoluteMouseDevice(path string) hidDevice {
	return hidDevice{path: path, name: NameAbsoluteMouse, mu: &h.mouseMutex, file: &h.g2, health: &h.absHealth, idPrefix: true}
}

// extendedKeyDevice is the absolute pointer's endpoint seen by the Consumer and
// System Control keys. It shares the handle, the lock and the health record,
// and it writes reports as given.
func (h *Hid) extendedKeyDevice() hidDevice {
	return hidDevice{path: HID2, name: NameAbsoluteMouse, mu: &h.mouseMutex, file: &h.g2, health: &h.absHealth}
}
```

Add after the `hidReopen...` const block:

```go
// absoluteReportLengthPath is where S03usbdev records the absolute pointer's
// report length. A variable so tests can point it at a scratch file.
var absoluteReportLengthPath = "/sys/kernel/config/usb_gadget/g0/functions/hid.GS2/report_length"

// readAbsoluteReportID returns the pointer's report ID if the gadget declares
// report IDs on hid.GS2, and 0 if it does not. The length is the tell: only
// the descriptor with IDs is AbsoluteMouseReportLenWithID long. A missing or
// unreadable file means no IDs, which is what an older gadget and hid-only
// mode both need.
func readAbsoluteReportID() byte {
	raw, err := os.ReadFile(absoluteReportLengthPath)
	if err != nil {
		return 0
	}
	if string(bytes.TrimSpace(raw)) == strconv.Itoa(AbsoluteMouseReportLenWithID) {
		return AbsolutePointerReportID
	}
	return 0
}
```

In `openDeviceNoLock`, replace:

```go
	device.set(file)
	return nil
```

with:

```go
	device.set(file)
	if device.file == &h.g2 {
		h.absReportID = readAbsoluteReportID()
	}
	return nil
```

In `writeHIDLocked`, directly before `if err := writeReport(device.path, file, data, hidWriteTimeout); err != nil {`, add:

```go
	if device.idPrefix && h.absReportID != 0 {
		if len(data) != AbsoluteMouseReportLen {
			return fmt.Errorf("%s: pointer report of %d bytes, want %d", device.path, len(data), AbsoluteMouseReportLen)
		}
		h.absBuf[0] = h.absReportID
		copy(h.absBuf[1:], data)
		data = h.absBuf[:]
	}
```

Confirm `bytes` and `strconv` are already imported in `hid.go` (they are at the top of the file).

- [ ] **Step 4: Run the whole package**

Run: `... go test -tags novision ./service/hid/`
Expected: `ok`. `TestBuildingADeviceDoesNotAllocate` must still pass.

Also run: `MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W):/repo" -v nanokvm-gomod:/go/pkg/mod -w /repo/server -e CGO_ENABLED=0 golang:1.25 go vet -tags novision ./service/hid/`
Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add server/service/hid/hid.go server/service/hid/report_id_test.go
git commit -m "hid: put the pointer's report ID in front of absolute reports when the gadget declares one"
```

---

### Task 3: Extended key reports, the route and the mode flag

**Files:**
- Create: `server/service/hid/extended_keys.go`
- Create: `server/service/hid/extended_keys_test.go`
- Modify: `server/proto/hid.go`
- Modify: `server/service/hid/status.go` (`GetHidMode`)
- Modify: `server/router/hid.go`

**Interfaces:**
- Consumes: `ConsumerReportID`, `SystemReportID` (Task 1); `readAbsoluteReportID`, `extendedKeyDevice`, `writeHID` (Task 2); `sleepPasteContext` (existing, `paste.go`); `inputcontrol.ManualAbsoluteMouse` (existing).
- Produces:
  - `func extendedKeyReports(page string, usage int) (press, release []byte, err error)`
  - `func (h *Hid) ExtendedKeysAvailable() bool`
  - `func (h *Hid) WriteExtendedKeyReport(report []byte) error`
  - `var errExtendedKeysUnavailable error`
  - `func pressExtendedKey(ctx context.Context, write func([]byte) error, press, release []byte, hold time.Duration) error`
  - `func (s *Service) SendKey(c *gin.Context)`
  - `proto.SendHidKeyReq{Page string; Usage int}`, `proto.GetHidModeRsp.ExtendedKeys bool`

- [ ] **Step 1: Write the failing tests**

Create `server/service/hid/extended_keys_test.go`:

```go
package hid

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"
)

func TestExtendedKeyReports(t *testing.T) {
	for _, tc := range []struct {
		page           string
		usage          int
		press, release []byte
	}{
		{"consumer", 0xe9, []byte{ConsumerReportID, 0xe9, 0x00}, []byte{ConsumerReportID, 0, 0}},
		{"consumer", 0x223, []byte{ConsumerReportID, 0x23, 0x02}, []byte{ConsumerReportID, 0, 0}},
		{"consumer", 0x3ff, []byte{ConsumerReportID, 0xff, 0x03}, []byte{ConsumerReportID, 0, 0}},
		{"consumer", 1, []byte{ConsumerReportID, 0x01, 0x00}, []byte{ConsumerReportID, 0, 0}},
		{"system", 0x81, []byte{SystemReportID, 0x81}, []byte{SystemReportID, 0}},
		{"system", 0x82, []byte{SystemReportID, 0x82}, []byte{SystemReportID, 0}},
		{"system", 0xb7, []byte{SystemReportID, 0xb7}, []byte{SystemReportID, 0}},
	} {
		press, release, err := extendedKeyReports(tc.page, tc.usage)
		if err != nil {
			t.Errorf("%s 0x%x: %s", tc.page, tc.usage, err)
			continue
		}
		if !bytes.Equal(press, tc.press) || !bytes.Equal(release, tc.release) {
			t.Errorf("%s 0x%x: press %x release %x, want %x %x", tc.page, tc.usage, press, release, tc.press, tc.release)
		}
	}
}

func TestExtendedKeyReportsRejectBadInput(t *testing.T) {
	for _, tc := range []struct {
		page  string
		usage int
	}{
		{"consumer", 0},
		{"consumer", -1},
		{"consumer", 0x400},
		{"system", 0x80},
		{"system", 0xb8},
		{"system", 0},
		{"keyboard", 0x04},
		{"", 0xe9},
	} {
		if _, _, err := extendedKeyReports(tc.page, tc.usage); err == nil {
			t.Errorf("%q 0x%x was accepted", tc.page, tc.usage)
		}
	}
}

func TestKeyReportsCarryNoPointerID(t *testing.T) {
	gadgetReportLength(t, "7\n")
	got := recordWrites(t)
	h := &Hid{absReportID: AbsolutePointerReportID}
	openDevice(t, h.extendedKeyDevice())

	report := []byte{ConsumerReportID, 0xe9, 0x00}
	if err := h.WriteExtendedKeyReport(report); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 || !bytes.Equal((*got)[0], report) {
		t.Fatalf("wrote %x, want exactly %x", *got, report)
	}
}

func TestKeyReportsNeedAGadgetWithIDs(t *testing.T) {
	for _, value := range []string{"6\n", ""} {
		gadgetReportLength(t, value)
		got := recordWrites(t)
		h := &Hid{}
		openDevice(t, h.extendedKeyDevice())

		err := h.WriteExtendedKeyReport([]byte{ConsumerReportID, 0xe9, 0x00})
		if !errors.Is(err, errExtendedKeysUnavailable) {
			t.Errorf("report_length %q: err %v, want errExtendedKeysUnavailable", value, err)
		}
		if len(*got) != 0 {
			t.Errorf("report_length %q: wrote %x, want nothing", value, *got)
		}
		if h.ExtendedKeysAvailable() {
			t.Errorf("report_length %q: ExtendedKeysAvailable is true", value)
		}
	}
}

func TestKeyReportsRejectAnythingButTheKeyIDs(t *testing.T) {
	gadgetReportLength(t, "7\n")
	got := recordWrites(t)
	h := &Hid{}
	openDevice(t, h.extendedKeyDevice())

	for _, report := range [][]byte{
		{AbsolutePointerReportID, 0, 0, 0, 0, 0, 0},
		{ConsumerReportID, 0xe9},
		{SystemReportID, 0x82, 0x00},
		{},
	} {
		if err := h.WriteExtendedKeyReport(report); err == nil {
			t.Errorf("% x was accepted", report)
		}
	}
	if len(*got) != 0 {
		t.Fatalf("wrote %x, want nothing", *got)
	}
}

func TestPressExtendedKeyPressesThenReleases(t *testing.T) {
	var got [][]byte
	write := func(r []byte) error { got = append(got, r); return nil }
	press, release := []byte{ConsumerReportID, 0xe9, 0}, []byte{ConsumerReportID, 0, 0}

	if err := pressExtendedKey(context.Background(), write, press, release, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || !bytes.Equal(got[0], press) || !bytes.Equal(got[1], release) {
		t.Fatalf("wrote %x, want press then release", got)
	}
}

func TestPressExtendedKeyReleasesWhenTheRequestEnds(t *testing.T) {
	var got [][]byte
	write := func(r []byte) error { got = append(got, r); return nil }
	press, release := []byte{SystemReportID, 0x82}, []byte{SystemReportID, 0}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := pressExtendedKey(ctx, write, press, release, time.Hour)

	if err == nil {
		t.Fatal("a cancelled request reported success")
	}
	if len(got) != 2 || !bytes.Equal(got[1], release) {
		t.Fatalf("wrote %x, want the release after the press", got)
	}
}

func TestPressExtendedKeySkipsTheWaitWhenThePressFails(t *testing.T) {
	var got [][]byte
	fail := errors.New("stalled")
	write := func(r []byte) error {
		got = append(got, r)
		if len(got) == 1 {
			return fail
		}
		return nil
	}
	press, release := []byte{ConsumerReportID, 0xe9, 0}, []byte{ConsumerReportID, 0, 0}

	start := time.Now()
	err := pressExtendedKey(context.Background(), write, press, release, time.Hour)
	if !errors.Is(err, fail) {
		t.Fatalf("err %v, want the press error", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("waited out the hold after a failed press")
	}
	if len(got) != 2 || !bytes.Equal(got[1], release) {
		t.Fatalf("wrote %x, want a release after the failed press", got)
	}
}
```

`recordWrites`, `gadgetReportLength` and `openDevice` come from `report_id_test.go` (Task 2) and `health_wiring_test.go`.

- [ ] **Step 2: Run the tests to verify they fail**

Run: `... go test -tags novision ./service/hid/ -run 'ExtendedKey|KeyReports|PressExtended'`
Expected: build failure, `undefined: extendedKeyReports`, `pressExtendedKey`, `errExtendedKeysUnavailable`.

- [ ] **Step 3: Add the proto types**

In `server/proto/hid.go`, replace `GetHidModeRsp` with:

```go
type GetHidModeRsp struct {
	Mode string `json:"mode"` // normal or hid-only
	// ExtendedKeys is true when the gadget carries the Consumer and System
	// Control reports, which only normal mode's descriptor declares.
	ExtendedKeys bool `json:"extendedKeys"`
}
```

and add after it:

```go
// SendHidKeyReq presses and releases one key outside the keyboard: a
// Consumer Control usage (media, volume) or a System Control usage (power,
// sleep, wake).
type SendHidKeyReq struct {
	Page  string `json:"page" form:"page" validate:"required,oneof=consumer system"`
	Usage int    `json:"usage" form:"usage" validate:"required"`
}
```

- [ ] **Step 4: Implement `extended_keys.go`**

Create `server/service/hid/extended_keys.go`:

```go
package hid

import (
	"context"
	"errors"
	"fmt"
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/inputcontrol"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Consumer and System Control keys share the absolute pointer's endpoint under
// their own report IDs. S03usbdev declares them; hid-only mode and an older
// gadget do not, and there a key report would reach the host as a garbled
// pointer report, so nothing is written.

const (
	extendedKeyHold = 50 * time.Millisecond

	consumerUsageMax = 0x3ff
	systemUsageMin   = 0x81
	systemUsageMax   = 0xb7
)

var errExtendedKeysUnavailable = errors.New("the USB gadget has no Consumer or System Control reports")

// extendedKeyReports builds the press and release reports for one key.
func extendedKeyReports(page string, usage int) (press, release []byte, err error) {
	switch page {
	case "consumer":
		if usage < 1 || usage > consumerUsageMax {
			return nil, nil, fmt.Errorf("consumer usage 0x%x is outside 0x1 to 0x%x", usage, consumerUsageMax)
		}
		return []byte{ConsumerReportID, byte(usage), byte(usage >> 8)}, []byte{ConsumerReportID, 0, 0}, nil
	case "system":
		if usage < systemUsageMin || usage > systemUsageMax {
			return nil, nil, fmt.Errorf("system usage 0x%x is outside 0x%x to 0x%x", usage, systemUsageMin, systemUsageMax)
		}
		return []byte{SystemReportID, byte(usage)}, []byte{SystemReportID, 0}, nil
	default:
		return nil, nil, fmt.Errorf("unknown key page %q", page)
	}
}

// ExtendedKeysAvailable reports whether the gadget declares the key reports.
// It reads configfs rather than the state kept for the open handle, because a
// caller may ask before anything has opened /dev/hidg2.
func (h *Hid) ExtendedKeysAvailable() bool {
	return readAbsoluteReportID() != 0
}

// WriteExtendedKeyReport writes one Consumer or System Control report as given.
func (h *Hid) WriteExtendedKeyReport(report []byte) error {
	switch {
	case len(report) == 3 && report[0] == ConsumerReportID:
	case len(report) == 2 && report[0] == SystemReportID:
	default:
		return fmt.Errorf("not a Consumer or System Control report: % x", report)
	}
	if !h.ExtendedKeysAvailable() {
		return errExtendedKeysUnavailable
	}
	return h.writeHID(h.extendedKeyDevice(), report)
}

// pressExtendedKey writes press, holds, and writes release. The release is
// written whatever happened before it: a key left down on the host repeats,
// and for Power Down or Sleep a held key is worse than a lost one.
func pressExtendedKey(ctx context.Context, write func([]byte) error, press, release []byte, hold time.Duration) error {
	err := write(press)
	if err == nil {
		err = sleepPasteContext(ctx, hold)
	}
	if releaseErr := write(release); err == nil {
		err = releaseErr
	}
	return err
}

// SendKey presses and releases one Consumer or System Control key.
func (s *Service) SendKey(c *gin.Context) {
	var req proto.SendHidKeyReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	press, release, err := extendedKeyReports(req.Page, req.Usage)
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}

	if !s.hid.ExtendedKeysAvailable() {
		rsp.ErrRsp(c, -2, "extended keys are not available in this USB mode")
		return
	}

	manual := s.newManualSession()
	defer manual.Close()
	reservation, err := manual.Reserve(c.Request.Context(), inputcontrol.ManualAbsoluteMouse, false, nil)
	if err != nil {
		log.Errorf("hid key failed to acquire HID control: %v", err)
		rsp.ErrRsp(c, -3, "HID control is busy")
		return
	}

	write := func(report []byte) error {
		return manual.Execute(func() error {
			return s.hid.WriteExtendedKeyReport(report)
		})
	}

	err = pressExtendedKey(c.Request.Context(), write, press, release, extendedKeyHold)
	reservation.Complete(err == nil)
	if err != nil {
		reportWriteFailure("hid key failed", err)
		rsp.ErrRsp(c, -3, "HID key failed")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("hid key %s 0x%x sent", req.Page, req.Usage)
}
```

Check `manual.Execute` and `reservation.Complete` against `service/hid/paste.go` lines 200-240 and `service/inputcontrol/coordinator.go`; use them exactly as `Paste` does. If `ParseFormRequest` rejects `usage` when validation tag `required` meets a valid value, keep the tag: every valid usage is non-zero.

- [ ] **Step 5: Report the flag from `GetHidMode`**

In `server/service/hid/status.go`, in `GetHidMode`, replace:

```go
	rsp.OkRspWithData(c, &proto.GetHidModeRsp{
		Mode: mode,
	})
```

with:

```go
	rsp.OkRspWithData(c, &proto.GetHidModeRsp{
		Mode:         mode,
		ExtendedKeys: GetHid().ExtendedKeysAvailable(),
	})
```

- [ ] **Step 6: Register the route**

In `server/router/hid.go`, after the line `api.POST("/hid/paste", service.Paste) // paste`, add:

```go
	api.POST("/hid/key", service.SendKey) // press one Consumer or System Control key
```

- [ ] **Step 7: Run the package, the router and vet**

Run: `... go test -tags novision ./service/hid/ ./router/ ./proto/`
Expected: `ok` for each (`no test files` is fine for a package without tests).

Run: `MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W):/repo" -v nanokvm-gomod:/go/pkg/mod -w /repo/server -e CGO_ENABLED=0 golang:1.25 sh -c 'go vet -tags novision ./... && GOOS=linux GOARCH=riscv64 go build -tags novision ./...'`
Expected: no output.

If `router` has a test that lists every route or every loopback-allowed path, update its expectation for `/api/hid/key` and say so in the report.

- [ ] **Step 8: Commit**

```bash
git add server/service/hid/extended_keys.go server/service/hid/extended_keys_test.go server/proto/hid.go server/service/hid/status.go server/router/hid.go
git commit -m "hid: POST /api/hid/key presses one Consumer or System Control key"
```

---

### Task 4: Web UI

**Files:**
- Modify: `web/src/api/hid.ts`
- Create: `web/src/hooks/useExtendedKeys.ts`
- Create: `web/src/pages/desktop/menu/keyboard/media-keys.tsx`
- Modify: `web/src/pages/desktop/menu/keyboard/index.tsx`
- Create: `web/src/pages/desktop/menu/power/host-power.tsx`
- Modify: `web/src/pages/desktop/menu/power/index.tsx`
- Modify: `web/src/i18n/locales/en.ts`

**Interfaces:**
- Consumes: `POST /api/hid/key` and `GET /api/hid/mode` `data.extendedKeys` (Task 3).
- Produces: `sendKey(page: 'consumer' | 'system', usage: number)`; `useExtendedKeys(): boolean`.

There is no frontend test runner. Verification is `pnpm lint` and `pnpm build`.

- [ ] **Step 1: API function**

Append to `web/src/api/hid.ts`:

```ts
// press and release one Consumer Control (media) or System Control (power) key
export function sendKey(page: 'consumer' | 'system', usage: number) {
  return http.post('/api/hid/key', { page, usage });
}
```

- [ ] **Step 2: Hook**

Create `web/src/hooks/useExtendedKeys.ts`:

```ts
import { useEffect, useState } from 'react';

import { getHidMode } from '@/api/hid.ts';

// The media and power keys exist only when the USB gadget declares them, which
// normal mode does and hid-only mode does not. The server reads that from the
// gadget, so the buttons are hidden rather than left to fail.
export function useExtendedKeys(): boolean {
  const [available, setAvailable] = useState(false);

  useEffect(() => {
    let active = true;
    getHidMode()
      .then((rsp) => {
        if (active && rsp.code === 0) {
          setAvailable(rsp.data.extendedKeys === true);
        }
      })
      .catch(() => {});
    return () => {
      active = false;
    };
  }, []);

  return available;
}
```

- [ ] **Step 3: Strings**

In `web/src/i18n/locales/en.ts`, inside `keyboard: {` after `dropdownRussian: 'Russian',`, add:

```ts
      mediaKeys: {
        title: 'Media keys',
        mute: 'Mute',
        volumeDown: 'Volume down',
        volumeUp: 'Volume up',
        previous: 'Previous track',
        playPause: 'Play or pause',
        next: 'Next track',
        stop: 'Stop'
      },
```

Inside `power: {` after `cancelBtn: 'No'`, change `cancelBtn: 'No'` to `cancelBtn: 'No',` and add:

```ts
      hostOs: 'Host OS',
      hostOsTip: 'Sent as USB keys. The host decides what they do.',
      sleep: 'Sleep',
      wake: 'Wake',
      powerDown: 'Power down',
      sleepConfirm: 'Put the host to sleep?',
      powerDownConfirm: 'Send the power-down key to the host?'
```

- [ ] **Step 4: Media keys row**

Create `web/src/pages/desktop/menu/keyboard/media-keys.tsx`:

```tsx
import { Tooltip } from 'antd';
import {
  PauseIcon,
  SkipBackIcon,
  SkipForwardIcon,
  SquareIcon,
  Volume1Icon,
  Volume2Icon,
  VolumeXIcon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { sendKey } from '@/api/hid.ts';
import { useExtendedKeys } from '@/hooks/useExtendedKeys.ts';

// Consumer page usages (HID Usage Tables, section 15).
const keys = [
  { usage: 0xe2, label: 'keyboard.mediaKeys.mute', Icon: VolumeXIcon },
  { usage: 0xea, label: 'keyboard.mediaKeys.volumeDown', Icon: Volume1Icon },
  { usage: 0xe9, label: 'keyboard.mediaKeys.volumeUp', Icon: Volume2Icon },
  { usage: 0xb6, label: 'keyboard.mediaKeys.previous', Icon: SkipBackIcon },
  { usage: 0xcd, label: 'keyboard.mediaKeys.playPause', Icon: PauseIcon },
  { usage: 0xb5, label: 'keyboard.mediaKeys.next', Icon: SkipForwardIcon },
  { usage: 0xb7, label: 'keyboard.mediaKeys.stop', Icon: SquareIcon }
];

export const MediaKeys = () => {
  const { t } = useTranslation();
  const available = useExtendedKeys();

  if (!available) return null;

  return (
    <div className="px-3 py-1.5">
      <div className="pb-1 text-xs text-neutral-400">{t('keyboard.mediaKeys.title')}</div>
      <div className="flex items-center space-x-1">
        {keys.map(({ usage, label, Icon }) => (
          <Tooltip key={usage} title={t(label)} placement="bottom">
            <div
              className="flex h-7 w-7 cursor-pointer items-center justify-center rounded hover:bg-neutral-700/70"
              onClick={() => sendKey('consumer', usage)}
            >
              <Icon size={16} />
            </div>
          </Tooltip>
        ))}
      </div>
    </div>
  );
};
```

If any of those icon names is not exported by the installed `lucide-react` (the build will say), pick the nearest exported icon and name it in the report.

In `web/src/pages/desktop/menu/keyboard/index.tsx`, add `import { MediaKeys } from './media-keys.tsx';` with the other local imports (keep the order prettier enforces), and add `<MediaKeys />` after `<Shortcuts />`.

- [ ] **Step 5: Host OS power group**

Create `web/src/pages/desktop/menu/power/host-power.tsx`:

```tsx
import { Divider, Popconfirm, Tooltip } from 'antd';
import { MoonIcon, PowerOffIcon, SunIcon } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { sendKey } from '@/api/hid.ts';
import { useExtendedKeys } from '@/hooks/useExtendedKeys.ts';

type HostPowerProps = {
  showConfirm: boolean;
};

// Generic Desktop System Control usages.
const SYSTEM_POWER_DOWN = 0x81;
const SYSTEM_SLEEP = 0x82;
const SYSTEM_WAKE_UP = 0x83;

export const HostPower = ({ showConfirm }: HostPowerProps) => {
  const { t } = useTranslation();
  const available = useExtendedKeys();

  if (!available) return null;

  const item = (Icon: LucideIcon, label: string, usage: number, confirm?: string) => {
    const row = (
      <div
        className="flex cursor-pointer select-none items-center space-x-2 rounded px-3 py-1.5 hover:bg-neutral-700/70"
        onClick={showConfirm && confirm ? undefined : () => sendKey('system', usage)}
      >
        <Icon size={16} />
        <span>{label}</span>
      </div>
    );

    if (!showConfirm || !confirm) return row;

    return (
      <Popconfirm
        placement="bottomLeft"
        title={confirm}
        okText={t('power.okBtn')}
        cancelText={t('power.cancelBtn')}
        onConfirm={() => sendKey('system', usage)}
        color="#404040"
      >
        {row}
      </Popconfirm>
    );
  };

  return (
    <>
      <Divider style={{ margin: '10px 0' }} />
      <Tooltip title={t('power.hostOsTip')} placement="right">
        <div className="px-1 pb-1 text-xs text-neutral-400">{t('power.hostOs')}</div>
      </Tooltip>
      <div className="flex flex-col space-y-1">
        {item(MoonIcon, t('power.sleep'), SYSTEM_SLEEP, t('power.sleepConfirm'))}
        {item(SunIcon, t('power.wake'), SYSTEM_WAKE_UP)}
        {item(PowerOffIcon, t('power.powerDown'), SYSTEM_POWER_DOWN, t('power.powerDownConfirm'))}
      </div>
    </>
  );
};
```

Wake has no confirmation: it cannot harm a running host.

In `web/src/pages/desktop/menu/power/index.tsx`, add `import { HostPower } from './host-power.tsx';` with the other local imports, and after the closing `</div>` of the `flex flex-col space-y-1` block that holds `<Reset .../>`, `<PowerShort .../>` and `<PowerLong .../>`, add:

```tsx
      <HostPower showConfirm={showConfirm} />
```

so it sits inside the outer `min-w-[200px]` div.

- [ ] **Step 6: Lint, format and build**

Run in `web/`: `pnpm format && pnpm lint && pnpm build`
Expected: all three succeed. `pnpm format` may reorder imports; keep its result.

Then run `git diff --stat -- web/pnpm-workspace.yaml web/package.json web/pnpm-lock.yaml` and confirm none of them changed. If one did, revert it and say so in the report.

- [ ] **Step 7: Commit**

```bash
git add web/src/api/hid.ts web/src/hooks/useExtendedKeys.ts web/src/pages/desktop/menu/keyboard web/src/pages/desktop/menu/power web/src/i18n/locales/en.ts
git commit -m "web: media keys in the keyboard menu, host sleep, wake and power down in the power menu"
```

Do not commit `web/dist`.

---

### Task 5: Deploy and verify on the board (controller only, operator-gated)

This task is not dispatched to a subagent. It needs the operator's word before each board step, and SSH on the board must be on.

- [ ] **Step 1: Build**

Build the server (`make app` with `DOCKER_TTY=`, then the `patchelf` RPATH step from AGENTS.md) and the web UI (`pnpm build` in `web/`).

- [ ] **Step 2: Ask the operator**, then stage and deploy

1. `scp` `kvmapp/system/init.d/S03usbdev` to both `/etc/init.d/S03usbdev` and `/kvmapp/system/init.d/S03usbdev` on root@10.0.0.222. Upload each beside the target and `mv` it into place.
2. Deploy the server and web UI with `tools/deploy/deploy-server`. Stage on `/data`, `DEPLOY_TIMEOUT=240`, run detached, and verify from `/proc/PID/exe` (`tools/deploy/check-deployed`).
3. Confirm the server still sends 6-byte pointer reports: `GET /api/hid/mode` shows `extendedKeys: false`, and the operator confirms the mouse works.
4. Run `/etc/init.d/S03usbdev stop_start`. Read the `usb watchdog` lines before touching anything else.

- [ ] **Step 3: Verify**

- `cat /sys/kernel/config/usb_gadget/g0/functions/hid.GS2/report_length` prints `7`.
- `GET /api/hid/mode` shows `extendedKeys: true`.
- On unraid, `/proc/bus/input/devices` lists the NanoKVM's "Consumer Control" and "System Control" inputs, and their `B: KEY=` bitmaps include `KEY_VOLUMEUP` (115) and `KEY_SLEEP` (142).
- The operator confirms the absolute mouse moves and clicks.
- Volume Up from the UI produces an event on the host's Consumer Control input.
- Sleep and Power Down are tested on the host only if the operator asks.

- [ ] **Step 4: Record**

Update the memory note `deploying-a-hid-descriptor-change.md` if anything in this deploy differed from it. Then merge `feat/hid-extended-keys` into `fork/integration` with `git merge --no-ff`, after the operator's word. Push only when asked.
