package install

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/utils/atomicfile"
)

// Variables rather than constants so the tests can stand in for the tools
// and the kernel.
var (
	// Mksquashfs packs the image. squashfs-tools is in the image's package
	// list; an older image has none, and Install falls back to plain files.
	Mksquashfs = "mksquashfs"
	// Filesystems lists the filesystems the kernel has. Slot B's mainline
	// kernel has no squashfs until its configuration gains CONFIG_SQUASHFS.
	Filesystems = "/proc/filesystems"
)

const (
	currentFile  = "current"
	imageSuffix  = ".sqfs"
	sumSuffix    = ".sha256"
	recordFile   = "installed.json"
	previousFile = "installed.previous.json"
)

// ImageSupported reports whether an add-on can be an image on this slot: the
// kernel mounts squashfs and mksquashfs is there to pack one. squashfs is
// built in where it is configured at all, so /proc/filesystems lists it
// without a module load.
func ImageSupported() bool {
	if _, err := exec.LookPath(Mksquashfs); err != nil {
		return false
	}
	f, err := os.Open(Filesystems)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) > 0 && fields[len(fields)-1] == "squashfs" {
			return true
		}
	}
	return false
}

// Current is the file name of the add-on's image in use, read from its
// current file, and whether there is one.
func Current(name string) (string, bool) {
	data, err := os.ReadFile(filepath.Join(addon.Dir(name), currentFile))
	if err != nil {
		return "", false
	}
	image := strings.TrimSpace(string(data))
	if !plainName(image) || !strings.HasSuffix(image, imageSuffix) {
		return "", false
	}
	return image, true
}

// LinkFile is what a link of the add-on names for file: sqfs/<file> while the
// add-on is an image, and <file> while it is plain files.
func LinkFile(name, file string) string {
	if _, ok := Current(name); ok {
		return addon.ImageDir + "/" + file
	}
	return file
}

// FilesDir is where the add-on's files are: the image's mount point while it
// is an image, and the add-on's directory while it is plain files.
func FilesDir(name string) string {
	if _, ok := Current(name); ok {
		return filepath.Join(addon.Dir(name), addon.ImageDir)
	}
	return addon.Dir(name)
}

// pack makes an image of stage at out. Every file and directory gets its mode
// from the manifest through a pseudo file, because the staged copies are on
// exFAT, which keeps no modes, and owner root. One processor and a small
// cache, because the board has 256 MB and a running KVM.
func pack(stage, out, ws string, members []Member) error {
	modes := map[string]os.FileMode{}
	for _, m := range members {
		if !m.Licences {
			modes[m.Name] = m.mode()
		}
	}
	var pseudo strings.Builder
	err := filepath.WalkDir(stage, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(stage, p)
		if err != nil || rel == "." {
			return err
		}
		rel = filepath.ToSlash(rel)
		mode := os.FileMode(0o644)
		if d.IsDir() {
			mode = 0o755
		} else if m, ok := modes[rel]; ok {
			mode = m
		}
		fmt.Fprintf(&pseudo, "%s m %o 0 0\n", rel, mode)
		return nil
	})
	if err != nil {
		return err
	}
	pseudoFile := filepath.Join(ws, "pseudo")
	if err := os.WriteFile(pseudoFile, []byte(pseudo.String()), 0o644); err != nil {
		return err
	}
	_ = os.Remove(out)
	if _, err := runTool(Mksquashfs, stage, out,
		"-comp", "lz4", "-noappend", "-no-xattrs", "-all-root", "-root-mode", "755",
		"-no-progress", "-quiet", "-processors", "1", "-mem", "32M",
		"-pf", pseudoFile); err != nil {
		_ = os.Remove(out)
		return fmt.Errorf("mksquashfs: %w", err)
	}
	return nil
}

// fileSum is the SHA-256 of a whole file.
func fileSum(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// checkImage compares the whole image with its sidecar.
func checkImage(dir, image string) error {
	data, err := os.ReadFile(filepath.Join(dir, image+sumSuffix))
	if err != nil {
		return fmt.Errorf("%s has no recorded sha256: %w", image, err)
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return fmt.Errorf("%s has no recorded sha256", image)
	}
	got, err := fileSum(filepath.Join(dir, image))
	if err != nil {
		return err
	}
	if got != strings.ToLower(fields[0]) {
		return fmt.Errorf("%s has sha256 %s, and %s%s records %s, so it is not mounted", image, got, image, sumSuffix, fields[0])
	}
	return nil
}

// mount detaches whatever image the add-on has mounted and mounts image,
// after checking it whole against its sidecar. The detach is lazy, so a
// daemon that still runs from the old image keeps it until it exits.
func mount(name, image string) error {
	if err := checkImage(addon.Dir(name), image); err != nil {
		return err
	}
	return mountChecked(name, image)
}

// mountChecked is mount for an image whose sum was taken just now.
func mountChecked(name, image string) error {
	dir := addon.Dir(name)
	mnt := filepath.Join(dir, addon.ImageDir)
	if err := addon.UnmountImage(name); err != nil {
		return err
	}
	if err := os.MkdirAll(mnt, 0o755); err != nil {
		return err
	}
	if out, err := exec.Command(addon.MountCmd, "-t", "squashfs", "-o", "loop,ro", filepath.Join(dir, image), mnt).CombinedOutput(); err != nil {
		return fmt.Errorf("mount %s: %v: %s", image, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Mount mounts the add-on's current image, checked whole first. S04addons
// does the same at boot; this is for the server, after a rollback or when the
// mount went away.
func Mount(name string) error {
	image, ok := Current(name)
	if !ok {
		return fmt.Errorf("%s has no image", name)
	}
	return mount(name, image)
}

// activate makes image, already in the add-on's directory, the one in use:
// it records the image's sum, mounts it, and only then switches current, so a
// failed mount leaves current naming the image that worked. The image it
// replaces is kept for Rollback, and any older one is removed.
func activate(name, image string) error {
	dir := addon.Dir(name)
	sum, err := fileSum(filepath.Join(dir, image))
	if err != nil {
		return err
	}
	if err := atomicfile.WriteFile(filepath.Join(dir, image+sumSuffix), []byte(sum+"  "+image+"\n"), 0o644); err != nil {
		return err
	}
	previous, _ := Current(name)
	// The sum was taken from the file just now, which is the check before
	// this mount.
	if err := mountChecked(name, image); err != nil {
		if previous != "" && previous != image {
			_ = mount(name, previous)
		}
		return err
	}
	if err := atomicfile.WriteFile(filepath.Join(dir, currentFile), []byte(image+"\n"), 0o644); err != nil {
		return err
	}
	prune(dir, image, previous)
	return nil
}

// prune removes every image in dir but the ones named, with their sidecars,
// and any image a failed install left half written.
func prune(dir string, keep ...string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	kept := map[string]bool{}
	for _, k := range keep {
		if k != "" {
			kept[k] = true
		}
	}
	for _, e := range entries {
		n := e.Name()
		base := strings.TrimSuffix(strings.TrimSuffix(n, sumSuffix), ".part")
		if !strings.HasSuffix(base, imageSuffix) || kept[n] {
			continue
		}
		if kept[base] && n == base+sumSuffix {
			continue
		}
		_ = os.Remove(filepath.Join(dir, n))
	}
}

// images lists the images in the add-on's directory that have a sidecar.
func images(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		n := e.Name()
		if strings.HasSuffix(n, imageSuffix) && e.Type().IsRegular() {
			if _, err := os.Stat(filepath.Join(dir, n+sumSuffix)); err == nil {
				out = append(out, n)
			}
		}
	}
	sort.Strings(out)
	return out
}

// imageName is a name for the new image that is not the one in use. A
// reinstall of the same version would otherwise write over the file the loop
// device is reading.
func imageName(dir string, m Manifest) string {
	current, _ := Current(m.Name)
	base := m.Name + "-" + m.Version
	name := base + imageSuffix
	for i := 2; name == current; i++ {
		name = fmt.Sprintf("%s.%d%s", base, i, imageSuffix)
	}
	return name
}

// Rollback switches the add-on back to the image the last update replaced,
// with its installed record. The caller stops the add-on first and starts it
// after.
func Rollback(name string) error {
	dir := addon.Dir(name)
	current, ok := Current(name)
	if !ok {
		return fmt.Errorf("%s is not an image, so there is nothing to roll back to", name)
	}
	var previous string
	for _, img := range images(dir) {
		if img != current {
			previous = img
		}
	}
	if previous == "" {
		return fmt.Errorf("%s has no previous image", name)
	}
	if err := mount(name, previous); err != nil {
		_ = mount(name, current)
		return err
	}
	if err := atomicfile.WriteFile(filepath.Join(dir, currentFile), []byte(previous+"\n"), 0o644); err != nil {
		return err
	}
	rec, prev := filepath.Join(dir, recordFile), filepath.Join(dir, previousFile)
	if _, err := os.Stat(prev); err == nil {
		tmp := rec + ".swap"
		_ = os.Rename(rec, tmp)
		_ = os.Rename(prev, rec)
		_ = os.Rename(tmp, prev)
	}
	return nil
}

// dropImages takes the add-on out of image form after a plain install: the
// mount, current, and every image. A slot without squashfs cannot mount them
// anyway, and S04addons would otherwise keep trying.
func dropImages(name string) error {
	if err := addon.UnmountImage(name); err != nil {
		return err
	}
	dir := addon.Dir(name)
	if err := os.Remove(filepath.Join(dir, currentFile)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	prune(dir)
	_ = os.Remove(filepath.Join(dir, addon.ImageDir))
	return nil
}
