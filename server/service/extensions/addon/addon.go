// Package addon keeps software the owner installs from the web UI on /data, so
// that it survives the switch to a new image.
//
// On a distribution image the root filesystem is a slot, and a new image
// replaces the whole slot. Tailscale and PicoClaw used to install into
// /usr/bin, record "start at boot" as a script in /etc/init.d, and keep their
// settings under /root, so every update lost all three. Here an add-on lives in
// Root/<name>/ instead, and three small files beside its own say what the root
// filesystem needs for it:
//
//	links    "<path> <file>"          a symlink at <path> to <file> in the directory
//	binds    "<directory> <subdir>"   a bind of <subdir> over <directory>
//	initd    "<script>"               the boot script, while the file enabled exists
//
// The image's S04addons reads them at every boot and puts the links, binds and
// boot script back. This package writes them, and applies the links and binds
// at once, so an add-on works on the running slot without a reboot. The format
// and the rules on paths are the ones S04addons checks: a link directly under
// /usr/bin or /usr/sbin, a bind directly under /root, and plain names.
//
// Off a distribution image, or with /data not mounted, OnData is false and the
// callers install the way they always have.
package addon

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Every path is a variable so the tests can point them at a scratch directory.
var (
	Root         = "/data/ironkvm/addons"
	DataDir      = "/data"
	DistroMarker = "/usr/share/ironkvm/deviceinfo"
	Mounts       = "/proc/mounts"
	// FsRoot prefixes every link and bind path. Empty on the device.
	FsRoot = ""
	// MountCmd is run as `MountCmd --bind <src> <dst>` and `UmountCmd <dst>`.
	MountCmd  = "mount"
	UmountCmd = "umount"
)

// Link is a symlink at Path, in the root filesystem, to File in the add-on's
// directory.
type Link struct{ Path, File string }

// Bind mounts Sub, a directory inside the add-on's directory, over Dir.
type Bind struct{ Dir, Sub string }

// Spec is what an add-on needs from the root filesystem.
type Spec struct {
	Name  string
	Links []Link
	Binds []Bind
	// Initd is the name of the boot script in /kvmapp/system/init.d, or empty.
	Initd string
}

// OnData reports whether add-ons belong on /data: a distribution image, which
// carries the device description, with /data mounted.
func OnData() bool {
	if _, err := os.Stat(DistroMarker); err != nil {
		return false
	}
	return isMounted(DataDir)
}

// Dir is the add-on's directory on /data.
func Dir(name string) string {
	return filepath.Join(Root, name)
}

// Record writes the add-on's links, binds and initd files, and applies the
// links and binds now. A directory a bind will cover that already holds files
// is copied into the add-on first, so nothing the owner had is hidden by it.
func Record(s Spec) error {
	if err := s.validate(); err != nil {
		return err
	}
	dir := Dir(s.Name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	var links, binds strings.Builder
	for _, l := range s.Links {
		fmt.Fprintf(&links, "%s %s\n", l.Path, l.File)
	}
	for _, b := range s.Binds {
		fmt.Fprintf(&binds, "%s %s\n", b.Dir, b.Sub)
	}
	if err := writeOrRemove(filepath.Join(dir, "links"), links.String()); err != nil {
		return err
	}
	if err := writeOrRemove(filepath.Join(dir, "binds"), binds.String()); err != nil {
		return err
	}
	initd := ""
	if s.Initd != "" {
		initd = s.Initd + "\n"
	}
	if err := writeOrRemove(filepath.Join(dir, "initd"), initd); err != nil {
		return err
	}

	for _, l := range s.Links {
		if err := makeLink(filepath.Join(dir, l.File), FsRoot+l.Path); err != nil {
			return err
		}
	}
	for _, b := range s.Binds {
		if err := makeBind(filepath.Join(dir, b.Sub), FsRoot+b.Dir); err != nil {
			return err
		}
	}
	return nil
}

// SetEnabled records whether the add-on's boot script belongs in /etc/init.d.
// S04addons acts on it at the next boot; the caller has already started or
// stopped the add-on on this one.
func SetEnabled(name string, on bool) error {
	marker := filepath.Join(Dir(name), "enabled")
	if on {
		if err := os.MkdirAll(Dir(name), 0o755); err != nil {
			return err
		}
		return os.WriteFile(marker, nil, 0o644)
	}
	if err := os.Remove(marker); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// Remove takes the add-on out: its links, if they still point into its
// directory, its binds, and the directory itself.
func Remove(s Spec) error {
	dir := Dir(s.Name)
	for _, l := range s.Links {
		target, err := os.Readlink(FsRoot + l.Path)
		if err == nil && strings.HasPrefix(target, dir+string(os.PathSeparator)) {
			_ = os.Remove(FsRoot + l.Path)
		}
	}
	for _, b := range s.Binds {
		if isMounted(FsRoot + b.Dir) {
			if out, err := exec.Command(UmountCmd, FsRoot+b.Dir).CombinedOutput(); err != nil {
				return fmt.Errorf("umount %s: %v: %s", b.Dir, err, strings.TrimSpace(string(out)))
			}
		}
	}
	return os.RemoveAll(dir)
}

func (s Spec) validate() error {
	if !plain(s.Name) {
		return fmt.Errorf("add-on name %q is not a plain name", s.Name)
	}
	for _, l := range s.Links {
		if !directlyUnder(l.Path, "/usr/bin/") && !directlyUnder(l.Path, "/usr/sbin/") {
			return fmt.Errorf("link %s is not directly under /usr/bin or /usr/sbin", l.Path)
		}
		if !plain(l.File) {
			return fmt.Errorf("link target %q is not a plain name", l.File)
		}
	}
	for _, b := range s.Binds {
		if !directlyUnder(b.Dir, "/root/") || !plain(b.Sub) {
			return fmt.Errorf("bind of %q over %s is not a plain name directly under /root", b.Sub, b.Dir)
		}
	}
	if s.Initd != "" && !validScript(s.Initd) {
		return fmt.Errorf("boot script %q is not S<nn><name>", s.Initd)
	}
	return nil
}

func plain(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsRune(name, '/')
}

func directlyUnder(path, prefix string) bool {
	return strings.HasPrefix(path, prefix) && plain(strings.TrimPrefix(path, prefix))
}

func validScript(name string) bool {
	if len(name) < 4 || name[0] != 'S' || name[1] < '0' || name[1] > '9' || name[2] < '0' || name[2] > '9' {
		return false
	}
	for _, r := range name[3:] {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-') {
			return false
		}
	}
	return true
}

func writeOrRemove(path, content string) error {
	if content == "" {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// makeLink replaces whatever is at path, a file from an install that predates
// this package included, with a link to target.
func makeLink(target, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if current, err := os.Readlink(path); err == nil && current == target {
		return nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Symlink(target, path)
}

func makeBind(src, dst string) error {
	if err := os.MkdirAll(src, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	if isMounted(dst) {
		return nil
	}
	// What the owner already has there is copied in before the bind covers
	// it. Only into an empty source, so a second install never overwrites
	// settings that are already on /data with older ones from the slot.
	if hasEntries(dst) && !hasEntries(src) {
		if out, err := exec.Command("cp", "-a", dst+"/.", src+"/").CombinedOutput(); err != nil {
			return fmt.Errorf("copy %s into %s: %v: %s", dst, src, err, strings.TrimSpace(string(out)))
		}
	}
	if out, err := exec.Command(MountCmd, "--bind", src, dst).CombinedOutput(); err != nil {
		return fmt.Errorf("bind %s over %s: %v: %s", src, dst, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func hasEntries(dir string) bool {
	entries, err := os.ReadDir(dir)
	return err == nil && len(entries) > 0
}

func isMounted(path string) bool {
	f, err := os.Open(Mounts)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) >= 2 && fields[1] == path {
			return true
		}
	}
	return false
}
