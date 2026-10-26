//go:build ignore

// gen_layouts writes the generated keyboard layouts of the paste-as-typing
// feature. It reads each layout from xkeyboard-config through libxkbcommon,
// presses every key of a 105-key keyboard (and the JIS and ABNT2 extras) with
// and without Shift and AltGr, and records what each press types.
//
// Run it through gen_layouts.sh, which pins the container image and the
// versions of xkeyboard-config, libxkbcommon and the X11 compose table:
//
//	sh server/service/hid/gen_layouts.sh
//
// It writes two files beside it:
//
//	layout_gen.go       the layouts that have no hand-written table
//	layout_xkb_test.go  the xkb version of each hand-written layout, and the
//	                    X11 compose results, for the tests that compare them
package main

/*
#cgo pkg-config: xkbcommon
#include <stdlib.h>
#include <xkbcommon/xkbcommon.h>
#include <xkbcommon/xkbcommon-compose.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unsafe"
)

// layoutSpec names one xkb layout and the ID the server gives it.
type layoutSpec struct {
	id      string
	layout  string
	variant string
}

// generated lists the layouts that get a generated table. The IDs follow
// PiKVM's keymap names, so a client that knows those can name them.
var generated = []layoutSpec{
	{"am", "am", ""},
	{"ar", "ara", ""},
	{"bepo", "fr", "bepo"},
	{"cz", "cz", ""},
	{"da", "dk", ""},
	{"de-ch", "ch", "de"},
	{"en-us-altgr-intl", "us", "altgr-intl"},
	{"en-us-colemak", "us", "colemak"},
	{"et", "ee", ""},
	{"fo", "fo", ""},
	{"fr-be", "be", ""},
	{"fr-ca", "ca", ""},
	{"fr-ch", "ch", "fr"},
	{"hr", "hr", ""},
	{"hu", "hu", ""},
	{"is", "is", ""},
	{"lt", "lt", ""},
	{"lv", "lv", ""},
	{"mk", "mk", ""},
	{"nl", "nl", ""},
	{"no", "no", ""},
	{"pl", "pl", ""},
	{"pt", "pt", ""},
	{"sl", "si", ""},
	{"th", "th", ""},
	{"tr", "tr", ""},
}

// compared lists the hand-written layouts and the xkb layout each is held
// against in layout_xkb_test.go. ko is US, so it is not listed twice.
var compared = []layoutSpec{
	{"us", "us", ""},
	{"uk", "gb", ""},
	{"de", "de", ""},
	{"fr", "fr", ""},
	{"es", "es", ""},
	{"it", "it", ""},
	{"pt-br", "br", ""},
	{"se", "se", ""},
	{"fi", "fi", ""},
	{"ru", "ru", ""},
	{"ja", "jp", ""},
}

// keyOrder maps each HID usage the tables use to its evdev keycode, in the
// order a table lists them. Where two keys type the same character, the
// earlier one in this order keeps it.
var keyOrder = []struct{ usage, evdev int }{
	// Letters, usage 4 (A) to 29 (Z).
	{4, 30}, {5, 48}, {6, 46}, {7, 32}, {8, 18}, {9, 33}, {10, 34}, {11, 35},
	{12, 23}, {13, 36}, {14, 37}, {15, 38}, {16, 50}, {17, 49}, {18, 24},
	{19, 25}, {20, 16}, {21, 19}, {22, 31}, {23, 20}, {24, 22}, {25, 47},
	{26, 17}, {27, 45}, {28, 21}, {29, 44},
	// Digits, usage 30 (1) to 39 (0).
	{30, 2}, {31, 3}, {32, 4}, {33, 5}, {34, 6}, {35, 7}, {36, 8}, {37, 9},
	{38, 10}, {39, 11},
	// Grave, minus, equal, brackets, backslash, semicolon, quote, comma,
	// period and slash. Usage 49 is also the ISO key beside Enter: Linux
	// reads usage 50 as the same evdev key, and Windows as the same scancode.
	{53, 41}, {45, 12}, {46, 13}, {47, 26}, {48, 27}, {49, 43}, {51, 39},
	{52, 40}, {54, 51}, {55, 52}, {56, 53},
	// Space. The plain level is always a space, which keysymRune leaves out;
	// some layouts put a character on its other levels (bepo: _ with AltGr).
	{44, 57},
	// The ISO key left of Z (KEY_102ND).
	{100, 86},
	// The JIS Ro key, which is also the ABNT2 key left of Right Shift
	// (KEY_RO), and the JIS Yen key (KEY_YEN).
	{135, 89}, {137, 124},
}

const (
	evdevOffset  = 8 // xkb keycode = evdev keycode + 8
	evdevShift   = 42
	evdevRAlt    = 100
	keysymLevel3 = 0xfe03   // ISO_Level3_Shift
	keysymVoid   = 0xffffff // VoidSymbol, a level that types nothing
	keysymSpace  = 0x20     // space, which every layout types the same
)

// deadAccents maps each dead keysym the tables support to the spacing accent
// that stands for it in a table. A dead key not listed here is left out.
var deadAccents = map[uint32]rune{
	0xfe50: '`', // dead_grave
	0xfe51: '´', // dead_acute
	0xfe52: '^', // dead_circumflex
	0xfe53: '~', // dead_tilde
	0xfe54: '¯', // dead_macron
	0xfe55: '˘', // dead_breve
	0xfe56: '˙', // dead_abovedot
	0xfe57: '¨', // dead_diaeresis
	0xfe58: '˚', // dead_abovering
	0xfe59: '˝', // dead_doubleacute
	0xfe5a: 'ˇ', // dead_caron
	0xfe5b: '¸', // dead_cedilla
	0xfe5c: '˛', // dead_ogonek
}

const deadFlag rune = 1 << 30

type keyRow struct {
	usage  int
	levels [4]rune
}

type genLayout struct {
	spec    layoutSpec
	name    string
	rows    []keyRow
	skipped []string // what a key types that a table cannot hold
}

func main() {
	ctx := C.xkb_context_new(C.XKB_CONTEXT_NO_FLAGS)
	if ctx == nil {
		fail("xkb_context_new failed")
	}
	defer C.xkb_context_unref(ctx)

	var gen, cmp []genLayout
	for _, spec := range generated {
		gen = append(gen, readLayout(ctx, spec))
	}
	for _, spec := range compared {
		cmp = append(cmp, readLayout(ctx, spec))
	}
	compose := readCompose(ctx)

	header := fmt.Sprintf(`// Code generated by gen_layouts.go; DO NOT EDIT.
//
// Source: xkeyboard-config %s, read through libxkbcommon %s with the
// evdev rules and model pc105. Regenerate with:
//
//	sh server/service/hid/gen_layouts.sh
//
// The key tables are derived from xkeyboard-config, which is distributed
// under the MIT/X11 license (HPND and MIT-style terms of its contributors,
// see the COPYING file of xkeyboard-config). Each table records the keysyms
// of one layout as Unicode characters.

package hid
`, os.Getenv("XKB_DATA_VERSION"), os.Getenv("XKBCOMMON_VERSION"))

	var out bytes.Buffer
	out.WriteString(header)
	out.WriteString(`
// generatedLayouts holds the layouts read from xkeyboard-config, in the
// order LayoutIDs lists them.
var generatedLayouts = []struct {
	id   string
	keys []keyDef
}{
`)
	for _, l := range gen {
		fmt.Fprintf(&out, "\t{%q, xkb%sKeys},\n", l.spec.id, goName(l.spec.id))
	}
	out.WriteString("}\n")
	for _, l := range gen {
		writeTable(&out, l)
	}
	writeFile("layout_gen.go", out.Bytes())

	out.Reset()
	out.WriteString(header)
	out.WriteString(`
// xkbLayouts holds the xkb version of each hand-written layout, so the tests
// can hold the two against each other.
var xkbLayouts = map[string][]keyDef{
`)
	for _, l := range cmp {
		fmt.Fprintf(&out, "\t%q: xkb%sKeys,\n", l.spec.id, goName(l.spec.id))
	}
	out.WriteString("}\n")
	for _, l := range cmp {
		writeTable(&out, l)
	}
	out.WriteString(`
// xkbCompose holds what the X11 compose table of en_US.UTF-8 types for each
// supported dead key followed by a letter or Space. A pair absent here types
// nothing composed.
var xkbCompose = map[rune]map[rune]rune{
`)
	accents := make([]rune, 0, len(compose))
	for a := range compose {
		accents = append(accents, a)
	}
	sort.Slice(accents, func(i, j int) bool { return accents[i] < accents[j] })
	for _, a := range accents {
		fmt.Fprintf(&out, "\t%s: {", quote(a))
		bases := make([]rune, 0, len(compose[a]))
		for b := range compose[a] {
			bases = append(bases, b)
		}
		sort.Slice(bases, func(i, j int) bool { return bases[i] < bases[j] })
		for i, b := range bases {
			if i > 0 {
				out.WriteString(", ")
			}
			fmt.Fprintf(&out, "%s: %s", quote(b), quote(compose[a][b]))
		}
		out.WriteString("},\n")
	}
	out.WriteString("}\n")
	writeFile("layout_xkb_test.go", out.Bytes())

	for _, l := range append(gen, cmp...) {
		if len(l.skipped) > 0 {
			fmt.Fprintf(os.Stderr, "%s: left out %s\n", l.spec.id, strings.Join(l.skipped, ", "))
		}
	}
}

// keyRank orders the places one character can be typed from. The table keeps
// a character only at its best place, so each character has one key and the
// table reads plainly.
type keyRank [5]int

func (a keyRank) less(b keyRank) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// rank prefers, in this order:
//
//   - a level without AltGr over one with it.
//   - a key of the main block over the ISO, JIS and ABNT2 extras, which a
//     host reads differently by OS: Windows US types \ on usage 100, and xkb
//     us types <. An extra still beats AltGr, so an ISO layout keeps < on
//     the key left of Z.
//   - a punctuation key over a digit and a digit over a letter. xkb gives
//     the letters and digits of most Latin layouts a shared set of AltGr
//     extras (@ on Q, [ on 8); Windows has only the layout's own, which sit
//     mostly on the punctuation keys (es and it type [ with AltGr on the key
//     right of P).
//   - no Shift over Shift, then the earlier key in keyOrder.
func rank(row, level int) keyRank {
	usage := keyOrder[row].usage
	extra, group := 0, 0
	switch {
	case usage >= 100:
		extra = 1
	case usage <= 29:
		group = 2
	case usage <= 39:
		group = 1
	}
	return keyRank{level >> 1, extra, group, level & 1, row}
}

// readLayout compiles one layout and records what each key types at each
// level, keeping each character only at its best place (see rank).
func readLayout(ctx *C.struct_xkb_context, spec layoutSpec) genLayout {
	rules, model := C.CString("evdev"), C.CString("pc105")
	layout, variant, options := C.CString(spec.layout), C.CString(spec.variant), C.CString("")
	defer C.free(unsafe.Pointer(rules))
	defer C.free(unsafe.Pointer(model))
	defer C.free(unsafe.Pointer(layout))
	defer C.free(unsafe.Pointer(variant))
	defer C.free(unsafe.Pointer(options))
	names := C.struct_xkb_rule_names{rules: rules, model: model, layout: layout, variant: variant, options: options}

	keymap := C.xkb_keymap_new_from_names(ctx, &names, C.XKB_KEYMAP_COMPILE_NO_FLAGS)
	if keymap == nil {
		fail("%s: cannot compile xkb layout %s(%s)", spec.id, spec.layout, spec.variant)
	}
	defer C.xkb_keymap_unref(keymap)
	if n := C.xkb_keymap_num_layouts(keymap); n != 1 {
		fail("%s: keymap has %d layouts, want 1", spec.id, n)
	}
	l := genLayout{spec: spec, name: C.GoString(C.xkb_keymap_layout_get_name(keymap, 0))}

	// AltGr exists only where Right Alt is a level 3 shift. Elsewhere it is
	// Alt, and Alt with a key types nothing.
	levels := 2
	if sym(keymap, evdevRAlt, false, false) == keysymLevel3 {
		levels = 4
	}

	type cell struct{ row, level int }
	best := make(map[rune]cell)
	for i, key := range keyOrder {
		row := keyRow{usage: key.usage}
		for level := 0; level < levels; level++ {
			ks := sym(keymap, key.evdev, level&1 != 0, level&2 != 0)
			r, ok := keysymRune(ks)
			if !ok {
				if ks != 0 && ks != keysymVoid && ks != keysymSpace {
					l.skipped = append(l.skipped, fmt.Sprintf("usage %d level %d %s", key.usage, level+1, keysymName(ks)))
				}
				continue
			}
			row.levels[level] = r
			if b, seen := best[r]; !seen || rank(i, level).less(rank(b.row, b.level)) {
				best[r] = cell{i, level}
			}
		}
		l.rows = append(l.rows, row)
	}
	for i := range l.rows {
		for level, r := range l.rows[i].levels {
			if r != 0 && best[r] != (cell{i, level}) {
				l.rows[i].levels[level] = 0
			}
		}
	}
	return l
}

// sym returns the keysym a key types with the given modifiers held. It
// presses them the way a host sees the report, Right Alt first.
func sym(keymap *C.struct_xkb_keymap, evdev int, shift, altGr bool) uint32 {
	state := C.xkb_state_new(keymap)
	defer C.xkb_state_unref(state)
	if altGr {
		C.xkb_state_update_key(state, C.xkb_keycode_t(evdevRAlt+evdevOffset), C.XKB_KEY_DOWN)
	}
	if shift {
		C.xkb_state_update_key(state, C.xkb_keycode_t(evdevShift+evdevOffset), C.XKB_KEY_DOWN)
	}
	return uint32(C.xkb_state_key_get_one_sym(state, C.xkb_keycode_t(evdev+evdevOffset)))
}

// keysymRune returns the character a keysym types, or the flagged accent of
// a supported dead key. Control characters, keysyms with no character and
// unsupported dead keys report false.
func keysymRune(ks uint32) (rune, bool) {
	if ks == 0 {
		return 0, false
	}
	if accent, ok := deadAccents[ks]; ok {
		return accent | deadFlag, true
	}
	r := rune(C.xkb_keysym_to_utf32(C.xkb_keysym_t(ks)))
	if r == 0 || unicode.IsControl(r) || r == ' ' {
		return 0, false
	}
	return r, true
}

func keysymName(ks uint32) string {
	buf := make([]byte, 64)
	n := C.xkb_keysym_get_name(C.xkb_keysym_t(ks), (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf)))
	if n < 0 {
		return fmt.Sprintf("0x%x", ks)
	}
	return string(buf[:n])
}

// readCompose feeds each supported dead key, then each ASCII letter and
// Space, through the X11 compose table of en_US.UTF-8.
func readCompose(ctx *C.struct_xkb_context) map[rune]map[rune]rune {
	locale := C.CString("en_US.UTF-8")
	defer C.free(unsafe.Pointer(locale))
	table := C.xkb_compose_table_new_from_locale(ctx, locale, C.XKB_COMPOSE_COMPILE_NO_FLAGS)
	if table == nil {
		fail("cannot load the en_US.UTF-8 compose table")
	}
	defer C.xkb_compose_table_unref(table)

	bases := []rune{' '}
	for r := 'a'; r <= 'z'; r++ {
		bases = append(bases, r, unicode.ToUpper(r))
	}
	result := make(map[rune]map[rune]rune)
	for ks, accent := range deadAccents {
		m := make(map[rune]rune)
		for _, base := range bases {
			state := C.xkb_compose_state_new(table, C.XKB_COMPOSE_STATE_NO_FLAGS)
			C.xkb_compose_state_feed(state, C.xkb_keysym_t(ks))
			C.xkb_compose_state_feed(state, C.xkb_keysym_t(C.xkb_utf32_to_keysym(C.uint32_t(base))))
			if C.xkb_compose_state_get_status(state) == C.XKB_COMPOSE_COMPOSED {
				buf := make([]byte, 16)
				n := C.xkb_compose_state_get_utf8(state, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf)))
				if rs := []rune(string(buf[:n])); len(rs) == 1 {
					m[base] = rs[0]
				}
			}
			C.xkb_compose_state_unref(state)
		}
		result[accent] = m
	}
	return result
}

func writeTable(out *bytes.Buffer, l genLayout) {
	variant := l.spec.layout
	if l.spec.variant != "" {
		variant += "(" + l.spec.variant + ")"
	}
	fmt.Fprintf(out, "\n// xkb%sKeys is %s, xkb %s.\nvar xkb%sKeys = []keyDef{\n", goName(l.spec.id), l.name, variant, goName(l.spec.id))
	for _, row := range l.rows {
		last := -1
		for level, r := range row.levels {
			if r != 0 {
				last = level
			}
		}
		if last < 0 {
			continue
		}
		fmt.Fprintf(out, "\t{%d, [4]rune{", row.usage)
		for i, r := range row.levels[:last+1] {
			if i > 0 {
				out.WriteString(", ")
			}
			out.WriteString(quote(r))
		}
		out.WriteString("}},\n")
	}
	out.WriteString("}\n")
}

// quote writes a rune as Go source. A dead key becomes "accent | deadFlag",
// a constant, so the tables are static data rather than code run at start; a
// combining mark, which shows only on a base letter, is escaped.
func quote(r rune) string {
	if r == 0 {
		return "0"
	}
	if r&deadFlag != 0 {
		return quote(r&^deadFlag) + " | deadFlag"
	}
	if unicode.Is(unicode.M, r) || !unicode.IsPrint(r) {
		return strconv.QuoteRuneToASCII(r)
	}
	return strconv.QuoteRune(r)
}

// goName turns a layout ID into the part of a Go identifier: "fr-be" is
// "FrBe".
func goName(id string) string {
	var b strings.Builder
	for _, part := range strings.Split(id, "-") {
		b.WriteString(strings.ToUpper(part[:1]) + part[1:])
	}
	return b.String()
}

func writeFile(name string, src []byte) {
	formatted, err := format.Source(src)
	if err != nil {
		fail("%s: %v", name, err)
	}
	if err := os.WriteFile(name, formatted, 0o644); err != nil {
		fail("%v", err)
	}
}

func fail(f string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen_layouts: "+f+"\n", args...)
	os.Exit(1)
}
