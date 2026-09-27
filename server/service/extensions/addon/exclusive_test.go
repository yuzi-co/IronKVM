//go:build linux

package addon

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// scratchDaemons points the pid files, /proc and both init.d directories at a
// temporary tree, and puts a boot script for each daemon in the package copy.
// onData decides whether it also looks like a distribution image with /data.
func scratchDaemons(t *testing.T, onData bool) {
	t.Helper()
	if onData {
		scratch(t)
	} else {
		saved := DistroMarker
		t.Cleanup(func() { DistroMarker = saved })
		DistroMarker = filepath.Join(t.TempDir(), "absent")
	}
	base := t.TempDir()
	savedTs, savedNb := Tailscale, NetBird
	savedInitd, savedPkg, savedProc := InitdDir, PkgInitdDir, ProcDir
	t.Cleanup(func() {
		Tailscale, NetBird = savedTs, savedNb
		InitdDir, PkgInitdDir, ProcDir = savedInitd, savedPkg, savedProc
	})
	InitdDir = filepath.Join(base, "etc-init.d")
	PkgInitdDir = filepath.Join(base, "kvmapp-init.d")
	ProcDir = filepath.Join(base, "proc")
	for _, d := range []string{InitdDir, PkgInitdDir, ProcDir, filepath.Join(base, "run")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	Tailscale.PidFile = filepath.Join(base, "run", "tailscaled.pid")
	NetBird.PidFile = filepath.Join(base, "run", "netbird.pid")
	for _, s := range []string{Tailscale.Initd, NetBird.Initd} {
		if err := os.WriteFile(filepath.Join(PkgInitdDir, s), []byte("#!/bin/sh\n# "+s+"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

// fakeRunning writes d's pid file and a /proc entry whose command line is argv.
func fakeRunning(t *testing.T, d Daemon, pid int, argv ...string) {
	t.Helper()
	if err := os.WriteFile(d.PidFile, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(ProcDir, strconv.Itoa(pid))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmdline"), []byte(strings.Join(argv, "\x00")+"\x00"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPIDNeedsTheDaemonsCommandLine(t *testing.T) {
	scratchDaemons(t, false)
	if _, ok := PID(NetBird); ok {
		t.Fatal("no pid file is not running")
	}

	// After a power cut the pid file can name a pid something else now holds.
	fakeRunning(t, NetBird, 4242, "/bin/sleep", "60")
	if Running(NetBird) {
		t.Fatal("a pid that is not netbird's is not netbird running")
	}

	fakeRunning(t, NetBird, 4243, "/usr/bin/netbird", "service", "run")
	if pid, ok := PID(NetBird); !ok || pid != 4243 {
		t.Fatalf("got %d %v, want 4243 true", pid, ok)
	}

	if err := os.WriteFile(NetBird.PidFile, []byte("junk\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if Running(NetBird) {
		t.Fatal("a pid file that is not a number is not running")
	}
}

func TestBootOffADistributionImageIsTheScriptInInitd(t *testing.T) {
	scratchDaemons(t, false)
	if BootEnabled(NetBird) {
		t.Fatal("nothing enabled yet")
	}
	if err := SetBoot(NetBird, true); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(filepath.Join(InitdDir, "S98netbird"))
	if err != nil || fi.Mode().Perm()&0o100 == 0 {
		t.Fatalf("the boot script must be in init.d and executable: %v", err)
	}
	if !BootEnabled(NetBird) {
		t.Fatal("start at boot must read back on")
	}
	if err := SetBoot(NetBird, false); err != nil {
		t.Fatal(err)
	}
	if BootEnabled(NetBird) {
		t.Fatal("start at boot must read back off")
	}
	if err := SetBoot(NetBird, false); err != nil {
		t.Fatalf("turning it off twice is fine: %v", err)
	}
}

func TestBootOnADistributionImageIsTheEnabledFile(t *testing.T) {
	scratchDaemons(t, true)
	if err := SetBoot(Tailscale, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(Dir("tailscale"), "enabled")); err != nil {
		t.Fatal("the enabled file must be on /data")
	}
	// A new image has no copy in /etc/init.d until S04addons puts it back.
	_ = os.Remove(filepath.Join(InitdDir, "S98tailscaled"))
	if !BootEnabled(Tailscale) {
		t.Fatal("on /data the enabled file is the record")
	}
	if err := SetBoot(Tailscale, false); err != nil {
		t.Fatal(err)
	}
	if BootEnabled(Tailscale) {
		t.Fatal("start at boot must read back off")
	}
}

func TestBlockedByNamesTheOtherWhenItRuns(t *testing.T) {
	scratchDaemons(t, false)
	fakeRunning(t, Tailscale, 100, "/usr/sbin/tailscaled", "--state=x")
	if got := BlockedBy("netbird"); got != "tailscale" {
		t.Fatalf("got %q", got)
	}
	if got := BlockedBy("tailscale"); got != "" {
		t.Fatalf("a daemon does not block itself: %q", got)
	}
}

func TestBlockedByNamesTheOtherWhenItStartsAtBoot(t *testing.T) {
	scratchDaemons(t, false)
	if err := SetBoot(NetBird, true); err != nil {
		t.Fatal(err)
	}
	if got := BlockedBy("tailscale"); got != "netbird" {
		t.Fatalf("got %q", got)
	}
}

func TestNothingBlocksWhenBothAreIdle(t *testing.T) {
	scratchDaemons(t, false)
	if err := CheckExclusive("netbird"); err != nil {
		t.Fatal(err)
	}
	if err := CheckExclusive("tailscale"); err != nil {
		t.Fatal(err)
	}
}

func TestCheckExclusiveNamesTheOther(t *testing.T) {
	scratchDaemons(t, false)
	fakeRunning(t, Tailscale, 100, "/usr/sbin/tailscaled")
	err := CheckExclusive("netbird")
	var be *BlockedError
	if !errors.As(err, &be) || be.By.Name != "tailscale" {
		t.Fatalf("want a BlockedError by tailscale, got %v", err)
	}
	for _, want := range []string{"Tailscale is running or starts at boot", "before you use NetBird"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("%q is missing from %q", want, err.Error())
		}
	}
}

// Two starts at once, one of each VPN: the check and the start happen under
// one lock, so exactly one of them goes ahead.
func TestExclusiveLetsOnlyOneOfTwoConcurrentStartsThrough(t *testing.T) {
	scratchDaemons(t, false)
	start := func(d Daemon, pid int) func() error {
		return func() error {
			time.Sleep(50 * time.Millisecond)
			fakeRunning(t, d, pid, "/usr/bin/"+d.Process)
			return nil
		}
	}
	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i, d := range []Daemon{Tailscale, NetBird} {
		wg.Add(1)
		go func(i int, d Daemon) {
			defer wg.Done()
			errs[i] = Exclusive(d.Name, start(d, 100+i))
		}(i, d)
	}
	wg.Wait()
	ok, blocked := 0, 0
	for _, err := range errs {
		var be *BlockedError
		switch {
		case err == nil:
			ok++
		case errors.As(err, &be):
			blocked++
		default:
			t.Fatalf("unexpected error %v", err)
		}
	}
	if ok != 1 || blocked != 1 {
		t.Fatalf("want one start and one refusal, got %d and %d", ok, blocked)
	}
}

func TestExclusiveDoesNotRunARefusedAction(t *testing.T) {
	scratchDaemons(t, false)
	fakeRunning(t, Tailscale, 100, "/usr/sbin/tailscaled")
	ran := false
	if err := Exclusive("netbird", func() error { ran = true; return nil }); err == nil || ran {
		t.Fatalf("got %v, ran %v", err, ran)
	}
}

// Off a distribution image the copy in /etc/init.d is the one boot runs, and
// nothing else brings it up to date after the package copy changes.
func TestRefreshInitdUpdatesAnEnabledCopy(t *testing.T) {
	scratchDaemons(t, false)
	if err := SetBoot(NetBird, true); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(NetBird.Script(), []byte("#!/bin/sh\n# new\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := RefreshInitd(NetBird); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(InitdDir, NetBird.Initd)); string(b) != "#!/bin/sh\n# new\n" {
		t.Fatalf("the copy is %q", b)
	}
	if err := RefreshInitd(Tailscale); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(InitdDir, Tailscale.Initd)); err == nil {
		t.Fatal("a daemon that does not start at boot gets no copy")
	}
}
