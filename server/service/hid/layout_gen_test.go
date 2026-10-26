package hid

import (
	"os"
	"reflect"
	"regexp"
	"sort"
	"testing"
)

func TestLayoutIDsAreUnique(t *testing.T) {
	seen := make(map[string]bool)
	for _, id := range LayoutIDs() {
		if seen[id] {
			t.Errorf("layout ID %q is listed twice", id)
		}
		seen[id] = true
	}
	for alias, id := range layoutAliases {
		if seen[alias] {
			t.Errorf("alias %q is also a layout ID", alias)
		}
		if !seen[id] {
			t.Errorf("alias %q names %q, which is no layout", alias, id)
		}
		layout, err := GetLayout(alias)
		if err != nil || layout.ID != id {
			t.Errorf("GetLayout(%q) = %v, %v, want %s", alias, layout, err, id)
		}
	}
}

// The web client lists the layouts in its own file. A layout the server has
// and the menu lacks cannot be chosen, and one the menu has and the server
// lacks fails every paste.
func TestWebListsEveryLayout(t *testing.T) {
	src, err := os.ReadFile("../../../web/src/jotai/paste.ts")
	if err != nil {
		t.Fatalf("mount the repository root, not server/: %v", err)
	}
	list := regexp.MustCompile(`(?s)pasteLayoutIds = \[(.*?)\]`).FindSubmatch(src)
	if list == nil {
		t.Fatal("web/src/jotai/paste.ts has no pasteLayoutIds")
	}
	var web []string
	for _, m := range regexp.MustCompile(`'([^']+)'`).FindAllSubmatch(list[1], -1) {
		web = append(web, string(m[1]))
	}
	server := LayoutIDs()
	sort.Strings(web)
	sort.Strings(server)
	if !reflect.DeepEqual(web, server) {
		t.Errorf("web lists %v, the server has %v", web, server)
	}
}

// Every pair deadCompose composes must compose the same on Linux.
func TestDeadComposeAgreesWithX11(t *testing.T) {
	for accent, pairs := range deadCompose {
		x11, ok := xkbCompose[accent]
		if !ok {
			t.Errorf("dead %q is not in the X11 compose results", accent)
			continue
		}
		bases, results := []rune(pairs[0]), []rune(pairs[1])
		if len(bases) != len(results) {
			t.Errorf("dead %q: %d bases, %d results", accent, len(bases), len(results))
			continue
		}
		for i, base := range bases {
			if got := x11[base]; got != results[i] {
				t.Errorf("dead %q + %q: deadCompose says %q, X11 composes %q", accent, base, results[i], got)
			}
		}
	}
}

// A generated layout types an accent alone with its dead key and Space only
// where the X11 compose table types that accent.
func TestGeneratedDeadKeyAndSpaceTypeTheAccent(t *testing.T) {
	for _, def := range generatedLayouts {
		layout := layouts[def.id]
		for accent, key := range layout.deadKeys {
			strokes, ok := layout.Keystrokes(accent)
			if !ok || len(strokes) != 2 || strokes[0] != key {
				continue
			}
			if got := xkbCompose[accent][' ']; got != accent {
				t.Errorf("%s: dead %q and Space type %q on Linux", def.id, accent, got)
			}
		}
	}
}

// linuxHost types key presses the way a Linux host with the layout of table
// does: a dead key waits for the next key and composes with it through the
// X11 compose table.
func linuxHost(t *testing.T, table []keyDef, strokes []Char) string {
	t.Helper()
	keys := make(map[Char]rune)
	for _, def := range table {
		for level, r := range def.levels {
			if r != 0 {
				keys[Char{levelModifiers[level], def.code}] = r
			}
		}
	}
	keys[Char{0, usageSpace}] = ' '
	keys[Char{0, usageEnter}] = '\n'
	keys[Char{0, usageTab}] = '\t'

	var out []rune
	var pending rune
	for _, key := range strokes {
		r, ok := keys[key]
		if !ok {
			t.Fatalf("key %+v types nothing", key)
		}
		if pending != 0 {
			composed, ok := xkbCompose[pending][r]
			if !ok {
				t.Fatalf("dead %q + %q composes nothing", pending, r)
			}
			out = append(out, composed)
			pending = 0
			continue
		}
		if r&deadFlag != 0 {
			pending = r &^ deadFlag
			continue
		}
		out = append(out, r)
	}
	if pending != 0 {
		t.Fatalf("dead %q left pending", pending)
	}
	return string(out)
}

// Every character a generated layout claims to type comes back from a host
// with that layout, and so does a sample text of letters, digits, AltGr
// characters and accented letters.
func TestGeneratedLayoutsRoundTrip(t *testing.T) {
	samples := map[string]string{
		"am":               "Բարեւ աշխարհ, ֆ՞ «»",
		"ar":               "مرحبا بالعالم 123 ١٢٣",
		"bepo":             "Ça été dur, à l'œuf: 42 € @ _ ů ê ž",
		"cz":               "Příliš žluťoučký kůň úpěl ďábelské ódy 123 @ ľ ő",
		"da":               "Æblegrød på øen 42 € @ ä ñ",
		"de-ch":            "Grüezi Zürich, ça va? 42 € @ ñ",
		"en-us-altgr-intl": "naïve café, señor 42 € @ ß",
		"en-us-colemak":    "The quick brown fox jumps over 123 @ ~",
		"et":               "Jõgeva ülikool, šokolaad ž 42 € @",
		"fo":               "Føroyar ðá 42 € @ é",
		"fr-be":            "Ça été à côté, où ? 42 € @ ñ",
		"fr-ca":            "Ça été à côté, où? 42 @ ë",
		"fr-ch":            "ça été à côté, Grüße 42 € @ ñ",
		"hr":               "Čšćđž ČŠĆĐŽ 42 € @ ő",
		"hu":               "Árvíztűrő tükörfúrógép 42 € @",
		"is":               "Þjóðin á Íslandi 42 € @ ö",
		"lt":               "Ąžuolas ėjo į šilą ū 42 € @",
		"lv":               "Ģērbjot čūsku ķēniņš 42 € @",
		"mk":               "Здраво свету, ѓ ќ ѕ 42",
		"nl":               "Één café, ëxtra 42 € @",
		"no":               "Blåbærsyltetøy 42 € @ é",
		"pl":               "Zażółć gęślą jaźń 42 € @",
		"pt":               "Ação é côncavo à 42 € @ ñ",
		"sl":               "Čšž ČŠŽ đć 42 € @ ő",
		"th":               "สวัสดีครับ ๑๒๓",
		"tr":               "Iğdır şehri, İstanbul ğüşıöç 42 € @",
	}

	for _, def := range generatedLayouts {
		layout := layouts[def.id]
		sample, ok := samples[def.id]
		if !ok {
			t.Errorf("%s: no sample text", def.id)
			continue
		}

		var chars []rune
		for r := range layout.direct {
			chars = append(chars, r)
		}
		for r := range layout.composed {
			chars = append(chars, r)
		}
		for r := range layout.deadKeys {
			chars = append(chars, r)
		}
		chars = append(chars, []rune(sample)...)

		for _, r := range chars {
			strokes, ok := layout.Keystrokes(r)
			if !ok {
				if containsRune(sample, r) {
					t.Errorf("%s: cannot type %q of the sample", def.id, r)
				}
				continue
			}
			if got := linuxHost(t, def.keys, strokes); got != string(r) {
				t.Errorf("%s: %q types %q", def.id, r, got)
			}
		}

		plan := planPaste(sample, layout)
		if plan.untypeableCount != 0 {
			t.Errorf("%s: sample has untypeable %+v", def.id, plan.untypeable)
			continue
		}
		var strokes []Char
		for _, step := range plan.steps {
			strokes = append(strokes, step.strokes...)
		}
		if got := linuxHost(t, def.keys, strokes); got != sample {
			t.Errorf("%s: sample types %q, want %q", def.id, got, sample)
		}
	}
}

// TestHandWrittenLayoutsAgainstXkb holds each hand-written table against the
// same layout in xkeyboard-config. The hand-written tables follow Windows,
// and they stay: where the two differ, Windows is what the tables were
// written from and what TestLayoutKeystrokes pins, and changing a layout
// ID's behaviour could break a host that works today. Every difference is listed
// here with the reason, so a regenerated xkb table that moves a character
// shows up as a failure to read, not as a silent change.
//
// Characters only xkb types are not listed: xkb gives most Latin layouts a
// set of AltGr extras (æ, ł, →, and so on) that Windows lacks.
func TestHandWrittenLayoutsAgainstXkb(t *testing.T) {
	differences := map[string]string{
		// Windows: AltGr+4 is a dead ~. xkb: AltGr+4 types ~, and the dead
		// ~ is AltGr and the ñ key. On Linux the hand table types ~ with a
		// trailing space and ã as "~a".
		"es": "~ÃÕãõ",
		// Windows: AltGr+7 is a dead ` and AltGr+é a dead ~. xkb: AltGr+7
		// types `, the dead ` is AltGr and the * key, ~ is Shift and ²,
		// and there is no dead ~, so ñ ã õ exist on Windows only.
		"fr": "`~ÀÈÌÒÙìòÃÑÕãñõ",
		// Windows: ´ ` ~ ^ ¨ are dead keys alone. xkb also types them
		// directly with AltGr, which Windows does not have; ¬ is AltGr+6
		// on Windows and AltGr and the ' key on xkb.
		"pt-br": "^`~¨¬´",
		// Windows: ´ is the dead key alone, € is AltGr+E. xkb types ´
		// with AltGr and the ' key, and puts € on AltGr+5 as well.
		"se": "´€",
		// fi is an alias of se. xkb fi puts | on more keys than the ISO
		// key, Windows only there.
		"fi": "|",
	}

	for id, table := range xkbLayouts {
		handID := id
		if alias, ok := layoutAliases[id]; ok {
			handID = alias
		}
		hand, xkb := layouts[handID], buildLayout(id, table)

		chars := make(map[rune]bool)
		for r := range hand.direct {
			chars[r] = true
		}
		for r := range hand.composed {
			chars[r] = true
		}
		for r := range hand.deadKeys {
			chars[r] = true
		}
		var differ []rune
		for r := range chars {
			want, _ := hand.Keystrokes(r)
			got, ok := xkb.Keystrokes(r)
			if !ok || !reflect.DeepEqual(got, want) {
				differ = append(differ, r)
			}
		}
		sort.Slice(differ, func(i, j int) bool { return differ[i] < differ[j] })
		want := []rune(differences[id])
		sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
		if string(differ) != string(want) {
			t.Errorf("%s: the hand-written and xkb tables differ on %q, the list says %q", id, string(differ), string(want))
		}
	}
}
