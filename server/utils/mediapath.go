// Ported from sipeed/NanoKVM (server/utils/mediapath.go, commit 5572e1de,
// PR #957), reduced to the checks on a path inside an image. Images themselves
// are named by their path under /data, as the drive API names them.
// Copyright (c) Sipeed. Licensed under the GNU General Public License v3.0.

package utils

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"unicode/utf8"
)

// maxNameLength is the FAT/exFAT limit in UTF-16 code units; counting runes is
// the closest practical approximation on the Go side.
const maxNameLength = 255

// forbiddenNameChars cannot appear in a name on any platform that mounts
// FAT/exFAT, so the file browser rejects them before the kernel sees them.
const forbiddenNameChars = `<>:"/\|?*`

// reservedNames are the DOS device names Windows still refuses, with or
// without an extension.
var reservedNames = map[string]bool{
	"CON": true, "PRN": true, "AUX": true, "NUL": true,
	"COM1": true, "COM2": true, "COM3": true, "COM4": true, "COM5": true,
	"COM6": true, "COM7": true, "COM8": true, "COM9": true,
	"LPT1": true, "LPT2": true, "LPT3": true, "LPT4": true, "LPT5": true,
	"LPT6": true, "LPT7": true, "LPT8": true, "LPT9": true,
}

// ErrInvalidMediaPath reports a path that does not name something inside an
// image, or a name the image's filesystem cannot hold.
var ErrInvalidMediaPath = errors.New("invalid media path")

// ValidateMediumPath normalizes a path inside an image and returns it slash
// separated and relative to the image's root ("/" means the root itself). It
// takes the same shape the web UI shows, so a leading slash is accepted and
// stripped. A path that climbs out of the image is refused rather than
// clamped, and so is every name a FAT or exFAT volume cannot hold: the host
// that reads the image next would choke on it.
func ValidateMediumPath(p string) (string, error) {
	if strings.ContainsRune(p, 0) {
		return "", fmt.Errorf("%w: contains NUL", ErrInvalidMediaPath)
	}

	p = strings.TrimPrefix(strings.TrimSpace(p), "/")
	if p == "" {
		return "/", nil
	}

	clean := path.Clean(p)
	if clean == "." {
		return "/", nil
	}
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("%w: %s escapes the image", ErrInvalidMediaPath, p)
	}

	for _, segment := range strings.Split(clean, "/") {
		if err := validateName(segment); err != nil {
			return "", fmt.Errorf("%w: %s", ErrInvalidMediaPath, err)
		}
	}

	return clean, nil
}

// ValidateFileName checks a single file name for FAT/exFAT compatibility. It
// is what an upload's browser-supplied name has to pass before it is written
// into an image.
func ValidateFileName(name string) error {
	if err := validateName(name); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidMediaPath, err)
	}
	return nil
}

func validateName(name string) error {
	if name == "" {
		return errors.New("empty name")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("reserved name %q", name)
	}
	if utf8.RuneCountInString(name) > maxNameLength {
		return fmt.Errorf("name longer than %d characters", maxNameLength)
	}
	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
		return fmt.Errorf("name %q ends with a dot or space", name)
	}

	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("name %q contains a control character", name)
		}
		if strings.ContainsRune(forbiddenNameChars, r) {
			return fmt.Errorf("name %q contains %q", name, r)
		}
	}

	base := name
	if i := strings.IndexByte(base, '.'); i >= 0 {
		base = base[:i]
	}
	if reservedNames[strings.ToUpper(base)] {
		return fmt.Errorf("name %q is reserved on Windows", name)
	}

	return nil
}
