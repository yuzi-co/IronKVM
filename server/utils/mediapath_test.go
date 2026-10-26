package utils

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateMediumPathNormalizes(t *testing.T) {
	tests := map[string]string{
		"":                "/",
		"/":               "/",
		"  /  ":           "/",
		".":               "/",
		"docs":            "docs",
		"/docs":           "docs",
		"/docs/":          "docs",
		"docs/./a.txt":    "docs/a.txt",
		"docs/sub/../a":   "docs/a",
		"/EFI/BOOT/x.efi": "EFI/BOOT/x.efi",
	}
	for in, want := range tests {
		got, err := ValidateMediumPath(in)
		if err != nil {
			t.Errorf("ValidateMediumPath(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ValidateMediumPath(%q) = %q, want %q", in, got, want)
		}
	}
}

// A path names something inside the medium. Every way out of it, and every
// name a FAT or exFAT volume cannot hold, is refused before the filesystem
// sees it.
func TestValidateMediumPathRefusesEscapesAndBadNames(t *testing.T) {
	bad := []string{
		"..",
		"../etc/shadow",
		"/../etc",
		"docs/../../etc",
		"a/../..",
		"docs\x00/a",
		`docs\..\..\etc`,
		"a:b",
		"CON",
		"aux.txt",
		"name.",
		"tab\there",
		strings.Repeat("n", 256),
	}
	for _, in := range bad {
		if got, err := ValidateMediumPath(in); !errors.Is(err, ErrInvalidMediaPath) {
			t.Errorf("ValidateMediumPath(%q) = %q, %v; want ErrInvalidMediaPath", in, got, err)
		}
	}
}

func TestValidateFileName(t *testing.T) {
	good := []string{"a.txt", "report 2026.pdf", "Ünïcødé.bin", ".hidden", strings.Repeat("n", 255)}
	for _, name := range good {
		if err := ValidateFileName(name); err != nil {
			t.Errorf("ValidateFileName(%q): %v", name, err)
		}
	}

	bad := []string{"", ".", "..", "name ", "name.", "a/b", `a\b`, "a?b", "nul", "LPT1.log", "x\x7f", strings.Repeat("n", 256)}
	for _, name := range bad {
		if err := ValidateFileName(name); err == nil {
			t.Errorf("ValidateFileName(%q) accepted it", name)
		}
	}
}
