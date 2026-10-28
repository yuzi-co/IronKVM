package install

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// Manifest describes one version of an add-on. The server compiles in the
// manifests of the versions it was tested with, today's pins; a signed index
// can add newer ones later (the note's step 4).
type Manifest struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Version string `json:"version"`
	Licence string `json:"licence"`
	// Source is where the add-on's source can be had, for the owner and for
	// the licence's offer.
	Source   string   `json:"source"`
	Artifact Artifact `json:"artifact"`
}

// Artifact is the file downloaded and what is taken out of it.
type Artifact struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	// Size is the download's size and Unpacked the members' total, for the
	// free space check before the download. Zero skips the check.
	Size     int64 `json:"size,omitempty"`
	Unpacked int64 `json:"unpacked,omitempty"`
	// Max bounds the download itself.
	Max    int64  `json:"max"`
	Format Format `json:"format"`
	// Members are the only files taken out of the artifact.
	Members []Member `json:"members"`
}

// Format is how the artifact is read. An apk is also checked: apk verifies
// Alpine's signature on it before anything is taken out.
type Format string

const (
	FormatTarGz Format = "tar.gz"
	FormatApk   Format = "apk"
)

// Member is a file taken out of the artifact.
type Member struct {
	// Path is the member's path inside the artifact. With Licences it is a
	// directory, ending in "/", and every regular file directly in it that is
	// not another member is a licence file.
	Path string `json:"path"`
	// Name is the name it is saved under, a plain name. The archive's own
	// path never becomes a path on the board.
	Name string `json:"name,omitempty"`
	// SHA256 is the member's sum. Empty only where the upstream publishes
	// none and the artifact's own sum, or its signature, covers it.
	SHA256 string `json:"sha256,omitempty"`
	// Mode is the file's mode in the image; 0644 when zero.
	Mode os.FileMode `json:"mode,omitempty"`
	// XZ members are decompressed, and SHA256 is the sum of the result.
	XZ bool `json:"xz,omitempty"`
	// Max bounds the member, decompressed. DefaultMemberMax when zero.
	Max      int64 `json:"max,omitempty"`
	Licences bool  `json:"licences,omitempty"`
}

// DefaultMemberMax bounds a member with no Max of its own.
const DefaultMemberMax = 128 << 20

// licenceMax bounds one licence file.
const licenceMax = 1 << 20

// LicencesDir is the directory the licence files are saved under, in the
// staged files, in the image, and copied out of it beside the image.
const LicencesDir = "licenses"

func (m Member) max() int64 {
	if m.Max > 0 {
		return m.Max
	}
	return DefaultMemberMax
}

func (m Member) mode() os.FileMode {
	if m.Mode != 0 {
		return m.Mode.Perm()
	}
	return 0o644
}

func (m Manifest) validate() error {
	if !plainName(m.Name) {
		return fmt.Errorf("add-on name %q is not a plain name", m.Name)
	}
	if m.Version == "" || strings.Trim(m.Version, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.+_-") != "" {
		return fmt.Errorf("%s: version %q has characters a file name may not", m.Name, m.Version)
	}
	switch m.Artifact.Format {
	case FormatTarGz, FormatApk:
	default:
		return fmt.Errorf("%s: no reader for the format %q", m.Name, m.Artifact.Format)
	}
	return validateMembers(m.Artifact.Members)
}

// validateMembers refuses a member that could write anywhere but its own
// name in the staging directory, or be read from outside the artifact.
func validateMembers(members []Member) error {
	if len(members) == 0 {
		return fmt.Errorf("the manifest lists no member")
	}
	seen := map[string]bool{}
	for _, m := range members {
		if !safePath(strings.TrimSuffix(m.Path, "/")) {
			return fmt.Errorf("member path %q leaves the artifact", m.Path)
		}
		if m.Licences {
			if !strings.HasSuffix(m.Path, "/") {
				return fmt.Errorf("licence member %q is not a directory", m.Path)
			}
			continue
		}
		if !plainName(m.Name) || m.Name == LicencesDir {
			return fmt.Errorf("member %s is saved under %q, which is not a plain name", m.Path, m.Name)
		}
		if seen[m.Name] {
			return fmt.Errorf("two members are saved as %s", m.Name)
		}
		seen[m.Name] = true
	}
	return nil
}

// plainName is a file name with no directory and nothing a shell, a pseudo
// file definition or a line of S04addons would split.
func plainName(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, "/\\ \t\n\"'")
}

// safePath is a relative path that stays inside the artifact: no leading
// slash and no ".." element. "./" prefixes, which tar writes, are allowed.
func safePath(p string) bool {
	if p == "" || strings.HasPrefix(p, "/") {
		return false
	}
	for _, e := range strings.Split(p, "/") {
		if e == ".." {
			return false
		}
	}
	return path.Clean(p) != ".."
}
