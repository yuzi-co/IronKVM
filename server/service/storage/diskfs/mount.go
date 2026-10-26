// Parts of this file are ported from sipeed/NanoKVM
// (server/service/storage/diskfs/mount.go, commit 5572e1de, PR #957).
// Copyright (c) Sipeed. Licensed under the GNU General Public License v3.0.

package diskfs

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

// MountRoot holds the mounts while an image is browsed. It is under /run: a
// reboot clears it, and it is not inside /data, where the image list walks
// every file. Tests point it at a temporary directory.
var MountRoot = "/run/ironkvm/media"

type mount struct {
	path       string
	mountpoint string
	tree       *tree
	writable   bool
	refs       int
}

var (
	mountsMu sync.Mutex
	mounts   = map[string]*mount{}
)

// keyOf is the name an image is tracked under: the file itself, with every
// link followed, so two spellings of one image share one mount.
func keyOf(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

// Mount opens the filesystem inside an image and returns a handle to it.
// Handles are reference counted, because a download may stream for minutes
// while the listing is refreshed, and both have to see the same mount. A
// reader shares a writable mount; a writer is refused while a read-only mount
// is open, since its readers were promised a filesystem that does not change.
//
// The caller must make sure the host cannot see the image while it is open for
// writing, and cannot write it while it is open at all.
func Mount(imagePath string, writable bool) (Image, error) {
	vol, err := Detect(imagePath)
	if err != nil {
		return nil, err
	}
	if !vol.Kind.Browsable() {
		return nil, fmt.Errorf("%w (%s)", ErrUnsupported, vol.Kind)
	}
	if writable && !vol.Kind.Writable() {
		return nil, fmt.Errorf("%w: %s cannot be written", ErrReadOnly, vol.Kind)
	}

	key := keyOf(imagePath)

	mountsMu.Lock()
	defer mountsMu.Unlock()

	if existing, ok := mounts[key]; ok {
		if writable && !existing.writable {
			return nil, fmt.Errorf("%w: it is open for reading", ErrBusy)
		}
		existing.refs++
		return &handle{mount: existing, writable: writable}, nil
	}

	m, err := attach(key, vol, writable)
	if err != nil {
		return nil, err
	}
	m.refs = 1
	mounts[key] = m

	return &handle{mount: m, writable: writable}, nil
}

// State reports whether the file browser has an image open, and whether for
// writing. The drive layer asks before it hands the image to the host.
func State(imagePath string) (open, writable bool) {
	mountsMu.Lock()
	defer mountsMu.Unlock()

	m, ok := mounts[keyOf(imagePath)]
	if !ok {
		return false, false
	}
	return true, m.writable
}

func attach(imagePath string, vol Volume, writable bool) (*mount, error) {
	if err := os.MkdirAll(MountRoot, 0o700); err != nil {
		return nil, fmt.Errorf("create %s: %w", MountRoot, err)
	}

	mountpoint := filepath.Join(MountRoot, mountName(imagePath))
	if err := os.Mkdir(mountpoint, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("create %s: %w", mountpoint, err)
	}

	loop, err := attachLoop(imagePath, vol.Offset, vol.Size, writable)
	if err != nil {
		_ = os.Remove(mountpoint)
		return nil, err
	}

	if err := mountVolume(loop.Path, mountpoint, vol, writable); err != nil {
		loop.Detach()
		_ = os.Remove(mountpoint)
		return nil, err
	}
	// The mount holds the device now; with autoclear set, its unmount frees it.
	loop.Close()

	t, err := openTree(mountpoint, vol.Kind, writable)
	if err != nil {
		unmount(mountpoint)
		return nil, err
	}

	return &mount{
		path:       imagePath,
		mountpoint: mountpoint,
		tree:       t,
		writable:   writable,
	}, nil
}

// fsCandidates are the kernel filesystems that can read a volume, in the order
// to try them, with the mount options each prefers. A bridge disc is read
// through UDF where the kernel has it: its ISO 9660 side of a Windows
// installer holds only a README. Slot B's kernel has no UDF, and an ISO 9660
// mount reads what that side holds.
func fsCandidates(vol Volume) []struct{ fs, options string } {
	type c = struct{ fs, options string }
	switch vol.Kind {
	case KindFAT12, KindFAT16, KindFAT32:
		// Long names in UTF-8 without the kernel's NLS_UTF8 table, which
		// slot A does not have.
		return []c{{"vfat", "utf8"}}
	case KindExFAT:
		return []c{{"exfat", ""}}
	case KindUDF:
		return []c{{"udf", "utf8"}}
	case KindISO9660:
		if vol.UDFBridge {
			return []c{{"udf", "utf8"}, {"iso9660", "utf8"}}
		}
		return []c{{"iso9660", "utf8"}}
	}
	return nil
}

func mountVolume(device, mountpoint string, vol Volume, writable bool) error {
	// Neither FAT nor exFAT can hold a link, but the image is untrusted and
	// the server runs as root: refusing to follow one keeps every path inside
	// the mount from resolving outside it, whatever the filesystem.
	flags := uintptr(unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV | unix.MS_NOSYMFOLLOW)
	if !writable {
		flags |= unix.MS_RDONLY
	}

	var errs []error
	for _, c := range fsCandidates(vol) {
		err := unix.Mount(device, mountpoint, c.fs, flags, c.options)
		// A kernel too old for an option answers EINVAL; try without it.
		if errors.Is(err, unix.EINVAL) && c.options != "" {
			err = unix.Mount(device, mountpoint, c.fs, flags, "")
		}
		if err == nil {
			return nil
		}
		if errors.Is(err, unix.ENODEV) {
			err = fmt.Errorf("this kernel has no %s driver", c.fs)
		}
		errs = append(errs, fmt.Errorf("%s: %w", c.fs, err))
	}
	return fmt.Errorf("%w: %w", ErrUnsupported, errors.Join(errs...))
}

// mountName builds a filesystem safe directory name that still says which
// image it belongs to.
func mountName(imagePath string) string {
	sum := sha256.Sum256([]byte(imagePath))

	base := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, filepath.Base(imagePath))
	if len(base) > 64 {
		base = base[:64]
	}

	return base + "-" + hex.EncodeToString(sum[:4])
}

// handle is one user's view of a mount. A handle opened for reading cannot
// write, even when it shares a writable mount.
type handle struct {
	mount    *mount
	writable bool
	once     sync.Once
}

func (h *handle) Kind() Kind     { return h.mount.tree.kind }
func (h *handle) Writable() bool { return h.writable }

func (h *handle) List(rel string, limit int) ([]Entry, bool, error) {
	return h.mount.tree.List(rel, limit)
}

func (h *handle) Open(rel string) (*os.File, Entry, error) {
	return h.mount.tree.Open(rel)
}

func (h *handle) Write(rel string, r io.Reader) (int64, error) {
	if !h.writable {
		return 0, ErrReadOnly
	}
	return h.mount.tree.Write(rel, r)
}

func (h *handle) Remove(rel string) error {
	if !h.writable {
		return ErrReadOnly
	}
	return h.mount.tree.Remove(rel)
}

func (h *handle) StatFS() (total, available uint64, err error) {
	return h.mount.tree.StatFS()
}

// Close releases the handle; the last one unmounts the image.
func (h *handle) Close() error {
	var err error
	h.once.Do(func() {
		mountsMu.Lock()
		defer mountsMu.Unlock()

		h.mount.refs--
		if h.mount.refs > 0 {
			return
		}
		delete(mounts, h.mount.path)
		_ = h.mount.tree.Close()
		err = unmount(h.mount.mountpoint)
	})
	return err
}

// unmount takes a mount down and removes its mount point. A file still open
// on it must not keep it, and the loop device under it, busy for good, so a
// busy mount is detached lazily.
func unmount(mountpoint string) error {
	var err error
	if unix.Unmount(mountpoint, 0) != nil {
		if lazyErr := unix.Unmount(mountpoint, unix.MNT_DETACH); lazyErr != nil {
			err = fmt.Errorf("unmount %s: %w", mountpoint, lazyErr)
		}
	}
	_ = os.Remove(mountpoint)
	return err
}

// CleanupStale takes down every mount left under MountRoot. At startup
// nothing can be mounted by this process yet, so whatever is there is left
// from a crash or a restart; the loop devices under those mounts were set up
// with autoclear and go with them.
func CleanupStale() {
	mountsMu.Lock()
	defer mountsMu.Unlock()

	for _, mountpoint := range staleMountpoints() {
		if err := unix.Unmount(mountpoint, unix.MNT_DETACH); err != nil {
			logrus.Errorf("clean up stale mount %s failed: %s", mountpoint, err)
		}
	}

	// Mount points exist only while an image is mounted, so whatever is
	// still in the root is a leftover. Removing a directory fails while
	// something is mounted on it, which is the safety wanted.
	if entries, err := os.ReadDir(MountRoot); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				_ = os.Remove(filepath.Join(MountRoot, entry.Name()))
			}
		}
	}

	for _, m := range mounts {
		_ = m.tree.Close()
	}
	mounts = map[string]*mount{}
}

func staleMountpoints() []string {
	data, err := os.ReadFile("/proc/self/mounts")
	if err != nil {
		return nil
	}

	var paths []string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		// /proc escapes spaces in mount points; ours never contain any.
		mountpoint := fields[1]
		if strings.HasPrefix(mountpoint, MountRoot+"/") {
			paths = append(paths, mountpoint)
		}
	}
	return paths
}
