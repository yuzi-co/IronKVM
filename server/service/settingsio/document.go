// Package settingsio exports the board's settings to a file and imports such
// a file, on the same board or on another one.
//
// The file is built from an allowlist: each section below copies named fields
// out of server.yaml or out of a named file in /etc/kvm. Nothing reaches the
// file unless a section names it, so a secret added to the configuration later
// stays out by default. docs/superpowers/specs/2026-10-26-settings-export-
// diagnostics-design.md lists what is exported, what is not, and why.
package settingsio

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	// Format names the file, so an import can refuse a file of another kind.
	Format = "ironkvm-settings"

	// Version is the version of the file format. Raise it when a section
	// changes in a way an older server cannot read. A new section does not
	// raise it: an older server skips a section it does not know.
	Version = 1

	// MaxDocumentBytes bounds a document an import accepts. Fifty macros at
	// their largest are some 5 MB of JSON.
	MaxDocumentBytes = 8 << 20
)

// Source says where a file came from. It is information for the person who
// reads the file; the import does not use it.
type Source struct {
	Application string `json:"application,omitempty"`
	Hostname    string `json:"hostname,omitempty"`
}

// Document is the settings file.
type Document struct {
	Format     string                     `json:"format"`
	Version    int                        `json:"version"`
	ExportedAt string                     `json:"exportedAt,omitempty"`
	Source     Source                     `json:"source"`
	Sections   map[string]json.RawMessage `json:"sections"`
}

// ErrNewerVersion is returned for a file that a newer server wrote.
var ErrNewerVersion = errors.New("newer settings file")

// ParseDocument reads a settings file and checks its format and version. It
// does not look inside the sections.
func ParseDocument(data []byte) (*Document, error) {
	if len(data) > MaxDocumentBytes {
		return nil, fmt.Errorf("the file is larger than %d MB", MaxDocumentBytes>>20)
	}

	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, errors.New("the file is not a settings file: it is not valid JSON")
	}

	return &doc, doc.check()
}

func (doc *Document) check() error {
	switch {
	case doc.Format != Format:
		return errors.New("the file is not an IronKVM settings file")
	case doc.Version > Version:
		return fmt.Errorf("%w: this file comes from newer firmware (format version %d, this board reads up to %d); update this board first",
			ErrNewerVersion, doc.Version, Version)
	case doc.Version < 1:
		return fmt.Errorf("the file has no valid format version (%d)", doc.Version)
	case doc.Sections == nil:
		return errors.New("the file has no sections")
	}
	return nil
}
