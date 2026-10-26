package settingsio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// importMu keeps two imports from interleaving their writes.
var importMu sync.Mutex

// Export reads every section from the board. A section that cannot be read
// is left out of the file and named in skipped, so one damaged file does not
// stop the export of the rest.
func (d *Device) Export(source Source, now time.Time) (doc *Document, skipped []string, err error) {
	snap := d.snapshot()
	doc = &Document{
		Format:     Format,
		Version:    Version,
		ExportedAt: now.UTC().Format(time.RFC3339),
		Source:     source,
		Sections:   map[string]json.RawMessage{},
	}
	for _, sec := range sections {
		value, err := sec.current(snap)
		if err != nil {
			log.Warnf("settings export: section %s: %s", sec.name(), err)
			skipped = append(skipped, sec.name())
			continue
		}
		raw, err := json.Marshal(value)
		if err != nil {
			return nil, nil, err
		}
		doc.Sections[sec.name()] = raw
	}
	return doc, skipped, nil
}

// The statuses of a section in a preview.
const (
	StatusChange  = "change"
	StatusSame    = "same"
	StatusInvalid = "invalid"
	StatusUnknown = "unknown"
)

// Change is one field the import changes. From and To are the values as
// JSON would hold them; a long list is summarised by its length.
type Change struct {
	Field string `json:"field"`
	From  any    `json:"from"`
	To    any    `json:"to"`
}

// SectionPreview is what the import does to one section.
type SectionPreview struct {
	Section  string   `json:"section"`
	Status   string   `json:"status"`
	Changes  []Change `json:"changes,omitempty"`
	Error    string   `json:"error,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// Preview is what an import would do, section by section.
type Preview struct {
	Version    int              `json:"version"`
	ExportedAt string           `json:"exportedAt,omitempty"`
	Source     Source           `json:"source"`
	Sections   []SectionPreview `json:"sections"`
}

// decoded is one section of a document, decoded and checked.
type decoded struct {
	sec      section
	current  any
	value    any
	warnings []string
	err      error
}

// decodeAll decodes every section of doc that this server knows, checks each
// one alone and then against the others, and returns them in allowlist order
// with the names of the sections it does not know.
func decodeAll(snap *snapshot, doc *Document) ([]*decoded, []string) {
	var known []*decoded
	for _, sec := range sections {
		raw, ok := doc.Sections[sec.name()]
		if !ok {
			continue
		}
		item := &decoded{sec: sec}
		if cur, err := sec.current(snap); err == nil {
			item.current = cur
		}
		item.value, item.warnings, item.err = sec.decode(snap, raw)
		known = append(known, item)
	}

	crossCheck(snap, known)

	var unknown []string
	for name := range doc.Sections {
		if _, ok := sectionByName(name); !ok {
			unknown = append(unknown, name)
		}
	}
	sort.Strings(unknown)
	return known, unknown
}

// crossCheck applies the rules that tie one section to another. Each one
// reads the other section as the import would leave it: the imported value
// when the file carries the section, the board's own otherwise.
func crossCheck(snap *snapshot, items []*decoded) {
	byName := map[string]*decoded{}
	for _, item := range items {
		byName[item.sec.name()] = item
	}

	powerLED := false
	if item, ok := byName["powerLed"]; ok && item.err == nil {
		powerLED = item.value.(powerLEDSettings).PowerLED
	} else if cur, err := powerLEDSection.get(snap); err == nil {
		powerLED = cur.PowerLED
	}

	// Without the power LED the board cannot tell a host that is off from one
	// that is hung (service/watchdog/api.go, service/nut/api.go).
	if item, ok := byName["watchdog"]; ok && item.err == nil {
		w := item.value.(watchdogSettings).config().WithDefaults()
		if w.Enabled && w.Action == "power" && !powerLED {
			item.err = errors.New("a power cycle needs the power LED connected")
		}
	}
	if item, ok := byName["nut"]; ok && item.err == nil {
		if item.value.(nutSettings).Enabled && !powerLED {
			item.err = errors.New("the NUT client needs the power LED connected")
		}
	}
}

// Preview says what importing doc would change. It writes nothing.
func (d *Device) Preview(doc *Document) Preview {
	snap := d.snapshot()
	items, unknown := decodeAll(snap, doc)

	out := Preview{Version: doc.Version, ExportedAt: doc.ExportedAt, Source: doc.Source, Sections: []SectionPreview{}}
	for _, item := range items {
		p := SectionPreview{Section: item.sec.name(), Warnings: item.warnings}
		if item.err != nil {
			p.Status = StatusInvalid
			p.Error = item.err.Error()
		} else {
			p.Changes = diff(item.current, item.value)
			p.Status = StatusSame
			if len(p.Changes) > 0 {
				p.Status = StatusChange
			}
		}
		out.Sections = append(out.Sections, p)
	}
	for _, name := range unknown {
		out.Sections = append(out.Sections, SectionPreview{
			Section:  name,
			Status:   StatusUnknown,
			Warnings: []string{"this firmware does not know the section; it is skipped"},
		})
	}
	return out
}

// ImportError is an import refused before anything was written. Sections
// holds the reason for each section that failed.
type ImportError struct {
	Sections map[string]string
}

func (e *ImportError) Error() string {
	names := make([]string, 0, len(e.Sections))
	for name := range e.Sections {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = name + ": " + e.Sections[name]
	}
	return "nothing was imported: " + strings.Join(parts, "; ")
}

// Import writes the chosen sections of doc. With no choice it writes every
// section this server knows. Every chosen section is checked before the first
// write, and one that fails stops the whole import. It returns the sections
// it wrote. The running services keep their settings until the next start.
func (d *Device) Import(doc *Document, chosen []string) ([]string, error) {
	importMu.Lock()
	defer importMu.Unlock()

	snap := d.snapshot()
	items, _ := decodeAll(snap, doc)

	want := map[string]bool{}
	for _, name := range chosen {
		want[name] = true
	}
	failed := map[string]string{}
	for name := range want {
		if _, ok := sectionByName(name); !ok {
			failed[name] = "this firmware does not know the section"
		} else if _, ok := doc.Sections[name]; !ok {
			failed[name] = "the file does not hold the section"
		}
	}

	var selected []*decoded
	for _, item := range items {
		if len(want) > 0 && !want[item.sec.name()] {
			continue
		}
		if item.err != nil {
			failed[item.sec.name()] = item.err.Error()
			continue
		}
		selected = append(selected, item)
	}
	if len(failed) > 0 {
		return nil, &ImportError{Sections: failed}
	}
	if len(selected) == 0 {
		return nil, errors.New("the file holds no section to import")
	}

	p := newPlan(snap)
	applied := make([]string, 0, len(selected))
	for _, item := range selected {
		if err := item.sec.stage(p, item.value); err != nil {
			return nil, fmt.Errorf("section %s: %w", item.sec.name(), err)
		}
		applied = append(applied, item.sec.name())
	}
	changes, err := p.changes()
	if err != nil {
		return nil, err
	}
	if err := d.commit(changes); err != nil {
		return nil, fmt.Errorf("writing the settings failed, and the files written so far were put back: %w", err)
	}
	return applied, nil
}

// maxPreviewValue is the longest JSON a value may have in a preview before it
// is summarised. A list of macros would otherwise fill the dialog.
const maxPreviewValue = 256

// diff returns the fields that differ between two values of one section, as
// dotted paths into their JSON form.
func diff(from, to any) []Change {
	a := flatten(toJSONValue(from))
	b := flatten(toJSONValue(to))

	keys := map[string]bool{}
	for k := range a {
		keys[k] = true
	}
	for k := range b {
		keys[k] = true
	}
	names := make([]string, 0, len(keys))
	for k := range keys {
		names = append(names, k)
	}
	sort.Strings(names)

	var changes []Change
	for _, k := range names {
		x, y := a[k], b[k]
		if jsonEqual(x, y) {
			continue
		}
		changes = append(changes, Change{Field: k, From: summarise(x), To: summarise(y)})
	}
	return changes
}

func toJSONValue(v any) any {
	if v == nil {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var out any
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if dec.Decode(&out) != nil {
		return nil
	}
	return out
}

// flatten turns nested objects into dotted paths. A list stays one value.
func flatten(v any) map[string]any {
	out := map[string]any{}
	if v == nil {
		return out
	}
	var walk func(prefix string, v any)
	walk = func(prefix string, v any) {
		obj, ok := v.(map[string]any)
		if !ok || (len(obj) == 0 && prefix != "") {
			out[prefix] = v
			return
		}
		for k, child := range obj {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			walk(key, child)
		}
	}
	walk("", v)
	return out
}

func jsonEqual(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return bytes.Equal(x, y)
}

// ListSummary stands in for a list too long to show in a preview.
type ListSummary struct {
	Items int `json:"items"`
}

func summarise(v any) any {
	list, ok := v.([]any)
	if !ok {
		return v
	}
	if data, _ := json.Marshal(list); len(data) <= maxPreviewValue {
		return v
	}
	return ListSummary{Items: len(list)}
}
