package mkfs

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"NanoKVM-Server/service/storage/diskfs"
)

// mkfsFATBPB is boot sector bytes 11 to 38, from the BIOS parameter block
// through the extended boot signature, as "mkfs.fat -C -f 2 -F 12" (dosfstools
// 4.2) writes them for each floppy size. Only the 2.88 MB root directory
// differs: DOS and Windows give it 240 entries, mkfs.fat 224.
var mkfsFATBPB = map[int64]string{
	737280:  "0002020100027000a005f90300090002000000000000000000000029",
	1474560: "000201010002e000400bf00900120002000000000000000000000029",
	2949120: "000202010002e0008016f00900240002000000000000000000000029",
}

func TestFloppyBootSectorMatchesMkfsFAT(t *testing.T) {
	for size, want := range mkfsFATBPB {
		disk, err := newFloppy(size, "", 0x12345678, time.Now())
		if err != nil {
			t.Fatalf("%d: %v", size, err)
		}
		if len(disk) != int(size) {
			t.Errorf("%d: %d bytes", size, len(disk))
		}
		ref, _ := hex.DecodeString(want)
		if size == 2949120 {
			binary.LittleEndian.PutUint16(ref[17-11:], 240)
		}
		if got := disk[11:39]; !bytes.Equal(got, ref) {
			t.Errorf("%d: BPB %x, want %x", size, got, ref)
		}
		if string(disk[54:62]) != "FAT12   " || string(disk[43:54]) != "NO NAME    " {
			t.Errorf("%d: type %q, label %q", size, disk[54:62], disk[43:54])
		}
		if disk[510] != 0x55 || disk[511] != 0xAA {
			t.Errorf("%d: no boot signature", size)
		}
		// Both FATs start with the media descriptor; the root directory
		// of an unlabelled disk is empty.
		f := floppyFormats[size]
		for i := 0; i < floppyNumFATs; i++ {
			at := (floppyReserved + i*int(f.fatSectors)) * sectorSize
			if got := disk[at : at+4]; !bytes.Equal(got, []byte{f.media, 0xFF, 0xFF, 0}) {
				t.Errorf("%d: FAT %d starts % x", size, i, got)
			}
		}
		root := (floppyReserved + floppyNumFATs*int(f.fatSectors)) * sectorSize
		if disk[root] != 0 {
			t.Errorf("%d: root directory starts %#x", size, disk[root])
		}
	}
}

// The boot code prints its message from where the BIOS loads the sector.
func TestFloppyBootCodePointsAtItsMessage(t *testing.T) {
	disk, err := newFloppy(1474560, "", 1, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	at := int(binary.LittleEndian.Uint16(disk[62+5:])) - 0x7C00
	if at != 62+len(floppyBootCode) || !bytes.HasPrefix(disk[at:], []byte(floppyBootMessage)) {
		t.Errorf("message at %d, code ends at %d", at, 62+len(floppyBootCode))
	}
}

func TestFloppyLabel(t *testing.T) {
	now := time.Date(2026, 10, 27, 12, 30, 10, 0, time.Local)
	disk, err := newFloppy(737280, "F6 DRIVERS", 7, now)
	if err != nil {
		t.Fatal(err)
	}
	if string(disk[43:54]) != "F6 DRIVERS " {
		t.Errorf("boot sector label %q", disk[43:54])
	}
	entry := disk[(1+2*3)*sectorSize:][:32]
	date, clock := fatTimestamp(now)
	if string(entry[:11]) != "F6 DRIVERS " || entry[11] != attrVolumeID ||
		binary.LittleEndian.Uint16(entry[22:]) != clock || binary.LittleEndian.Uint16(entry[24:]) != date {
		t.Errorf("label entry % x", entry)
	}

	if _, err := newFloppy(737280, "TWELVE CHARS", 7, now); err == nil {
		t.Error("a 12 character label passed")
	}
}

func TestCreateFloppy(t *testing.T) {
	dir := t.TempDir()
	for size := range floppyFormats {
		path := filepath.Join(dir, "f.img")
		if err := CreateFloppy(path, size, "DOS"); err != nil {
			t.Fatalf("%d: %v", size, err)
		}
		info, err := os.Stat(path)
		if err != nil || info.Size() != size {
			t.Fatalf("%d: %v, %v", size, info, err)
		}
		if vol, err := diskfs.Detect(path); err != nil || vol.Kind != diskfs.KindFAT12 || vol.Offset != 0 {
			t.Errorf("%d: detected %+v, %v", size, vol, err)
		}
		fsckFloppy(t, path)

		if err := CreateFloppy(path, size, ""); !errors.Is(err, fs.ErrExist) {
			t.Errorf("%d: an existing image gave %v", size, err)
		}
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}

	for _, size := range []int64{0, 368640, 1474560 + 512, 64 << 20} {
		path := filepath.Join(dir, "bad.img")
		if err := CreateFloppy(path, size, ""); err == nil {
			t.Errorf("a floppy of %d bytes was made", size)
		}
		if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%d: left %s behind", size, path)
		}
	}
}

func TestIsFloppySize(t *testing.T) {
	for size, want := range map[int64]bool{737280: true, 1474560: true, 2949120: true, 1440000: false, 0: false} {
		if IsFloppySize(size) != want {
			t.Errorf("IsFloppySize(%d) = %v", size, !want)
		}
	}
}

// fsckFloppy checks the volume with dosfstools' checker, when the host has
// it. A floppy has no partition table, so it reads the file directly.
func fsckFloppy(t *testing.T, path string) {
	t.Helper()
	tool, err := exec.LookPath("fsck.fat")
	if err != nil {
		t.Log("no fsck.fat here, the volume is not checked by it")
		return
	}
	out, err := exec.Command(tool, "-n", "-v", path).CombinedOutput()
	if err != nil {
		t.Errorf("fsck.fat %s: %v\n%s", path, err, out)
	} else {
		t.Logf("fsck.fat: clean\n%s", out)
	}
}
