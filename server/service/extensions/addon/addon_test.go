package addon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scratch points every variable at a temporary tree: a root filesystem, a
// /data with a mounts file that says it is mounted, and mount and umount stubs
// that record their arguments. The mount stub also adds the line a real bind
// would, so a second bind of the same directory is seen as already there.
func scratch(t *testing.T) (fsroot string, calls string) {
	t.Helper()
	base := t.TempDir()
	fsroot = filepath.Join(base, "root")
	data := filepath.Join(base, "data")
	for _, d := range []string{fsroot + "/usr/bin", fsroot + "/usr/sbin", fsroot + "/root", data} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	marker := filepath.Join(base, "deviceinfo")
	if err := os.WriteFile(marker, []byte("DEVICE=one-board\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mounts := filepath.Join(base, "mounts")
	if err := os.WriteFile(mounts, []byte("/dev/x "+data+" exfat rw 0 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	calls = filepath.Join(base, "calls")
	mountStub := filepath.Join(base, "mount")
	umountStub := filepath.Join(base, "umount")
	writeStub(t, mountStub, `echo "mount $*" >> "`+calls+`"; echo "x $3 none rw 0 0" >> "`+mounts+`"`)
	writeStub(t, umountStub, `echo "umount $*" >> "`+calls+`"`)

	saved := []string{Root, DataDir, DistroMarker, Mounts, FsRoot, MountCmd, UmountCmd}
	t.Cleanup(func() {
		Root, DataDir, DistroMarker, Mounts, FsRoot, MountCmd, UmountCmd =
			saved[0], saved[1], saved[2], saved[3], saved[4], saved[5], saved[6]
	})
	Root = filepath.Join(data, "ironkvm", "addons")
	DataDir = data
	DistroMarker = marker
	Mounts = mounts
	FsRoot = fsroot
	MountCmd = mountStub
	UmountCmd = umountStub
	return fsroot, calls
}

func writeStub(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

var vpn = Spec{
	Name:  "vpn",
	Links: []Link{{"/usr/bin/tool", "tool"}, {"/usr/sbin/toold", "toold"}},
	Initd: "S98vpn",
}

func TestOnData(t *testing.T) {
	scratch(t)
	if !OnData() {
		t.Fatal("a distribution image with /data mounted must answer true")
	}

	_ = os.WriteFile(Mounts, []byte("/dev/x /other exfat rw 0 0\n"), 0o644)
	if OnData() {
		t.Fatal("with /data not mounted the add-ons have nowhere to go")
	}

	scratch(t)
	_ = os.Remove(DistroMarker)
	if OnData() {
		t.Fatal("off a distribution image the server installs as it always has")
	}
}

// The files are read by S04addons in the image, so their format is a contract
// between two repositories and is held byte for byte.
func TestRecordWritesTheContract(t *testing.T) {
	fsroot, _ := scratch(t)
	dir := Dir("vpn")
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "tool"), []byte("bin"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "toold"), []byte("bin"), 0o755)

	if err := Record(vpn); err != nil {
		t.Fatal(err)
	}
	if got := read(t, filepath.Join(dir, "links")); got != "/usr/bin/tool tool\n/usr/sbin/toold toold\n" {
		t.Fatalf("links is %q", got)
	}
	if got := read(t, filepath.Join(dir, "initd")); got != "S98vpn\n" {
		t.Fatalf("initd is %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "binds")); err == nil {
		t.Fatal("an add-on with no binds must have no binds file")
	}
	for _, l := range vpn.Links {
		target, err := os.Readlink(fsroot + l.Path)
		if err != nil || target != filepath.Join(dir, l.File) {
			t.Fatalf("%s links to %q (%v), want %s", l.Path, target, err, filepath.Join(dir, l.File))
		}
	}
}

// An install from before this package left a real file in /usr/bin. The link
// replaces it, or the old binary would keep running.
func TestRecordReplacesAnOldBinary(t *testing.T) {
	fsroot, _ := scratch(t)
	_ = os.WriteFile(fsroot+"/usr/bin/tool", []byte("old"), 0o755)
	if err := Record(Spec{Name: "vpn", Links: []Link{{"/usr/bin/tool", "tool"}}}); err != nil {
		t.Fatal(err)
	}
	if target, err := os.Readlink(fsroot + "/usr/bin/tool"); err != nil || target != filepath.Join(Dir("vpn"), "tool") {
		t.Fatalf("the old binary was not replaced by the link: %q %v", target, err)
	}
}

func TestRecordRefusesWhatS04addonsRefuses(t *testing.T) {
	scratch(t)
	for _, s := range []Spec{
		{Name: "a/b"},
		{Name: ".."},
		{Name: "vpn", Links: []Link{{"/etc/shadow", "tool"}}},
		{Name: "vpn", Links: []Link{{"/usr/bin/../../etc/passwd", "tool"}}},
		{Name: "vpn", Links: []Link{{"/usr/bin/tool", "../x"}}},
		{Name: "vpn", Binds: []Bind{{"/etc/kvm", "home"}}},
		{Name: "vpn", Binds: []Bind{{"/root/.agent", "a/b"}}},
		{Name: "vpn", Initd: "../../bin/sh"},
		{Name: "vpn", Initd: "S9x"},
	} {
		if err := Record(s); err == nil {
			t.Errorf("Record(%+v) was accepted", s)
		}
	}
}

func TestSetEnabled(t *testing.T) {
	scratch(t)
	marker := filepath.Join(Dir("vpn"), "enabled")
	if err := SetEnabled("vpn", true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("enabled was not written")
	}
	if err := SetEnabled("vpn", false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("enabled was not removed")
	}
	if err := SetEnabled("vpn", false); err != nil {
		t.Fatalf("disabling twice must not be an error: %v", err)
	}
}

// PicoClaw keeps its settings in /root/.picoclaw. An owner who set it up before
// this change has them on the slot, and the bind must not hide them.
func TestBindCopiesWhatTheOwnerHadAndBindsOnce(t *testing.T) {
	fsroot, calls := scratch(t)
	_ = os.MkdirAll(fsroot+"/root/.agent", 0o755)
	_ = os.WriteFile(fsroot+"/root/.agent/config.json", []byte(`{"key":"k"}`), 0o600)
	agent := Spec{Name: "agent", Binds: []Bind{{"/root/.agent", "home"}}}

	if err := Record(agent); err != nil {
		t.Fatal(err)
	}
	if got := read(t, filepath.Join(Dir("agent"), "home", "config.json")); got != `{"key":"k"}` {
		t.Fatalf("the owner's settings were not carried into the add-on: %q", got)
	}
	if got := read(t, filepath.Join(Dir("agent"), "binds")); got != "/root/.agent home\n" {
		t.Fatalf("binds is %q", got)
	}
	if err := Record(agent); err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(read(t, calls), "mount --bind"); n != 1 {
		t.Fatalf("bound %d times, want once", n)
	}
}

// A second install must not overwrite settings already on /data with the ones
// a fresh slot happens to hold.
func TestBindNeverOverwritesSettingsOnData(t *testing.T) {
	fsroot, _ := scratch(t)
	home := filepath.Join(Dir("agent"), "home")
	_ = os.MkdirAll(home, 0o755)
	_ = os.WriteFile(filepath.Join(home, "config.json"), []byte("on data"), 0o600)
	_ = os.MkdirAll(fsroot+"/root/.agent", 0o755)
	_ = os.WriteFile(fsroot+"/root/.agent/config.json", []byte("on the slot"), 0o600)

	if err := Record(Spec{Name: "agent", Binds: []Bind{{"/root/.agent", "home"}}}); err != nil {
		t.Fatal(err)
	}
	if got := read(t, filepath.Join(home, "config.json")); got != "on data" {
		t.Fatalf("settings on /data were overwritten: %q", got)
	}
}

func TestRemoveLeavesNothing(t *testing.T) {
	fsroot, calls := scratch(t)
	both := Spec{Name: "vpn", Links: vpn.Links, Binds: []Bind{{"/root/.vpn", "home"}}, Initd: "S98vpn"}
	if err := Record(both); err != nil {
		t.Fatal(err)
	}
	_ = SetEnabled("vpn", true)
	// A link somebody else owns is not the add-on's to remove.
	_ = os.Remove(fsroot + "/usr/sbin/toold")
	_ = os.Symlink("/elsewhere", fsroot+"/usr/sbin/toold")

	if err := Remove(both); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(fsroot + "/usr/bin/tool"); err == nil {
		t.Fatal("the add-on's link was left behind")
	}
	if target, _ := os.Readlink(fsroot + "/usr/sbin/toold"); target != "/elsewhere" {
		t.Fatal("a link that points elsewhere was removed")
	}
	if _, err := os.Stat(Dir("vpn")); err == nil {
		t.Fatal("the add-on's directory was left behind")
	}
	if !strings.Contains(read(t, calls), "umount "+fsroot+"/root/.vpn") {
		t.Fatal("the bind was not undone")
	}
}
