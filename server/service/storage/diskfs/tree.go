// Parts of this file are ported from sipeed/NanoKVM
// (server/service/storage/diskfs/mount.go, commit 5572e1de, PR #957).
// Copyright (c) Sipeed. Licensed under the GNU General Public License v3.0.

package diskfs

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sys/unix"

	"NanoKVM-Server/utils"
)

// tree is a filesystem the file browser works in: the mount point of an
// image, or in tests a plain directory. Every path goes through os.Root, so
// nothing in it - a "..", a link the image carries - resolves outside it. The
// image is untrusted and the server runs as root.
type tree struct {
	root     *os.Root
	dir      string
	kind     Kind
	writable bool

	closeOnce sync.Once
	closeErr  error
}

// OpenDir opens a directory as an image's filesystem, without mounting
// anything. The file browser uses it for what it has mounted itself; tests
// use it for a temporary directory.
func OpenDir(dir string, kind Kind, writable bool) (Image, error) {
	return openTree(dir, kind, writable)
}

func openTree(dir string, kind Kind, writable bool) (*tree, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	return &tree{root: root, dir: dir, kind: kind, writable: writable}, nil
}

func (t *tree) Kind() Kind     { return t.kind }
func (t *tree) Writable() bool { return t.writable }

// cleanRel turns a path relative to the image root into the form os.Root
// takes: "." for the root, no leading slash. A path that climbs above the root
// at any step is refused here, where path.Clean would quietly clamp it; os.Root
// refuses the escapes left, such as a link that leads out.
func cleanRel(rel string) (string, error) {
	if strings.ContainsRune(rel, 0) || !isInside(rel) {
		return "", fmt.Errorf("%w: %q escapes the image", utils.ErrInvalidMediaPath, rel)
	}
	name := strings.TrimPrefix(path.Clean("/"+rel), "/")
	if name == "" {
		return ".", nil
	}
	return name, nil
}

// readDirChunk is how many entries a listing reads at a time.
const readDirChunk = 256

func (t *tree) List(rel string, limit int) ([]Entry, bool, error) {
	name, err := cleanRel(rel)
	if err != nil {
		return nil, false, err
	}

	dir, err := t.root.Open(name)
	if err != nil {
		return nil, false, err
	}
	defer func() {
		_ = dir.Close()
	}()

	entries := []Entry{}
	truncated := false
	for !truncated {
		dirents, err := dir.ReadDir(readDirChunk)
		for _, dirent := range dirents {
			// A partial upload is not a file yet, and a link or a device
			// is nothing the browser can hand out.
			if IsTransferTemp(dirent.Name()) || !(dirent.Type().IsRegular() || dirent.IsDir()) {
				continue
			}
			if len(entries) == limit {
				truncated = true
				break
			}
			info, err := dirent.Info()
			if err != nil {
				continue
			}
			entries = append(entries, entryOf(info))
		}
		if errors.Is(err, io.EOF) || len(dirents) == 0 {
			break
		}
		if err != nil {
			return nil, false, err
		}
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	return entries, truncated, nil
}

// isInside reports whether rel, taken from the image root, never climbs above
// it at any step.
func isInside(rel string) bool {
	depth := 0
	for _, part := range strings.Split(rel, "/") {
		switch part {
		case "", ".":
		case "..":
			depth--
			if depth < 0 {
				return false
			}
		default:
			depth++
		}
	}
	return true
}

func (t *tree) Open(rel string) (*os.File, Entry, error) {
	name, err := cleanRel(rel)
	if err != nil {
		return nil, Entry{}, err
	}

	// The entry is opened before it is looked at, and a link is refused
	// rather than followed, so what is checked is what gets streamed.
	file, err := t.root.OpenFile(name, os.O_RDONLY|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, Entry{}, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, Entry{}, err
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, Entry{}, fmt.Errorf("%s is not a regular file", rel)
	}

	return file, entryOf(info), nil
}

// Write stores r as rel inside the image and reports how many bytes landed.
// The bytes go to a hidden name first and are renamed into place, so a failed
// or cancelled upload never leaves a half written file where the host will
// see it.
func (t *tree) Write(rel string, r io.Reader) (int64, error) {
	if !t.writable {
		return 0, ErrReadOnly
	}
	name, err := cleanRel(rel)
	if err != nil {
		return 0, err
	}
	if name == "." {
		return 0, fmt.Errorf("%w: a file name is required", utils.ErrInvalidMediaPath)
	}

	if _, err := t.root.Lstat(name); err == nil {
		return 0, fmt.Errorf("%s: %w", rel, os.ErrExist)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return 0, err
	}

	tempName := path.Join(path.Dir(name), UploadPrefix+randomSuffix())
	file, err := t.root.OpenFile(tempName, os.O_WRONLY|os.O_CREATE|os.O_EXCL|unix.O_NOFOLLOW, 0o644)
	if err != nil {
		return 0, err
	}

	written, err := io.Copy(file, r)
	if err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = t.root.Remove(tempName)
		return 0, err
	}

	if err := t.root.Rename(tempName, name); err != nil {
		_ = t.root.Remove(tempName)
		return 0, err
	}

	return written, nil
}

// Remove deletes a file, or a directory with everything it holds: the web UI
// confirms the delete first, and answering an explicit yes with "directory not
// empty" helps no one. os.Root keeps the walk inside the image.
func (t *tree) Remove(rel string) error {
	if !t.writable {
		return ErrReadOnly
	}
	name, err := cleanRel(rel)
	if err != nil {
		return err
	}
	if name == "." {
		return fmt.Errorf("%w: refusing to remove the image root", utils.ErrInvalidMediaPath)
	}
	// RemoveAll succeeds on a path that is not there; a delete of nothing is
	// a mistake the caller should hear about.
	if _, err := t.root.Lstat(name); err != nil {
		return err
	}
	return t.root.RemoveAll(name)
}

func (t *tree) StatFS() (total, available uint64, err error) {
	return utils.Statfs(t.dir)
}

func (t *tree) Close() error {
	t.closeOnce.Do(func() {
		t.closeErr = t.root.Close()
	})
	return t.closeErr
}

func entryOf(info os.FileInfo) Entry {
	return Entry{
		Name:    info.Name(),
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime(),
	}
}

func randomSuffix() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "00000000"
	}
	return hex.EncodeToString(buf)
}
