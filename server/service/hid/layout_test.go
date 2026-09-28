package hid

import (
	"reflect"
	"testing"
)

// legacyUSCharMap is the US table the MCP and PicoClaw type actions used
// before the layouts were added. GetCharMap("") must still return it.
var legacyUSCharMap = map[rune]Char{
	'a': {0, 4}, 'b': {0, 5}, 'c': {0, 6}, 'd': {0, 7}, 'e': {0, 8},
	'f': {0, 9}, 'g': {0, 10}, 'h': {0, 11}, 'i': {0, 12}, 'j': {0, 13},
	'k': {0, 14}, 'l': {0, 15}, 'm': {0, 16}, 'n': {0, 17}, 'o': {0, 18},
	'p': {0, 19}, 'q': {0, 20}, 'r': {0, 21}, 's': {0, 22}, 't': {0, 23},
	'u': {0, 24}, 'v': {0, 25}, 'w': {0, 26}, 'x': {0, 27}, 'y': {0, 28},
	'z': {0, 29},
	'A': {2, 4}, 'B': {2, 5}, 'C': {2, 6}, 'D': {2, 7}, 'E': {2, 8},
	'F': {2, 9}, 'G': {2, 10}, 'H': {2, 11}, 'I': {2, 12}, 'J': {2, 13},
	'K': {2, 14}, 'L': {2, 15}, 'M': {2, 16}, 'N': {2, 17}, 'O': {2, 18},
	'P': {2, 19}, 'Q': {2, 20}, 'R': {2, 21}, 'S': {2, 22}, 'T': {2, 23},
	'U': {2, 24}, 'V': {2, 25}, 'W': {2, 26}, 'X': {2, 27}, 'Y': {2, 28},
	'Z': {2, 29},
	'1': {0, 30}, '2': {0, 31}, '3': {0, 32}, '4': {0, 33}, '5': {0, 34},
	'6': {0, 35}, '7': {0, 36}, '8': {0, 37}, '9': {0, 38}, '0': {0, 39},
	'!': {2, 30}, '@': {2, 31}, '#': {2, 32}, '$': {2, 33}, '%': {2, 34},
	'^': {2, 35}, '&': {2, 36}, '*': {2, 37}, '(': {2, 38}, ')': {2, 39},
	'\n': {0, 40}, '\t': {0, 43}, ' ': {0, 44},
	'-': {0, 45}, '=': {0, 46}, '[': {0, 47}, ']': {0, 48}, '\\': {0, 49},
	';': {0, 51}, '\'': {0, 52}, '`': {0, 53}, ',': {0, 54}, '.': {0, 55}, '/': {0, 56},
	'_': {2, 45}, '+': {2, 46}, '{': {2, 47}, '}': {2, 48}, '|': {2, 49},
	':': {2, 51}, '"': {2, 52}, '~': {2, 53}, '<': {2, 54}, '>': {2, 55}, '?': {2, 56},
}

func TestGetCharMapKeepsLegacyUSTable(t *testing.T) {
	for _, id := range []string{"", "en", "us", "EN", "unknown"} {
		if got := GetCharMap(id); !reflect.DeepEqual(got, legacyUSCharMap) {
			t.Errorf("GetCharMap(%q) differs from the legacy US table", id)
		}
	}
}

func TestGetLayout(t *testing.T) {
	for _, id := range LayoutIDs() {
		layout, err := GetLayout(id)
		if err != nil {
			t.Fatalf("GetLayout(%q): %v", id, err)
		}
		if layout.ID != id {
			t.Errorf("GetLayout(%q).ID = %q", id, layout.ID)
		}
	}
	if len(LayoutIDs()) != len(layouts) {
		t.Errorf("LayoutIDs lists %d layouts, the table holds %d", len(LayoutIDs()), len(layouts))
	}
	for _, id := range []string{"", "en", " EN ", "US"} {
		layout, err := GetLayout(id)
		if err != nil || layout.ID != "us" {
			t.Errorf("GetLayout(%q) = %v, %v, want us", id, layout, err)
		}
	}
	if _, err := GetLayout("xx"); err == nil {
		t.Error("GetLayout(xx) should fail")
	}
}

// A key press that types two characters is a typo in a table.
func TestLayoutKeysTypeOneCharacterEach(t *testing.T) {
	for id, layout := range layouts {
		seen := make(map[Char]rune)
		for r, key := range layout.direct {
			if other, ok := seen[key]; ok {
				t.Errorf("%s: %+v types both %q and %q", id, key, other, r)
			}
			seen[key] = r
		}
		for r, key := range layout.deadKeys {
			if other, ok := seen[key]; ok {
				t.Errorf("%s: dead %+v types both %q and %q", id, key, other, r)
			}
			seen[key] = r
		}
	}
}

func TestLayoutTypesPrintableASCII(t *testing.T) {
	// What each layout has no key for. Everything else from space to tilde
	// must be typeable, on its own key or as a dead key and Space.
	missing := map[string]string{
		"it": "`~",
		"ru": "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ@#$^&[]{}'`~<>|",
	}
	for id, layout := range layouts {
		for r := rune(' '); r <= '~'; r++ {
			_, ok := layout.Keystrokes(r)
			wantMissing := containsRune(missing[id], r)
			if ok == wantMissing {
				t.Errorf("%s: Keystrokes(%q) ok = %v, want %v", id, r, ok, !wantMissing)
			}
		}
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

func TestLayoutLetters(t *testing.T) {
	// Every Latin layout types the 26 letters in both cases, uppercase being
	// the lowercase key with Shift.
	for id, layout := range layouts {
		if id == "ru" {
			continue
		}
		for r := 'a'; r <= 'z'; r++ {
			lower, ok := layout.direct[r]
			if !ok || lower.Modifiers != 0 {
				t.Errorf("%s: %q = %+v, %v", id, r, lower, ok)
				continue
			}
			upper := layout.direct[r-'a'+'A']
			if upper != (Char{modShift, lower.Code}) {
				t.Errorf("%s: %q = %+v, want Shift and usage %d", id, r-'a'+'A', upper, lower.Code)
			}
		}
	}
}

func strokes(keys ...Char) []Char {
	return keys
}

func TestLayoutKeystrokes(t *testing.T) {
	space := Char{0, usageSpace}
	tests := []struct {
		layout string
		char   rune
		want   []Char
	}{
		// US
		{"us", 'a', strokes(Char{0, 4})},
		{"us", '~', strokes(Char{modShift, 53})},
		{"us", '\t', strokes(Char{0, usageTab})},

		// UK
		{"uk", '"', strokes(Char{modShift, 31})},
		{"uk", '@', strokes(Char{modShift, 52})},
		{"uk", '£', strokes(Char{modShift, 32})},
		{"uk", '#', strokes(Char{0, 49})},
		{"uk", '\\', strokes(Char{0, 100})},
		{"uk", '€', strokes(Char{modAltGr, 33})},

		// German: Y and Z swap, AltGr symbols, dead ^ ´ `
		{"de", 'z', strokes(Char{0, 28})},
		{"de", 'Y', strokes(Char{modShift, 29})},
		{"de", 'ß', strokes(Char{0, 45})},
		{"de", 'Ü', strokes(Char{modShift, 47})},
		{"de", '@', strokes(Char{modAltGr, 20})},
		{"de", '{', strokes(Char{modAltGr, 36})},
		{"de", ']', strokes(Char{modAltGr, 38})},
		{"de", '\\', strokes(Char{modAltGr, 45})},
		{"de", '|', strokes(Char{modAltGr, 100})},
		{"de", '€', strokes(Char{modAltGr, 8})},
		{"de", 'é', strokes(Char{0, 46}, Char{0, 8})},
		{"de", 'È', strokes(Char{modShift, 46}, Char{modShift, 8})},
		{"de", 'â', strokes(Char{0, 53}, Char{0, 4})},
		{"de", '^', strokes(Char{0, 53}, space)},
		{"de", '´', strokes(Char{0, 46}, space)},
		{"de", '`', strokes(Char{modShift, 46}, space)},

		// French: AZERTY letters, shifted digits, dead ^ ¨, AltGr dead ~ `
		{"fr", 'a', strokes(Char{0, 20})},
		{"fr", 'q', strokes(Char{0, 4})},
		{"fr", 'w', strokes(Char{0, 29})},
		{"fr", 'm', strokes(Char{0, 51})},
		{"fr", ',', strokes(Char{0, 16})},
		{"fr", '1', strokes(Char{modShift, 30})},
		{"fr", 'é', strokes(Char{0, 31})},
		{"fr", '@', strokes(Char{modAltGr, 39})},
		{"fr", '^', strokes(Char{modAltGr, 38})}, // AltGr 9 is not dead, so it wins
		{"fr", 'ê', strokes(Char{0, 47}, Char{0, 8})},
		{"fr", 'Î', strokes(Char{0, 47}, Char{modShift, 12})},
		{"fr", 'ë', strokes(Char{modShift, 47}, Char{0, 8})},
		{"fr", '¨', strokes(Char{modShift, 47}, space)},
		{"fr", 'ñ', strokes(Char{modAltGr, 31}, Char{0, 17})},
		{"fr", '~', strokes(Char{modAltGr, 31}, space)},
		{"fr", 'À', strokes(Char{modAltGr, 36}, Char{modShift, 20})},

		// Spanish: ñ on its own key, dead accents, AltGr dead ~
		{"es", 'ñ', strokes(Char{0, 51})},
		{"es", 'á', strokes(Char{0, 52}, Char{0, 4})},
		{"es", 'Ó', strokes(Char{0, 52}, Char{modShift, 18})},
		{"es", 'ü', strokes(Char{modShift, 52}, Char{0, 24})},
		{"es", 'à', strokes(Char{0, 47}, Char{0, 4})},
		{"es", 'ê', strokes(Char{modShift, 47}, Char{0, 8})},
		{"es", '~', strokes(Char{modAltGr, 33}, space)},
		{"es", 'ã', strokes(Char{modAltGr, 33}, Char{0, 4})},
		{"es", '¿', strokes(Char{modShift, 46})},

		// Italian: no dead keys, Shift and AltGr brackets
		{"it", 'è', strokes(Char{0, 47})},
		{"it", 'é', strokes(Char{modShift, 47})},
		{"it", '[', strokes(Char{modAltGr, 47})},
		{"it", '{', strokes(Char{modShift | modAltGr, 47})},
		{"it", '@', strokes(Char{modAltGr, 51})},
		{"it", '#', strokes(Char{modAltGr, 52})},

		// Brazilian: ç on its own key, dead ~ ^ ´ ` ¨, / on usage 135
		{"pt-br", 'ç', strokes(Char{0, 51})},
		{"pt-br", 'ã', strokes(Char{0, 52}, Char{0, 4})},
		{"pt-br", 'ô', strokes(Char{modShift, 52}, Char{0, 18})},
		{"pt-br", 'á', strokes(Char{0, 47}, Char{0, 4})},
		{"pt-br", 'ü', strokes(Char{modShift, 35}, Char{0, 24})},
		{"pt-br", '/', strokes(Char{0, 135})},
		{"pt-br", '?', strokes(Char{modShift, 135})},
		{"pt-br", '"', strokes(Char{modShift, 53})},

		// Swedish
		{"se", 'å', strokes(Char{0, 47})},
		{"se", 'ö', strokes(Char{0, 51})},
		{"se", '@', strokes(Char{modAltGr, 31})},
		{"se", 'é', strokes(Char{0, 46}, Char{0, 8})},
		{"se", '~', strokes(Char{modAltGr, 48}, space)},

		// Russian
		{"ru", 'й', strokes(Char{0, 20})},
		{"ru", 'Я', strokes(Char{modShift, 29})},
		{"ru", 'ё', strokes(Char{0, 53})},
		{"ru", 'ю', strokes(Char{0, 55})},
		{"ru", '.', strokes(Char{0, 56})},
		{"ru", ',', strokes(Char{modShift, 56})},
		{"ru", '№', strokes(Char{modShift, 32})},
		{"ru", '"', strokes(Char{modShift, 31})},

		// Japanese JIS
		{"ja", '@', strokes(Char{0, 47})},
		{"ja", '"', strokes(Char{modShift, 31})},
		{"ja", '=', strokes(Char{modShift, 45})},
		{"ja", ':', strokes(Char{0, 52})},
		{"ja", '\\', strokes(Char{0, 135})},
		{"ja", '_', strokes(Char{modShift, 135})},
		{"ja", '|', strokes(Char{modShift, 137})},

		// Korean types Latin text as US
		{"ko", '@', strokes(Char{modShift, 31})},
	}

	for _, tt := range tests {
		layout, err := GetLayout(tt.layout)
		if err != nil {
			t.Fatal(err)
		}
		got, ok := layout.Keystrokes(tt.char)
		if !ok || !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: Keystrokes(%q) = %+v, %v, want %+v", tt.layout, tt.char, got, ok, tt.want)
		}
	}
}

func TestLayoutCannotType(t *testing.T) {
	tests := []struct {
		layout string
		chars  string
	}{
		{"us", "éñü€ж"},
		{"uk", "éж"},
		{"de", "ñçжÿ"},
		{"fr", "Éß"},
		{"it", "`~ñ"},
		{"ru", "aZ@é"},
		{"ja", "あ¥"},
		{"ko", "한"},
	}
	for _, tt := range tests {
		layout, err := GetLayout(tt.layout)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range tt.chars {
			if got, ok := layout.Keystrokes(r); ok {
				t.Errorf("%s: Keystrokes(%q) = %+v, want untypeable", tt.layout, r, got)
			}
		}
	}
}

// A composed character is always its layout's dead key followed by a key
// that types the base letter directly.
func TestComposedUseADeadKeyAndALetter(t *testing.T) {
	for id, layout := range layouts {
		deadKeys := make(map[Char]bool)
		for _, key := range layout.deadKeys {
			deadKeys[key] = true
		}
		for r, keys := range layout.composed {
			if len(keys) != 2 || !deadKeys[keys[0]] {
				t.Errorf("%s: %q = %+v, want a dead key and a letter", id, r, keys)
				continue
			}
			if _, direct := layout.direct[r]; direct {
				t.Errorf("%s: %q is both direct and composed", id, r)
			}
		}
		if len(layout.deadKeys) == 0 && len(layout.composed) != 0 {
			t.Errorf("%s: composes without dead keys", id)
		}
	}
}
