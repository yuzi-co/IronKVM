# Two Virtual Drives Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** The managed host sees a writable removable disk (`lun.0`) and a CD/DVD-ROM (`lun.1`) at the same time, and inserting or ejecting either one never rebinds the UDC.

**Architecture:** S03usbdev creates `lun.1` as a fixed CD-ROM beside `lun.0`. A new drive layer in `server/service/storage/drives.go` inserts and ejects by writing each LUN's `file` (and the disk's `ro`), and new `/api/storage/drives` routes expose it. The old mount endpoints become thin wrappers. The web image menu shows both drives and inserts `.iso` into CD and `.img` into Disk by default.

**Tech Stack:** POSIX sh (S03usbdev), Go 1.25 with gin (server), React, TypeScript, antd, jotai, i18next (web).

**Spec:** `docs/superpowers/specs/2026-09-27-two-drives-design.md`

## Global Constraints

- Work on a branch `feat/two-drives` cut from `fork/integration`.
- No Claude attribution in commit messages: no `Co-Authored-By` or `Claude-Session` lines.
- Drive ids and types are exactly `disk` (`lun.0`) and `cdrom` (`lun.1`).
- `lun.1` inquiry string is exactly `NanoKVM USB CD/DVD-ROM  0520` (vendor padded to 8, product padded to 16, version).
- Images must pass `isMountableImage`: a `.iso` or `.img` under `/data`, no traversal.
- Insert and eject never close the HID handles, never write the UDC, and never call `hid.NoteUSBGadgetMutated()`.
- The locked-medium error text is exactly `the host holds the medium, eject it on the host first`.
- The kernel has no `forced_eject`. Nothing may write that attribute.
- New UI strings go only in `web/src/i18n/locales/en.ts`; the other locales fall back to English.
- Go tests: `MSYS_NO_PATHCONV=1 docker run --rm -v "$(pwd -W):/repo" -v nanokvm-gomod:/go/pkg/mod -w /repo/server -e CGO_ENABLED=0 golang:1.25 go test -tags novision ./service/storage/ ./router/ ./proto/` from the repo root. Docker Desktop must be running.
- Web checks: `cd web && pnpm format && pnpm lint && pnpm build`.
- Gadget script test: `sh tools/usbdev/test-usb-descriptors.sh` from the repo root.

## Review Focus

- The same image under two spellings (`/data/x.iso` and `/data/./x.iso`) must still count as one file for the "not in both drives" rule. Pinned in Task 2.
- A drive id from the URL that is not `disk` or `cdrom` (`lun.0`, `../x`, empty) must fail with "no such drive" and touch nothing. Pinned in Task 2.
- Inserting into a drive whose current medium is locked must fail with the locked-medium error and must not write the new file. Pinned in Task 2.
- Ejecting an empty drive must succeed without writing anything, so a click on Eject after the host already ejected is harmless. Pinned in Task 2.
- Re-inserting the image a drive already holds (for example to change the disk's read-only flag) must succeed, not trip the other-drive rule. Pinned in Task 2.

---

## File Structure

| File | Responsibility |
|------|----------------|
| `kvmapp/system/init.d/S03usbdev` | Creates `lun.1` as a CD-ROM beside `lun.0` |
| `tools/usbdev/test-usb-descriptors.sh` | Asserts the `lun.1` attributes |
| `server/proto/storage.go` | `DriveInfo`, `GetDrivesRsp`, `InsertDriveReq` |
| `server/service/storage/drives.go` (new) | The drive layer: list, insert, eject, which drive holds a file |
| `server/service/storage/drives_test.go` (new) | Drive layer tests against a fake configfs |
| `server/service/storage/image.go` | Handlers: new drive routes, the old endpoints as wrappers, delete guard. Loses the UDC rebind path |
| `server/service/storage/image_test.go` | Loses the `writeMountTarget` tests; gains wrapper and delete-guard tests |
| `server/router/storage.go` | Registers the drive routes |
| `web/src/api/storage.ts` | `Drive`, `getDrives`, `insertDrive`, `ejectDrive`; drops the old mount calls |
| `web/src/pages/desktop/menu/image/drives.tsx` (new) | The drives section: file, eject, disk read-only switch |
| `web/src/pages/desktop/menu/image/images.tsx` | Insert into a chosen drive, eject from the drive that holds it |
| `web/src/pages/desktop/menu/image/index.tsx` | Owns the drive list; drops the CD-ROM mode switch |
| `web/src/i18n/locales/en.ts` | New strings |

---

### Task 1: Gadget: `lun.1` as a fixed CD-ROM

**Files:**
- Modify: `kvmapp/system/init.d/S03usbdev` (the `usb_kept disk` block, about lines 906-925)
- Test: `tools/usbdev/test-usb-descriptors.sh` (the mass storage section, about lines 347-366)

**Interfaces:**
- Produces: `functions/mass_storage.disk0/lun.1` with `removable=1`, `cdrom=1`, `ro=1`, the CD inquiry string and no `file`. Task 2 finds the CD drive by that directory existing.

- [ ] **Step 1: Write the failing test**

In `tools/usbdev/test-usb-descriptors.sh`, after the line
`   "an empty marker leaves no backing file, so the raw eMMC is never exported"`, add:

```sh
present functions/mass_storage.disk0/lun.1 "the CD drive is a second LUN"
is functions/mass_storage.disk0/lun.1/removable 1 "the CD drive is removable"
is functions/mass_storage.disk0/lun.1/cdrom     1 "the CD drive is a CD-ROM"
is functions/mass_storage.disk0/lun.1/ro        1 "the CD drive is read-only"
is functions/mass_storage.disk0/lun.1/inquiry_string "NanoKVM USB CD/DVD-ROM  0520" "CD inquiry string"
absent functions/mass_storage.disk0/lun.1/file "the CD drive starts empty"
```

After the line `is functions/mass_storage.disk0/lun.0/cdrom 0 "the LUN is not a CD-ROM"`, add:

```sh
absent functions/mass_storage.disk0/lun.1/file "the disk image never goes into the CD drive"
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `sh tools/usbdev/test-usb-descriptors.sh`
Expected: FAIL on "the CD drive is a second LUN".

- [ ] **Step 3: Implement**

In `kvmapp/system/init.d/S03usbdev`, directly after the comment block that ends
`# Legacy BIOS reads no 0x55AA signature and hangs in a HLT loop.` and before the `fi` that closes `if usb_kept disk`, add:

```sh

            # A second LUN, fixed as a CD-ROM, so an install ISO and a disk
            # image can be attached together. Neither LUN changes type after
            # this, so a media change is a write to its file and never needs
            # a UDC rebind. Left without a file, it reports no medium.
            mkdir functions/mass_storage.disk0/lun.1
            echo 1 > functions/mass_storage.disk0/lun.1/removable
            echo 1 > functions/mass_storage.disk0/lun.1/cdrom
            echo 1 > functions/mass_storage.disk0/lun.1/ro
            echo "NanoKVM USB CD/DVD-ROM  0520" > functions/mass_storage.disk0/lun.1/inquiry_string
```

Use the same indentation as the surrounding block (12 spaces).

- [ ] **Step 4: Run the test to verify it passes**

Run: `sh tools/usbdev/test-usb-descriptors.sh`
Expected: PASS, with no failures in any section (the hid-only sections must still pass: `lun.1` lives under the same function, so it goes away with it).

Also run: `sh tools/usbdev/test-hid-only-mode.sh`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add kvmapp/system/init.d/S03usbdev tools/usbdev/test-usb-descriptors.sh
git commit -m "S03usbdev: a second LUN, fixed as a CD-ROM, beside the disk

Issue yuzi-co/ironkvm-dist #8. lun.1 is created once as a removable
read-only CD-ROM with no medium, so an install ISO and a disk image can
be attached together and a media change never needs a UDC rebind."
```

---

### Task 2: Server: the drive layer

**Files:**
- Modify: `server/proto/storage.go`
- Create: `server/service/storage/drives.go`
- Test: `server/service/storage/drives_test.go`

**Interfaces:**
- Consumes: `isMountableImage(path string) bool` and `normalizeMountedImage(content string) string` from `image.go` (unchanged).
- Produces, all in package `storage`:
  - `const DriveDisk = "disk"`, `const DriveCdrom = "cdrom"`
  - `var massStorageDir string` (tests override it)
  - `var writeAttr func(path string, data []byte) error` (tests override it)
  - `var driveDefs []driveDef`, where `driveDef` has fields `id string`, `lun string`
  - `var errNoDrive, errInvalidImage, errInOtherDrive, errMediumLocked error`
  - `func listDrives() ([]proto.DriveInfo, error)`
  - `func insertDrive(id string, file string, ro bool) error`
  - `func ejectDrive(id string) error`
  - `func loadedDrive(file string) (string, error)`: the id of the drive holding `file`, or `""`
- Produces, in package `proto`: `DriveInfo{ID, Type, File string; Ro bool}` (JSON `id`, `type`, `file`, `ro`), `GetDrivesRsp{Drives []DriveInfo}` (JSON `drives`), `InsertDriveReq{File string; Ro bool}` (JSON `file`, `ro`).

- [ ] **Step 1: Add the proto types**

Append to `server/proto/storage.go`:

```go
type DriveInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	File string `json:"file"`
	Ro   bool   `json:"ro"`
}

type GetDrivesRsp struct {
	Drives []DriveInfo `json:"drives"`
}

type InsertDriveReq struct {
	File string `json:"file" validate:"required"`
	Ro   bool   `json:"ro" validate:"omitempty"`
}
```

- [ ] **Step 2: Write the failing tests**

Create `server/service/storage/drives_test.go`:

```go
package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// fakeGadget lays out the named LUNs the way configfs shows them with no
// medium: file holds a newline, ro holds the flag. It points the drive layer
// at the directory for the length of the test.
func fakeGadget(t *testing.T, luns ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, lun := range luns {
		path := filepath.Join(dir, lun)
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatalf("setup: %s", err)
		}
		ro := "0\n"
		if lun == "lun.1" {
			ro = "1\n"
		}
		if err := os.WriteFile(filepath.Join(path, "file"), []byte("\n"), 0o666); err != nil {
			t.Fatalf("setup: %s", err)
		}
		if err := os.WriteFile(filepath.Join(path, "ro"), []byte(ro), 0o666); err != nil {
			t.Fatalf("setup: %s", err)
		}
	}

	oldDir := massStorageDir
	massStorageDir = dir
	t.Cleanup(func() { massStorageDir = oldDir })
	return dir
}

// recordWrites replaces writeAttr with one that logs "lun/attr=value" and
// then writes, unless fail returns an error for that write.
func recordWrites(t *testing.T, fail func(path string, data []byte) error) *[]string {
	t.Helper()
	var writes []string
	oldWrite := writeAttr
	writeAttr = func(path string, data []byte) error {
		rel := filepath.Base(filepath.Dir(path)) + "/" + filepath.Base(path)
		writes = append(writes, rel+"="+strings.TrimSpace(string(data)))
		if fail != nil {
			if err := fail(path, data); err != nil {
				return err
			}
		}
		return os.WriteFile(path, data, 0o666)
	}
	t.Cleanup(func() { writeAttr = oldWrite })
	return &writes
}

func readAttr(t *testing.T, dir string, lun string, attr string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, lun, attr))
	if err != nil {
		t.Fatalf("read %s/%s: %s", lun, attr, err)
	}
	return strings.TrimSpace(string(data))
}

// lockedEject makes every eject of a loaded medium fail the way the kernel
// does when the host has locked it.
func lockedEject(path string, data []byte) error {
	if filepath.Base(path) == "file" && strings.TrimSpace(string(data)) == "" {
		return &os.PathError{Op: "write", Path: path, Err: syscall.EBUSY}
	}
	return nil
}

func TestListDrivesShowsBothLunsEmpty(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")

	drives, err := listDrives()
	if err != nil {
		t.Fatalf("listDrives: %s", err)
	}
	if len(drives) != 2 {
		t.Fatalf("got %d drives, want 2: %+v", len(drives), drives)
	}
	if drives[0].ID != DriveDisk || drives[0].Type != DriveDisk || drives[0].File != "" || drives[0].Ro {
		t.Fatalf("disk is %+v", drives[0])
	}
	if drives[1].ID != DriveCdrom || drives[1].Type != DriveCdrom || drives[1].File != "" || !drives[1].Ro {
		t.Fatalf("cdrom is %+v", drives[1])
	}
}

func TestListDrivesOnAnOldGadgetShowsOnlyTheDisk(t *testing.T) {
	fakeGadget(t, "lun.0")

	drives, err := listDrives()
	if err != nil {
		t.Fatalf("listDrives: %s", err)
	}
	if len(drives) != 1 || drives[0].ID != DriveDisk {
		t.Fatalf("got %+v, want only the disk", drives)
	}
}

func TestListDrivesWithTheDiskFunctionOffIsEmpty(t *testing.T) {
	fakeGadget(t)

	drives, err := listDrives()
	if err != nil {
		t.Fatalf("listDrives: %s", err)
	}
	if len(drives) != 0 {
		t.Fatalf("got %+v, want none", drives)
	}
}

func TestInsertIntoTheCdDriveWritesOnlyItsFile(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	writes := recordWrites(t, nil)

	if err := insertDrive(DriveCdrom, "/data/win11.iso", true); err != nil {
		t.Fatalf("insert: %s", err)
	}

	if got := readAttr(t, dir, "lun.1", "file"); got != "/data/win11.iso" {
		t.Fatalf("lun.1/file is %q", got)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "" {
		t.Fatalf("lun.0/file is %q, want it untouched", got)
	}
	if strings.Join(*writes, " ") != "lun.1/file=/data/win11.iso" {
		t.Fatalf("writes %v, want only lun.1/file", *writes)
	}
}

func TestInsertIntoTheDiskSetsRoBeforeTheFile(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := os.WriteFile(filepath.Join(dir, "lun.0", "file"), []byte("/data/old.img\n"), 0o666); err != nil {
		t.Fatalf("setup: %s", err)
	}
	writes := recordWrites(t, nil)

	if err := insertDrive(DriveDisk, "/data/drivers.img", true); err != nil {
		t.Fatalf("insert: %s", err)
	}

	// The kernel takes a change to ro only while no medium is present.
	want := "lun.0/file= lun.0/ro=1 lun.0/file=/data/drivers.img"
	if strings.Join(*writes, " ") != want {
		t.Fatalf("writes %v, want %s", *writes, want)
	}
	if got := readAttr(t, dir, "lun.0", "ro"); got != "1" {
		t.Fatalf("lun.0/ro is %q", got)
	}
}

func TestInsertWritableDiskClearsRo(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := os.WriteFile(filepath.Join(dir, "lun.0", "ro"), []byte("1\n"), 0o666); err != nil {
		t.Fatalf("setup: %s", err)
	}

	if err := insertDrive(DriveDisk, "/data/scratch.img", false); err != nil {
		t.Fatalf("insert: %s", err)
	}
	if got := readAttr(t, dir, "lun.0", "ro"); got != "0" {
		t.Fatalf("lun.0/ro is %q, want 0", got)
	}
}

func TestInsertNeverWritesRoOnTheCdDrive(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	writes := recordWrites(t, nil)

	if err := insertDrive(DriveCdrom, "/data/a.iso", false); err != nil {
		t.Fatalf("insert: %s", err)
	}
	for _, w := range *writes {
		if strings.HasPrefix(w, "lun.1/ro=") {
			t.Fatalf("wrote %s, the CD drive stays read-only", w)
		}
	}
}

func TestInsertRefusesAnImageInTheOtherDrive(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}

	err := insertDrive(DriveDisk, "/data/a.iso", false)
	if !errors.Is(err, errInOtherDrive) {
		t.Fatalf("got %v, want errInOtherDrive", err)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "" {
		t.Fatalf("lun.0/file is %q, want it untouched", got)
	}
}

func TestInsertRefusesTheSameImageSpelledDifferently(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}

	err := insertDrive(DriveDisk, "/data/./a.iso", false)
	if !errors.Is(err, errInOtherDrive) {
		t.Fatalf("got %v, want errInOtherDrive", err)
	}
}

func TestInsertTheImageTheDriveAlreadyHolds(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveDisk, "/data/a.img", false); err != nil {
		t.Fatalf("setup: %s", err)
	}

	if err := insertDrive(DriveDisk, "/data/a.img", true); err != nil {
		t.Fatalf("re-insert: %s", err)
	}
	if got := readAttr(t, dir, "lun.0", "ro"); got != "1" {
		t.Fatalf("lun.0/ro is %q, want 1", got)
	}
}

func TestInsertRejectsAPathOutsideData(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	writes := recordWrites(t, nil)

	for _, file := range []string{"/etc/shadow", "/data/../etc/x.iso", "/dev/mmcblk0p3", "/data/a.txt"} {
		if err := insertDrive(DriveDisk, file, false); !errors.Is(err, errInvalidImage) {
			t.Fatalf("%s: got %v, want errInvalidImage", file, err)
		}
	}
	if len(*writes) != 0 {
		t.Fatalf("writes %v, want none", *writes)
	}
}

func TestUnknownDriveIdsTouchNothing(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	writes := recordWrites(t, nil)

	for _, id := range []string{"", "lun.0", "../lun.0", "CDROM", "floppy"} {
		if err := insertDrive(id, "/data/a.iso", true); !errors.Is(err, errNoDrive) {
			t.Fatalf("insert %q: got %v, want errNoDrive", id, err)
		}
		if err := ejectDrive(id); !errors.Is(err, errNoDrive) {
			t.Fatalf("eject %q: got %v, want errNoDrive", id, err)
		}
	}
	if len(*writes) != 0 {
		t.Fatalf("writes %v, want none", *writes)
	}
}

func TestInsertIntoAMissingCdDriveFails(t *testing.T) {
	fakeGadget(t, "lun.0")

	if err := insertDrive(DriveCdrom, "/data/a.iso", true); !errors.Is(err, errNoDrive) {
		t.Fatalf("got %v, want errNoDrive", err)
	}
}

func TestEjectEmptiesTheDrive(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}

	if err := ejectDrive(DriveCdrom); err != nil {
		t.Fatalf("eject: %s", err)
	}
	if got := readAttr(t, dir, "lun.1", "file"); got != "" {
		t.Fatalf("lun.1/file is %q, want empty", got)
	}
}

func TestEjectAnEmptyDriveWritesNothing(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	writes := recordWrites(t, lockedEject)

	if err := ejectDrive(DriveDisk); err != nil {
		t.Fatalf("eject: %s", err)
	}
	if len(*writes) != 0 {
		t.Fatalf("writes %v, want none", *writes)
	}
}

func TestEjectALockedMediumReportsIt(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}
	recordWrites(t, lockedEject)

	err := ejectDrive(DriveCdrom)
	if !errors.Is(err, errMediumLocked) {
		t.Fatalf("got %v, want errMediumLocked", err)
	}
	if err.Error() != "the host holds the medium, eject it on the host first" {
		t.Fatalf("message is %q", err.Error())
	}
}

func TestInsertOverALockedMediumKeepsIt(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}
	recordWrites(t, lockedEject)

	if err := insertDrive(DriveCdrom, "/data/b.iso", true); !errors.Is(err, errMediumLocked) {
		t.Fatalf("got %v, want errMediumLocked", err)
	}
	if got := readAttr(t, dir, "lun.1", "file"); got != "/data/a.iso" {
		t.Fatalf("lun.1/file is %q, want the old image kept", got)
	}
}

func TestLoadedDriveFindsTheHolder(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveCdrom, "/data/a.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}

	for file, want := range map[string]string{"/data/a.iso": DriveCdrom, "/data/./a.iso": DriveCdrom, "/data/b.iso": ""} {
		got, err := loadedDrive(file)
		if err != nil {
			t.Fatalf("%s: %s", file, err)
		}
		if got != want {
			t.Fatalf("%s: got %q, want %q", file, got, want)
		}
	}
}
```

- [ ] **Step 3: Run the tests to verify they fail**

Run the Go test command from Global Constraints.
Expected: FAIL to compile with `undefined: massStorageDir` (and the other drive layer names).

- [ ] **Step 4: Implement the drive layer**

Create `server/service/storage/drives.go`:

```go
package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"NanoKVM-Server/proto"
)

// The two virtual drives. Each is one LUN of the gadget's mass storage
// function, and neither ever changes type: S03usbdev creates lun.1 as a
// CD-ROM, and lun.0 stays a disk. So a media change is a write to the LUN's
// file, which the host sees as an insert or an eject, and the gadget is never
// re-enumerated.
const (
	DriveDisk  = "disk"
	DriveCdrom = "cdrom"
)

type driveDef struct {
	id  string
	lun string
}

var driveDefs = []driveDef{
	{id: DriveDisk, lun: "lun.0"},
	{id: DriveCdrom, lun: "lun.1"},
}

// massStorageDir is the gadget's mass storage function. Tests point it at a
// temporary directory laid out the same way.
var massStorageDir = "/sys/kernel/config/usb_gadget/g0/functions/mass_storage.disk0"

// writeAttr writes one configfs attribute. Tests replace it, because a plain
// file cannot return the errors the kernel does.
var writeAttr = func(path string, data []byte) error {
	return os.WriteFile(path, data, 0o666)
}

var (
	errNoDrive      = errors.New("no such drive")
	errInvalidImage = errors.New("not an image under /data")
	errInOtherDrive = errors.New("the image is loaded in the other drive")
	// The kernel refuses to eject a medium the host has locked, as Linux does
	// while a CD is mounted. 5.10 has no forced_eject, so only the host can
	// release it.
	errMediumLocked = errors.New("the host holds the medium, eject it on the host first")
)

func lunPath(d driveDef, attr string) string {
	return filepath.Join(massStorageDir, d.lun, attr)
}

func driveExists(d driveDef) bool {
	info, err := os.Stat(filepath.Join(massStorageDir, d.lun))
	return err == nil && info.IsDir()
}

// findDrive returns the drive with this id, if the gadget has its LUN. A
// gadget built before lun.1 existed has only the disk.
func findDrive(id string) (driveDef, error) {
	for _, d := range driveDefs {
		if d.id == id && driveExists(d) {
			return d, nil
		}
	}
	return driveDef{}, errNoDrive
}

func readDrive(d driveDef) (proto.DriveInfo, error) {
	file, err := os.ReadFile(lunPath(d, "file"))
	if err != nil {
		return proto.DriveInfo{}, err
	}
	ro, err := os.ReadFile(lunPath(d, "ro"))
	if err != nil {
		return proto.DriveInfo{}, err
	}

	return proto.DriveInfo{
		ID:   d.id,
		Type: d.id,
		File: normalizeMountedImage(string(file)),
		Ro:   strings.TrimSpace(string(ro)) == "1",
	}, nil
}

// listDrives returns the drives the gadget has, in LUN order.
func listDrives() ([]proto.DriveInfo, error) {
	drives := []proto.DriveInfo{}
	for _, d := range driveDefs {
		if !driveExists(d) {
			continue
		}
		info, err := readDrive(d)
		if err != nil {
			return nil, err
		}
		drives = append(drives, info)
	}
	return drives, nil
}

// loadedDrive returns the id of the drive holding file, or "" if none does.
func loadedDrive(file string) (string, error) {
	drives, err := listDrives()
	if err != nil {
		return "", err
	}
	clean := filepath.Clean(file)
	for _, d := range drives {
		if d.File != "" && filepath.Clean(d.File) == clean {
			return d.ID, nil
		}
	}
	return "", nil
}

func eject(d driveDef) error {
	err := writeAttr(lunPath(d, "file"), []byte("\n"))
	if errors.Is(err, syscall.EBUSY) {
		return errMediumLocked
	}
	return err
}

// ejectDrive removes the drive's medium. An empty drive is left alone.
func ejectDrive(id string) error {
	d, err := findDrive(id)
	if err != nil {
		return err
	}
	info, err := readDrive(d)
	if err != nil {
		return err
	}
	if info.File == "" {
		return nil
	}
	return eject(d)
}

// insertDrive loads file into the drive, replacing what it holds. ro applies
// to the disk only; the CD drive is always read-only.
func insertDrive(id string, file string, ro bool) error {
	if !isMountableImage(file) {
		return errInvalidImage
	}
	d, err := findDrive(id)
	if err != nil {
		return err
	}

	// A writable disk and a CD on one backing file would corrupt it.
	holder, err := loadedDrive(file)
	if err != nil {
		return err
	}
	if holder != "" && holder != id {
		return errInOtherDrive
	}

	info, err := readDrive(d)
	if err != nil {
		return err
	}
	if info.File != "" {
		if err := eject(d); err != nil {
			return err
		}
	}

	// The kernel takes a change to ro only while no medium is present.
	if d.id == DriveDisk {
		flag := "0"
		if ro {
			flag = "1"
		}
		if err := writeAttr(lunPath(d, "ro"), []byte(flag)); err != nil {
			return err
		}
	}

	return writeAttr(lunPath(d, "file"), []byte(filepath.Clean(file)))
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run the Go test command from Global Constraints.
Expected: PASS for `./service/storage/`, `./router/`, `./proto/`.

- [ ] **Step 6: Commit**

```bash
git add server/proto/storage.go server/service/storage/drives.go server/service/storage/drives_test.go
git commit -m "storage: a drive layer for the disk on lun.0 and the CD on lun.1

Insert and eject write only the LUN's file, and the disk's ro while no
medium is present, so neither re-enumerates the gadget. One image cannot
be in both drives, and a medium the host has locked is reported, since
5.10 has no forced_eject."
```

---

### Task 3: Server: drive routes, the old endpoints as wrappers, the delete guard

**Files:**
- Modify: `server/service/storage/image.go`
- Modify: `server/router/storage.go`
- Test: `server/service/storage/image_test.go`

**Interfaces:**
- Consumes: everything Task 2 produces.
- Produces:
  - Handlers `(*Service).GetDrives`, `(*Service).InsertDrive`, `(*Service).EjectDrive`
  - Routes `GET /api/storage/drives`, `POST /api/storage/drives/:id/insert` (body `{file, ro}`), `POST /api/storage/drives/:id/eject`
  - `var hidOnly func() bool` (tests override it)
  - `var errNoCdDrive error` with text `no CD drive`
  - `func legacyMount(file string, cdrom bool) error`, `func legacyMounted() (string, error)`, `func legacyCdrom() (int64, error)`

- [ ] **Step 1: Write the failing tests**

In `server/service/storage/image_test.go`, delete `TestWriteMountTargetLeavesTheGadgetEmptyWhenUnmounting` and `TestWriteMountTargetMountsARealImage` (their function goes away in Step 3; the drive layer never writes a device path on eject, and Task 2 tests that). Add `"errors"` to the imports, and remove `os` or `path/filepath` if no remaining test in the file uses them (the helpers `fakeGadget` and `readAttr` live in `drives_test.go`, same package). Then append:

```go
func TestLegacyMountWithCdromGoesToTheCdDrive(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")

	if err := legacyMount("/data/a.iso", true); err != nil {
		t.Fatalf("mount: %s", err)
	}
	if got := readAttr(t, dir, "lun.1", "file"); got != "/data/a.iso" {
		t.Fatalf("lun.1/file is %q", got)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "" {
		t.Fatalf("lun.0/file is %q, want it untouched", got)
	}
}

func TestLegacyMountWithoutCdromGoesToAWritableDisk(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")

	if err := legacyMount("/data/a.img", false); err != nil {
		t.Fatalf("mount: %s", err)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "/data/a.img" {
		t.Fatalf("lun.0/file is %q", got)
	}
	if got := readAttr(t, dir, "lun.0", "ro"); got != "0" {
		t.Fatalf("lun.0/ro is %q, want 0", got)
	}
}

func TestLegacyMountOfNothingEjectsBoth(t *testing.T) {
	dir := fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveDisk, "/data/a.img", false); err != nil {
		t.Fatalf("setup: %s", err)
	}
	if err := insertDrive(DriveCdrom, "/data/b.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}

	if err := legacyMount("", false); err != nil {
		t.Fatalf("unmount: %s", err)
	}
	if readAttr(t, dir, "lun.0", "file") != "" || readAttr(t, dir, "lun.1", "file") != "" {
		t.Fatalf("drives not empty")
	}
}

func TestLegacyMountOfNothingOnAnOldGadgetEjectsTheDisk(t *testing.T) {
	dir := fakeGadget(t, "lun.0")
	if err := insertDrive(DriveDisk, "/data/a.img", false); err != nil {
		t.Fatalf("setup: %s", err)
	}

	if err := legacyMount("", false); err != nil {
		t.Fatalf("unmount: %s", err)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "" {
		t.Fatalf("lun.0/file is %q", got)
	}
}

func TestLegacyCdromMountOnAnOldGadgetFails(t *testing.T) {
	dir := fakeGadget(t, "lun.0")

	err := legacyMount("/data/a.iso", true)
	if !errors.Is(err, errNoCdDrive) {
		t.Fatalf("got %v, want errNoCdDrive", err)
	}
	if got := readAttr(t, dir, "lun.0", "file"); got != "" {
		t.Fatalf("lun.0/file is %q, the disk must not take the CD's image", got)
	}
}

func TestLegacyMountedPrefersTheCd(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")
	if err := insertDrive(DriveDisk, "/data/a.img", false); err != nil {
		t.Fatalf("setup: %s", err)
	}

	got, err := legacyMounted()
	if err != nil || got != "/data/a.img" {
		t.Fatalf("got %q, %v, want the disk's image", got, err)
	}

	if err := insertDrive(DriveCdrom, "/data/b.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}
	got, err = legacyMounted()
	if err != nil || got != "/data/b.iso" {
		t.Fatalf("got %q, %v, want the CD's image", got, err)
	}
}

func TestLegacyCdromReportsACdMedium(t *testing.T) {
	fakeGadget(t, "lun.0", "lun.1")

	if got, err := legacyCdrom(); err != nil || got != 0 {
		t.Fatalf("empty: got %d, %v, want 0", got, err)
	}
	if err := insertDrive(DriveCdrom, "/data/b.iso", true); err != nil {
		t.Fatalf("setup: %s", err)
	}
	if got, err := legacyCdrom(); err != nil || got != 1 {
		t.Fatalf("loaded: got %d, %v, want 1", got, err)
	}
}

func TestLegacyCdromOnAnOldGadgetIsZero(t *testing.T) {
	fakeGadget(t, "lun.0")

	if got, err := legacyCdrom(); err != nil || got != 0 {
		t.Fatalf("got %d, %v, want 0", got, err)
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run the Go test command from Global Constraints.
Expected: FAIL to compile with `undefined: legacyMount` (and `legacyMounted`, `legacyCdrom`, `errNoCdDrive`).

- [ ] **Step 3: Implement**

In `server/service/storage/image.go`:

1. Replace the `const` block with:

```go
const (
	imageDirectory = "/data"
	// legacyNoImageDevice is what older builds wrote to mean "no image". It is
	// only ever read now, never written.
	legacyNoImageDevice = "/dev/mmcblk0p3"
)

var errNoCdDrive = errors.New("no CD drive")

// hidOnly reports whether the gadget runs without its mass storage function.
// Tests replace it.
var hidOnly = func() bool {
	mode, err := hid.GetMode()
	return err == nil && mode == hid.ModeHidOnly
}
```

2. Delete the function `writeMountTarget`, and replace `MountImage`, `GetMountedImage` and `GetCdRom` whole with the code below (it also adds the three new handlers and the legacy helpers). The `fmt`, `os/exec`, `strconv` and `time` imports become unused; remove them. Add `"errors"`. `os`, `path/filepath`, `strings`, `gin`, `log`, `proto`, `hid` and `utils` stay in use.

```go
// legacyMount serves the single-drive mount call older clients make. An
// image goes into the CD drive when cdrom is set, otherwise into a writable
// disk. No image ejects both drives.
func legacyMount(file string, cdrom bool) error {
	if file == "" {
		for _, d := range driveDefs {
			if err := ejectDrive(d.id); err != nil && !errors.Is(err, errNoDrive) {
				return err
			}
		}
		return nil
	}

	if !cdrom {
		return insertDrive(DriveDisk, file, false)
	}

	// Switching lun.0 into a CD-ROM was the path this replaces, so a gadget
	// without lun.1 fails rather than falling back to it.
	err := insertDrive(DriveCdrom, file, true)
	if errors.Is(err, errNoDrive) {
		return errNoCdDrive
	}
	return err
}

// legacyMounted returns the CD's image if it has one, else the disk's.
func legacyMounted() (string, error) {
	drives, err := listDrives()
	if err != nil {
		return "", err
	}

	file := ""
	for _, d := range drives {
		if d.File == "" {
			continue
		}
		if d.ID == DriveCdrom {
			return d.File, nil
		}
		file = d.File
	}
	return file, nil
}

// legacyCdrom returns 1 when the CD drive holds a medium.
func legacyCdrom() (int64, error) {
	drives, err := listDrives()
	if err != nil {
		return 0, err
	}
	for _, d := range drives {
		if d.ID == DriveCdrom && d.File != "" {
			return 1, nil
		}
	}
	return 0, nil
}

func (s *Service) MountImage(c *gin.Context) {
	var req proto.MountImageReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if hidOnly() {
		rsp.ErrRsp(c, -2, errNoDrive.Error())
		return
	}

	if err := legacyMount(req.File, req.Cdrom); err != nil {
		log.Errorf("mount image %q (cdrom %t) failed: %s", req.File, req.Cdrom, err)
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	rsp.OkRsp(c)
	log.Debugf("mount image %s success", req.File)
}

func (s *Service) GetMountedImage(c *gin.Context) {
	var rsp proto.Response

	if hidOnly() {
		rsp.OkRspWithData(c, &proto.GetMountedImageRsp{File: ""})
		return
	}

	file, err := legacyMounted()
	if err != nil {
		rsp.ErrRsp(c, -2, "read failed")
		return
	}

	rsp.OkRspWithData(c, &proto.GetMountedImageRsp{File: file})
}

func (s *Service) GetCdRom(c *gin.Context) {
	var rsp proto.Response

	if hidOnly() {
		rsp.OkRspWithData(c, &proto.GetCdRomRsp{Cdrom: 0})
		return
	}

	cdrom, err := legacyCdrom()
	if err != nil {
		rsp.ErrRsp(c, -2, "read failed")
		return
	}

	rsp.OkRspWithData(c, &proto.GetCdRomRsp{Cdrom: cdrom})
}

func (s *Service) GetDrives(c *gin.Context) {
	var rsp proto.Response

	if hidOnly() {
		rsp.OkRspWithData(c, &proto.GetDrivesRsp{Drives: []proto.DriveInfo{}})
		return
	}

	drives, err := listDrives()
	if err != nil {
		log.Errorf("read drives failed: %s", err)
		rsp.ErrRsp(c, -2, "read drives failed")
		return
	}

	rsp.OkRspWithData(c, &proto.GetDrivesRsp{Drives: drives})
}

func (s *Service) InsertDrive(c *gin.Context) {
	var req proto.InsertDriveReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if hidOnly() {
		rsp.ErrRsp(c, -2, errNoDrive.Error())
		return
	}

	id := c.Param("id")
	if err := insertDrive(id, req.File, req.Ro); err != nil {
		log.Errorf("insert %q into %q failed: %s", req.File, id, err)
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	rsp.OkRsp(c)
	log.Debugf("inserted %s into %s", req.File, id)
}

func (s *Service) EjectDrive(c *gin.Context) {
	var rsp proto.Response

	if hidOnly() {
		rsp.ErrRsp(c, -2, errNoDrive.Error())
		return
	}

	id := c.Param("id")
	if err := ejectDrive(id); err != nil {
		log.Errorf("eject %q failed: %s", id, err)
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	rsp.OkRsp(c)
	log.Debugf("ejected %s", id)
}
```

3. In `DeleteImage`, directly after the `isMountableImage` check, add:

```go
	// Removing an image a drive is serving pulls the medium out from under
	// the host.
	holder, err := loadedDrive(req.File)
	if err != nil {
		rsp.ErrRsp(c, -2, "read drives failed")
		return
	}
	if holder != "" {
		rsp.ErrRsp(c, -4, "the image is loaded in the "+holder+" drive")
		return
	}
```

4. In `server/router/storage.go`, after the `GetCdRom` line, add:

```go
	api.GET("/storage/drives", service.GetDrives)                // list the virtual drives
	api.POST("/storage/drives/:id/insert", service.InsertDrive) // insert an image into a drive
	api.POST("/storage/drives/:id/eject", service.EjectDrive)   // eject a drive's image
```

5. Confirm the UDC path is gone: `grep -n "UDC\|NoteUSBGadgetMutated\|CloseNoLock\|inquiry_string\|cdrom\"" server/service/storage/image.go` prints nothing. The `hid` import stays, for `hidOnly`.

- [ ] **Step 4: Run the tests to verify they pass**

Run the Go test command from Global Constraints.
Expected: PASS, including the Task 2 tests.

Also run `go vet` the same way (replace `go test` with `go vet` in the command).
Expected: no output.

- [ ] **Step 5: Commit**

```bash
git add server/service/storage/image.go server/service/storage/image_test.go server/router/storage.go
git commit -m "storage: routes for the two drives, the old mount calls on top of them

GET /api/storage/drives lists them, and insert and eject act on one.
The old mount, mounted and cdrom endpoints map onto the drives, so a
mount no longer closes HID or rebinds the UDC. Delete refuses an image
a drive is serving."
```

---

### Task 4: Web UI: two drives in the image menu

**Files:**
- Modify: `web/src/api/storage.ts`
- Create: `web/src/pages/desktop/menu/image/drives.tsx`
- Modify: `web/src/pages/desktop/menu/image/images.tsx`
- Modify: `web/src/pages/desktop/menu/image/index.tsx`
- Modify: `web/src/i18n/locales/en.ts`

**Interfaces:**
- Consumes: the Task 3 routes. `GET /api/storage/drives` returns `{code: 0, data: {drives: [{id, type, file, ro}]}}`. Insert and eject return `{code: 0}` or `{code: -2, msg: <error text>}`.
- Produces: `Drive`, `DriveId`, `getDrives()`, `insertDrive(id, file, ro)`, `ejectDrive(id)` in `@/api/storage.ts`; `<Drives>` and the new `<Images>` props.

There is no web test runner in this repo. The checks are `pnpm format`, `pnpm lint` and `pnpm build`, and the behaviour is checked on the board in Task 5.

- [ ] **Step 1: The API client**

Replace `web/src/api/storage.ts` with:

```ts
import { http } from '@/lib/http.ts';

export type DriveId = 'disk' | 'cdrom';

export type Drive = {
  id: DriveId;
  type: DriveId;
  file: string;
  ro: boolean;
};

// get image list
export function getImages() {
  return http.get('/api/storage/image');
}

// list the virtual drives: the disk, and the CD when the gadget has it
export function getDrives() {
  return http.get('/api/storage/drives');
}

// insert an image into a drive; ro applies to the disk only
export function insertDrive(id: DriveId, file: string, ro: boolean) {
  return http.post(`/api/storage/drives/${id}/insert`, { file, ro });
}

// eject a drive's image
export function ejectDrive(id: DriveId) {
  return http.post(`/api/storage/drives/${id}/eject`);
}

export function deleteImage(file: string) {
  const data = {
    file
  };
  return http.post('/api/storage/image/delete', data);
}
```

- [ ] **Step 2: The strings**

In `web/src/i18n/locales/en.ts`, inside `image: {`, after `refresh: 'Refresh the image list',`, add:

```ts
      disk: 'Disk',
      cdrom: 'CD',
      driveEmpty: 'Empty',
      eject: 'Eject',
      readOnly: 'Read-only',
      readOnlyTip: 'Applies to the next image inserted into the disk.',
      noDrives: 'No virtual drives. Turn on the virtual disk in Settings.',
      insertFailed: 'Insert failed',
      ejectFailed: 'Eject failed',
      insertInto: 'Insert into {{drive}}. Click to change.',
      loadedIn: 'In the {{drive}} drive'
```

Keep the existing keys (`mountMode`, `mountFailed` and the rest); the other locales still carry them.

- [ ] **Step 3: The drives section**

Create `web/src/pages/desktop/menu/image/drives.tsx`:

```tsx
import { useState } from 'react';
import { Button, notification, Switch, Tooltip } from 'antd';
import clsx from 'clsx';
import { DiscIcon, HardDriveIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/storage.ts';

type DrivesProps = {
  drives: api.Drive[];
  diskRo: boolean;
  setDiskRo: (ro: boolean) => void;
  onDrivesChanged: () => void;
};

// Drives shows what each virtual drive holds, with an eject button per drive
// and the read-only switch that the next disk insert uses.
export const Drives = ({ drives, diskRo, setDiskRo, onDrivesChanged }: DrivesProps) => {
  const { t } = useTranslation();
  const [notify, contextHolder] = notification.useNotification();
  const [ejecting, setEjecting] = useState('');

  function eject(id: api.DriveId) {
    if (ejecting) return;
    setEjecting(id);

    api
      .ejectDrive(id)
      .then((rsp) => {
        if (rsp.code !== 0) {
          notify.open({ message: t('image.ejectFailed'), description: rsp.msg, duration: 10 });
        }
      })
      .finally(() => {
        setEjecting('');
        onDrivesChanged();
      });
  }

  if (drives.length === 0) {
    return <div className="text-sm text-neutral-500">{t('image.noDrives')}</div>;
  }

  return (
    <>
      <div className="flex flex-col space-y-3">
        {drives.map((drive) => {
          const Icon = drive.id === 'cdrom' ? DiscIcon : HardDriveIcon;

          return (
            <div key={drive.id} className="flex items-center space-x-2">
              <Icon size={16} />
              <span className="w-[48px]">{t(`image.${drive.id}`)}</span>
              <span
                className={clsx(
                  'flex-1 truncate',
                  drive.file ? 'text-blue-500' : 'text-neutral-500'
                )}
              >
                {drive.file ? drive.file.replace(/^.*[\\/]/, '') : t('image.driveEmpty')}
              </span>

              {drive.id === 'disk' && (
                <Tooltip title={t('image.readOnlyTip')}>
                  <div className="flex items-center space-x-1">
                    <span className="text-xs text-neutral-400">{t('image.readOnly')}</span>
                    <Switch size="small" checked={diskRo} onChange={setDiskRo} />
                  </div>
                </Tooltip>
              )}

              <Button
                size="small"
                disabled={!drive.file}
                loading={ejecting === drive.id}
                onClick={() => eject(drive.id)}
              >
                {t('image.eject')}
              </Button>
            </div>
          );
        })}
      </div>

      {contextHolder}
    </>
  );
};
```

- [ ] **Step 4: The image list**

In `web/src/pages/desktop/menu/image/images.tsx`:

1. Imports: replace the `lucide-react` import with

```tsx
import {
  ArrowBigDownDashIcon,
  ArrowBigUpDashIcon,
  DiscIcon,
  HardDriveIcon,
  LoaderCircleIcon,
  PackageIcon,
  PackageSearchIcon,
  Trash2Icon
} from 'lucide-react';
```

change the first line to `import { useEffect, useState, type MouseEvent as ReactMouseEvent } from 'react';`, change the `antd` import to `import { Button, Modal, notification, Tooltip, Typography } from 'antd';`, and delete `import { client } from '@/lib/websocket.ts';`. No insert or eject re-enumerates the gadget any more, so the input socket stays up.

2. Replace `type ImagesProps` and everything from `export const Images` down to (not including) the `// show delete image modal` comment with:

```tsx
type ImagesProps = {
  isOpen: boolean;
  drives: api.Drive[];
  diskRo: boolean;
  onDrivesChanged: () => void;
};

// defaultDrive picks where a click inserts an image: an ISO into the CD drive
// and anything else into the disk, falling back to whichever drive exists.
function defaultDrive(image: string, available: api.DriveId[]): api.DriveId {
  const preferred: api.DriveId = image.toLowerCase().endsWith('.iso') ? 'cdrom' : 'disk';
  return available.includes(preferred) ? preferred : available[0];
}

export const Images = ({ isOpen, drives, diskRo, onDrivesChanged }: ImagesProps) => {
  const { t } = useTranslation();
  const [notify, contextHolder] = notification.useNotification();

  const [isLoading, setIsLoading] = useState(false);
  const [images, setImages] = useState<string[]>([]);
  const [busyImage, setBusyImage] = useState('');
  const [targets, setTargets] = useState<Record<string, api.DriveId>>({});
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedImage, setSelectedImage] = useState('');
  const [deletingImage, setDeletingImage] = useState('');

  const available = drives.map((drive) => drive.id);

  function loadedIn(image: string): api.DriveId | undefined {
    return drives.find((drive) => drive.file === image)?.id;
  }

  function targetOf(image: string): api.DriveId {
    const chosen = targets[image];
    return chosen && available.includes(chosen) ? chosen : defaultDrive(image, available);
  }

  // get image list
  //
  // force skips the in-flight guard. The effect below passes it: an update event
  // can arrive while an earlier request is still out, and that request may have
  // been answered before the change it announces.
  const getImages = useStableCallback((force = false) => {
    if (isLoading && !force) return;
    setIsLoading(true);

    api
      .getImages()
      .then((rsp) => {
        if (rsp.code !== 0) {
          return;
        }

        const files = rsp.data?.files;
        setImages(files?.length > 0 ? files : []);
        onDrivesChanged();
      })
      .finally(() => {
        setIsLoading(false);
      });
  });

  useEffect(() => {
    if (!isOpen) return;

    getImages(true);

    const handleImageUpdated = () => {
      getImages(true);
    };
    window.addEventListener(imageUpdatedEvent, handleImageUpdated);

    return () => {
      window.removeEventListener(imageUpdatedEvent, handleImageUpdated);
    };
  }, [isOpen, getImages]);

  // flip the drive a click will insert into, when both drives exist
  function toggleTarget(e: ReactMouseEvent, image: string) {
    e.stopPropagation();
    if (available.length < 2) return;

    const next: api.DriveId = targetOf(image) === 'cdrom' ? 'disk' : 'cdrom';
    setTargets((prev) => ({ ...prev, [image]: next }));
  }

  // eject the image from the drive holding it, or insert it into its target
  function insertOrEject(image: string) {
    if (busyImage || available.length === 0) return;
    setBusyImage(image);

    const loaded = loadedIn(image);
    const target = targetOf(image);
    const request = loaded
      ? api.ejectDrive(loaded)
      : api.insertDrive(target, image, target === 'disk' ? diskRo : true);

    request
      .then((rsp) => {
        if (rsp.code !== 0) {
          openNotification(!!loaded, rsp.msg);
        }
      })
      .finally(() => {
        setBusyImage('');
        onDrivesChanged();
      });
  }
```

3. In `showDeleteModal`, replace `const isMounted = mountedImage === image;` with `const isMounted = !!loadedIn(image);`.

4. Replace `openNotification` with:

```tsx
  // show insert/eject failed notification
  function openNotification(isEject: boolean, description: string) {
    notify.open({
      message: t(isEject ? 'image.ejectFailed' : 'image.insertFailed'),
      description,
      duration: 10
    });
  }
```

5. Replace the `{images.map((image) => ( ... ))}` block inside the list `div` with:

```tsx
        {images.map((image) => {
          const loaded = loadedIn(image);
          const drive = loaded ?? targetOf(image);
          const DriveIcon = drive === 'cdrom' ? DiscIcon : HardDriveIcon;
          const driveName = t(`image.${drive}`);

          return (
            <div
              key={image}
              className={clsx(
                'group flex cursor-pointer select-none items-center space-x-1 rounded px-1 py-2 hover:bg-neutral-700/70',
                loaded && 'text-blue-500'
              )}
              onClick={() => insertOrEject(image)}
            >
              <div className="flex h-[24px] w-[24px] items-center justify-center">
                {busyImage === image ? (
                  <LoaderCircleIcon className="animate-spin" size={18} />
                ) : (
                  <PackageIcon size={18} />
                )}
              </div>

              <div className="flex-1 truncate">{image.replace(/^.*[\\/]/, '')}</div>

              {available.length > 0 && (
                <Tooltip
                  title={
                    loaded
                      ? t('image.loadedIn', { drive: driveName })
                      : t('image.insertInto', { drive: driveName })
                  }
                  mouseEnterDelay={0.6}
                >
                  <div
                    className={clsx(
                      'flex h-[24px] w-[24px] items-center justify-center rounded',
                      !loaded && available.length > 1 && 'hover:bg-neutral-500/50'
                    )}
                    onClick={(e) => (loaded ? e.stopPropagation() : toggleTarget(e, image))}
                  >
                    <DriveIcon size={16} />
                  </div>
                </Tooltip>
              )}

              <div className="flex h-[24px] w-[24px] items-center justify-center rounded">
                {loaded ? (
                  <ArrowBigDownDashIcon size={22} className="hidden text-red-500 group-hover:block" />
                ) : (
                  <ArrowBigUpDashIcon size={22} className="hidden text-blue-500 group-hover:block" />
                )}
              </div>

              <div
                className={clsx(
                  'flex h-[24px] w-[24px] items-center justify-center rounded hover:bg-neutral-500/50',
                  loaded ? 'cursor-not-allowed text-neutral-500' : 'text-neutral-300 hover:text-red-500'
                )}
                onClick={(e) => showDeleteModal(e, image)}
              >
                {deletingImage === image ? (
                  <LoaderCircleIcon className="animate-spin text-red-500" size={16} />
                ) : (
                  <Trash2Icon size={16} />
                )}
              </div>
            </div>
          );
        })}
```

6. Confirm nothing refers to `mountedImage`, `mountingImage`, `setIsMounted`, `cdrom` or `client` any more: `grep -n "mountedImage\|mountingImage\|setIsMounted\|cdrom\b\|client\." web/src/pages/desktop/menu/image/images.tsx` prints only the `'cdrom'` literals.

- [ ] **Step 5: The menu**

Replace `web/src/pages/desktop/menu/image/index.tsx` with:

```tsx
import { useEffect, useState } from 'react';
import { Divider, Modal, Tooltip } from 'antd';
import clsx from 'clsx';
import { useSetAtom } from 'jotai';
import { DiscIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/storage.ts';
import { submenuOpenCountAtom } from '@/jotai/settings.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { Drives } from './drives.tsx';
import { Images } from './images.tsx';
import { Tips } from './tips.tsx';

export const Image = () => {
  const { t } = useTranslation();
  const setSubmenuOpenCount = useSetAtom(submenuOpenCountAtom);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [drives, setDrives] = useState<api.Drive[]>([]);
  const [diskRo, setDiskRo] = useState(false);

  const isMounted = drives.some((drive) => !!drive.file);

  const refreshDrives = useStableCallback(() => {
    api.getDrives().then((rsp) => {
      if (rsp.code !== 0) return;

      const list: api.Drive[] = rsp.data?.drives ?? [];
      setDrives(list);

      // A loaded disk shows its real flag. An empty one keeps the operator's
      // choice for the next insert.
      const disk = list.find((drive) => drive.id === 'disk');
      if (disk?.file) {
        setDiskRo(disk.ro);
      }
    });
  });

  useEffect(() => {
    refreshDrives();
  }, [refreshDrives]);

  function toggleModal(open: boolean) {
    setIsModalOpen(open);
    setSubmenuOpenCount((count) => (open ? count + 1 : Math.max(0, count - 1)));
  }

  return (
    <>
      <Tooltip title={t('image.title')} placement="bottom" mouseEnterDelay={0.6}>
        <div
          className={clsx(
            'flex h-[30px] w-[30px] cursor-pointer items-center justify-center rounded hover:bg-neutral-700',
            isMounted ? 'text-blue-500' : 'text-neutral-300 hover:text-white'
          )}
          onClick={() => toggleModal(true)}
        >
          <DiscIcon size={18} />
        </div>
      </Tooltip>

      <Modal open={isModalOpen} footer={null} onCancel={() => toggleModal(false)}>
        <div className="flex items-center space-x-1">
          <span className="text-xl font-bold">{t('image.title')}</span>
          <Tips />
        </div>

        <Divider style={{ margin: '24px 0' }} />

        <div className="flex flex-col space-y-6">
          <Drives
            drives={drives}
            diskRo={diskRo}
            setDiskRo={setDiskRo}
            onDrivesChanged={refreshDrives}
          />

          <Divider style={{ margin: '24px 0 0 0' }} />

          <Images
            isOpen={isModalOpen}
            drives={drives}
            diskRo={diskRo}
            onDrivesChanged={refreshDrives}
          />
        </div>
      </Modal>
    </>
  );
};
```

- [ ] **Step 6: Check**

Run: `cd web && pnpm format && pnpm lint && pnpm build`
Expected: no lint errors, and the build succeeds. If `pnpm format` changed files, include them in the commit.

Also run `grep -rn "getMountedImage\|mountImage\|getCdRom" web/src` and expect no output.

- [ ] **Step 7: Commit**

```bash
git add web/src/api/storage.ts web/src/i18n/locales/en.ts web/src/pages/desktop/menu/image/
git commit -m "web: the disk and the CD drive side by side in the image menu

Each drive shows its image and an eject button, and the disk has a
read-only switch. A click inserts an ISO into the CD and anything else
into the disk, and a per-row icon changes that. The CD-ROM mode switch
is gone, and the input socket is no longer closed around a mount."
```

---

### Task 5: Deploy and verify on the board (controller only, operator-gated)

This task is not dispatched to a subagent. It needs the operator's word before each board step, and SSH on the board must be on.

- [ ] **Step 1: Build**

Build the server (`make app` with `DOCKER_TTY=`, run as its `docker run` command because `make` is not installed on this workstation, then the `patchelf` RPATH step from AGENTS.md) and the web UI (`pnpm build` in `web/`).

- [ ] **Step 2: Ask the operator**, then deploy the server first

1. Deploy the server and web UI with `tools/deploy/deploy-server`. Stage on `/data`, `DEPLOY_TIMEOUT=240`, run detached, and verify from `/proc/PID/exe` (`tools/deploy/check-deployed`).
2. With the old gadget still running, `GET /api/storage/drives` lists only `disk`, and the image menu shows one drive. This is the "new server, old gadget" case from the spec.

- [ ] **Step 3: Ask the operator**, then deploy the gadget script

1. `scp` `kvmapp/system/init.d/S03usbdev` to both `/etc/init.d/S03usbdev` and `/kvmapp/system/init.d/S03usbdev` on root@10.0.0.222. Upload each beside the target and `mv` it into place.
2. Run `/etc/init.d/S03usbdev stop_start`. This re-enumerates the gadget once. Read the `usb watchdog` lines before touching anything else.
3. `ls /sys/kernel/config/usb_gadget/g0/functions/mass_storage.disk0/` shows `lun.0` and `lun.1`; `lun.1/cdrom`, `lun.1/ro` and `lun.1/removable` read `1`.

- [ ] **Step 4: Verify**

On unraid (`ssh root@unraid`):

1. Record the baseline: `lsusb | grep -i sipeed` (the device number) and the `H: Handlers=` lines for the NanoKVM in `/proc/bus/input/devices`.
2. Put a 1 MiB test ISO on the board, for example `genisoimage` output or a truncated copy, at `/data/drive-test.iso`, and note its `md5sum`. Put a 1 MiB image at `/data/drive-test.img`.
3. From the UI, insert `drive-test.iso` (it goes to CD by default) and `drive-test.img` (to Disk). On unraid, `lsblk` shows both an `sdX` and an `srX`.
4. `dd if=/dev/srX bs=64k | md5sum` equals the ISO's `md5sum`.
5. Eject each from the drives section. After every insert and eject, `lsusb` shows the same device number and the input handlers are unchanged.
6. `mount /dev/srX /mnt` on unraid, then Eject CD in the UI: the UI shows "the host holds the medium, eject it on the host first". `umount /mnt`, then Eject succeeds.
7. The operator confirms the keyboard and mouse worked throughout.
8. Delete refuses an image while it is loaded.

Not tested: a Windows 11 install from the CD drive, in line with #9.

- [ ] **Step 5: Record**

Update the memory note `media-swap-without-udc-reset.md`: the server no longer rebinds the UDC for a CD mount, and `lun.1` is the CD drive. Then merge `feat/two-drives` into `fork/integration` with `git merge --no-ff`, after the operator's word. Push only when asked. Close yuzi-co/ironkvm-dist #8 with the results.
