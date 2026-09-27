package addon

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Paths the boot and exclusivity checks read. Variables so the tests can point
// them at a scratch tree.
var (
	InitdDir    = "/etc/init.d"
	PkgInitdDir = "/kvmapp/system/init.d"
	ProcDir     = "/proc"
)

// Daemon is an add-on that runs as a daemon in the addons memory group.
type Daemon struct {
	Name    string // the add-on's name, as in Spec
	Title   string // how the page names it
	Initd   string // its boot script
	PidFile string // written by the boot script
	Process string // an argument its /proc/<pid>/cmdline holds, by base name
}

// Tailscale and NetBird each hold the addons group near its memory.high on
// their own, idle; together they hold it at the limit (measured on image w,
// 2026-09-27). So only one of them may run or start at boot.
var (
	Tailscale = Daemon{
		Name: "tailscale", Title: "Tailscale", Initd: "S98tailscaled",
		PidFile: "/var/run/tailscaled.pid", Process: "tailscaled",
	}
	NetBird = Daemon{
		Name: "netbird", Title: "NetBird", Initd: "S98netbird",
		PidFile: "/var/run/netbird.pid", Process: "netbird",
	}
)

// exclusive is read at every call, so a test that moves a pid file is seen.
func exclusive() []Daemon { return []Daemon{Tailscale, NetBird} }

// Script is the boot script in the package copy, which start, stop and restart
// run. /etc/init.d holds a copy only while the daemon starts at boot.
func (d Daemon) Script() string { return filepath.Join(PkgInitdDir, d.Initd) }

// PID is the daemon's process id, when its pid file names a live process
// whose command line holds d.Process. A stale pid file after a power cut can
// name a pid that something else now holds, and that is not the daemon.
func PID(d Daemon) (int, bool) {
	b, err := os.ReadFile(d.PidFile)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	cmdline, err := os.ReadFile(filepath.Join(ProcDir, strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return 0, false
	}
	for _, arg := range bytes.Split(cmdline, []byte{0}) {
		if len(arg) > 0 && filepath.Base(string(arg)) == d.Process {
			return pid, true
		}
	}
	return 0, false
}

// Running reports whether the daemon runs now.
func Running(d Daemon) bool {
	_, ok := PID(d)
	return ok
}

// BootEnabled reports whether the daemon starts at boot. On a distribution
// image that is the add-on's enabled file on /data, which S04addons reads;
// anywhere else it is the script in /etc/init.d.
func BootEnabled(d Daemon) bool {
	if OnData() {
		return exists(filepath.Join(Dir(d.Name), "enabled"))
	}
	return exists(filepath.Join(InitdDir, d.Initd))
}

// SetBoot turns start at boot on or off. It puts the boot script into
// /etc/init.d or takes it out, which is the whole record off a distribution
// image; on one it also writes the enabled file, so the next image keeps it.
func SetBoot(d Daemon, on bool) error {
	dst := filepath.Join(InitdDir, d.Initd)
	if on {
		if err := copyFile(d.Script(), dst, 0o755); err != nil {
			return err
		}
	} else if err := os.Remove(dst); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if OnData() {
		return SetEnabled(d.Name, on)
	}
	return nil
}

// BlockedBy names the other daemon when it runs or starts at boot, and is
// empty when self may go ahead.
func BlockedBy(self string) string {
	for _, d := range exclusive() {
		if d.Name != self && (Running(d) || BootEnabled(d)) {
			return d.Name
		}
	}
	return ""
}

// BlockedError refuses an action because the other VPN runs or starts at
// boot.
type BlockedError struct {
	Self, By Daemon
}

func (e *BlockedError) Error() string {
	return fmt.Sprintf("%s is running or starts at boot. Only one VPN runs at a time: "+
		"stop %s and turn off its start at boot before you use %s.",
		e.By.Title, e.By.Title, e.Self.Title)
}

// CheckExclusive is the one check both add-ons call before install, start,
// up, login and turning start at boot on.
func CheckExclusive(self string) error {
	by := BlockedBy(self)
	if by == "" {
		return nil
	}
	return &BlockedError{Self: byName(self), By: byName(by)}
}

func byName(name string) Daemon {
	for _, d := range exclusive() {
		if d.Name == name {
			return d
		}
	}
	return Daemon{Name: name, Title: name}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// copyFile writes a copy through a temporary file and a rename, so a boot
// that reads /etc/init.d never sees half a script.
func copyFile(src, dst string, mode os.FileMode) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, b, mode); err != nil {
		return err
	}
	// WriteFile honours the umask.
	if err := os.Chmod(tmp, mode); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

// vpnMu serialises every action that could leave both VPNs running or
// enabled at boot. CheckExclusive alone is a check followed by an action, and
// two requests, one per VPN, could both pass the check before either acted.
var vpnMu sync.Mutex

// LockVPN takes the lock that the exclusivity check and the action it guards
// hold together, and returns the function that releases it. Long work that
// cannot start a daemon (a download, apk) must not run under it.
func LockVPN() func() {
	vpnMu.Lock()
	return vpnMu.Unlock
}

// Exclusive runs action under the VPN lock, after CheckExclusive, and returns
// the refusal instead when the other VPN runs or starts at boot.
func Exclusive(self string, action func() error) error {
	defer LockVPN()()
	if err := CheckExclusive(self); err != nil {
		return err
	}
	return action()
}

// RefreshInitd brings the copy in /etc/init.d up to date with the package
// copy while the daemon starts at boot. Off a distribution image nothing else
// refreshes it: S04addons, which puts it back at boot, only runs on one.
func RefreshInitd(d Daemon) error {
	if !BootEnabled(d) {
		return nil
	}
	return copyFile(d.Script(), filepath.Join(InitdDir, d.Initd), 0o755)
}
