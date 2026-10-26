// Package licenses reads the image's license manifest, the directory that
// ironkvm-dist's licenses/make-manifest.sh writes into the root filesystem:
// MANIFEST, COMPONENTS, NOTICE and the license texts under texts/.
//
// It hands out files from that directory and nothing else. A name is checked
// before it is opened, and it is opened through an os.Root, so neither "..",
// an absolute path nor a symlink that points out of the directory reaches a
// file outside it.
package licenses

import (
	"errors"
	"io/fs"
	"os"
	"strings"
)

// DefaultDir is where the image keeps the manifest. A development build that
// was not made by ironkvm-dist has no such directory.
const DefaultDir = "/usr/share/licenses/ironkvm"

// MaxFileSize caps what Open hands out. NOTICE is the largest file, a few
// hundred kilobytes; the cap is there so a stray file cannot tie the server up.
const MaxFileSize = 8 << 20

// maxEntries caps the listing, for the same reason.
const maxEntries = 4096

var (
	// ErrNotInstalled means the directory does not exist.
	ErrNotInstalled = errors.New("license manifest not installed on this image")
	// ErrBadPath means the name is not a plain relative path inside the
	// directory, or it leads out of it.
	ErrBadPath = errors.New("invalid license file path")
	// ErrNotFound means no regular file has that name.
	ErrNotFound = errors.New("no such license file")
	// ErrTooLarge means the file is over MaxFileSize.
	ErrTooLarge = errors.New("license file too large")
)

// File is one entry of the listing.
type File struct {
	// Path is relative to the directory, with forward slashes.
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// Manifest reads one manifest directory.
type Manifest struct {
	dir string
}

// New reads the manifest in dir.
func New(dir string) *Manifest {
	return &Manifest{dir: dir}
}

func (m *Manifest) openRoot() (*os.Root, error) {
	root, err := os.OpenRoot(m.dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotInstalled
		}
		return nil, err
	}
	return root, nil
}

// List names every regular file in the directory, in lexical order. Symlinks
// are left out.
func (m *Manifest) List() ([]File, error) {
	root, err := m.openRoot()
	if err != nil {
		return nil, err
	}
	defer root.Close()

	files := []File{}
	err = fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if len(files) == maxEntries {
			return fs.SkipAll
		}
		files = append(files, File{Path: path, Size: info.Size()})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// validName reports whether name is a plain relative path: no "." or ".."
// element, no leading or doubled slash, no backslash and no NUL.
func validName(name string) bool {
	return name != "." && fs.ValidPath(name) &&
		!strings.ContainsAny(name, "\\\x00")
}

// Open opens a regular file of the directory by its relative path and
// returns it with its size. The caller closes the file.
func (m *Manifest) Open(name string) (*os.File, int64, error) {
	if !validName(name) {
		return nil, 0, ErrBadPath
	}
	root, err := m.openRoot()
	if err != nil {
		return nil, 0, err
	}
	defer root.Close()

	file, err := root.Open(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, 0, ErrNotFound
		}
		// os.Root refuses a symlink that leads out of the directory.
		return nil, 0, ErrBadPath
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, err
	}
	if !info.Mode().IsRegular() {
		file.Close()
		return nil, 0, ErrNotFound
	}
	if info.Size() > MaxFileSize {
		file.Close()
		return nil, 0, ErrTooLarge
	}
	return file, info.Size(), nil
}
