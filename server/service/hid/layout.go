package hid

import (
	"fmt"
	"strings"
	"unicode"
)

// Char is one key press: the modifier byte of the keyboard report and the
// HID usage of the key.
type Char struct {
	Modifiers int
	Code      int
}

// Modifier bits of the keyboard report's first byte.
const (
	modShift = 0x02 // Left Shift
	modAltGr = 0x40 // Right Alt, which a layout with a third level reads as AltGr
)

// levelModifiers holds the modifiers that select each level of a key: the key
// alone, with Shift, with AltGr, and with Shift and AltGr.
var levelModifiers = [4]int{0, modShift, modAltGr, modShift | modAltGr}

// Keys that type the same thing on every layout.
const (
	usageEnter = 40
	usageTab   = 43
	usageSpace = 44
)

// deadFlag marks a rune in a key table as a dead key: pressing it types
// nothing, and the next key types the accented letter. The flag sits above
// the Unicode range, so it cannot collide with a real character.
const deadFlag rune = 1 << 30

func dead(r rune) rune {
	return r | deadFlag
}

// keyDef is one physical key: its HID usage and what it types at each level.
// A zero rune means that level types nothing worth sending.
type keyDef struct {
	code   int
	levels [4]rune
}

func k(code int, levels ...rune) keyDef {
	def := keyDef{code: code}
	copy(def.levels[:], levels)
	return def
}

// letters builds the letter keys from a row that gives, for each usage from
// 4 (A) to 29 (Z), the lowercase letter that key types. A '_' leaves the key
// out, for a layout that puts something else there.
func letters(row string) []keyDef {
	defs := make([]keyDef, 0, 26)
	code := 4
	for _, r := range row {
		if r != '_' {
			defs = append(defs, k(code, r, unicode.ToUpper(r)))
		}
		code++
	}
	return defs
}

// with returns base with each override replacing the key of the same usage.
// An override for a usage base does not have is added at the end.
func with(base []keyDef, overrides ...keyDef) []keyDef {
	defs := append([]keyDef(nil), base...)
	for _, override := range overrides {
		replaced := false
		for i := range defs {
			if defs[i].code == override.code {
				defs[i] = override
				replaced = true
				break
			}
		}
		if !replaced {
			defs = append(defs, override)
		}
	}
	return defs
}

// deadCompose lists what each dead key composes with the letter typed after
// it. The pairs are the ones Windows and the X11 compose tables agree on.
var deadCompose = map[rune][2]string{
	'´': {"aeiouyAEIOUY", "áéíóúýÁÉÍÓÚÝ"},
	'`': {"aeiouAEIOU", "àèìòùÀÈÌÒÙ"},
	'^': {"aeiouAEIOU", "âêîôûÂÊÎÔÛ"},
	'¨': {"aeiouyAEIOU", "äëïöüÿÄËÏÖÜ"},
	'~': {"anoANO", "ãñõÃÑÕ"},
}

// Layout maps characters to the key presses that type them on a target whose
// active keyboard layout is this one.
type Layout struct {
	ID string

	// direct holds the characters one key press types.
	direct map[rune]Char
	// deadKeys holds the dead key for each accent. Pressed before Space it
	// types the accent alone.
	deadKeys map[rune]Char
	// composed holds the characters a dead key and a letter type together.
	composed map[rune][]Char
}

// buildLayout turns a key table into a Layout. Where two keys type the same
// character, the one earlier in the table wins, so a table lists letters
// first and the plainer key of a pair before the other.
func buildLayout(id string, defs []keyDef) *Layout {
	l := &Layout{
		ID:       id,
		direct:   make(map[rune]Char),
		deadKeys: make(map[rune]Char),
		composed: make(map[rune][]Char),
	}

	for _, def := range defs {
		for level, r := range def.levels {
			if r == 0 {
				continue
			}
			key := Char{Modifiers: levelModifiers[level], Code: def.code}
			if r&deadFlag != 0 {
				r &^= deadFlag
				if _, ok := l.deadKeys[r]; !ok {
					l.deadKeys[r] = key
				}
				continue
			}
			if _, ok := l.direct[r]; !ok {
				l.direct[r] = key
			}
		}
	}

	l.direct[' '] = Char{0, usageSpace}
	l.direct['\n'] = Char{0, usageEnter}
	l.direct['\t'] = Char{0, usageTab}

	for accent, deadKey := range l.deadKeys {
		pairs, ok := deadCompose[accent]
		if !ok {
			continue
		}
		bases, results := []rune(pairs[0]), []rune(pairs[1])
		for i, base := range bases {
			result := results[i]
			if _, ok := l.direct[result]; ok {
				continue
			}
			baseKey, ok := l.direct[base]
			if !ok {
				continue
			}
			if _, ok := l.composed[result]; !ok {
				l.composed[result] = []Char{deadKey, baseKey}
			}
		}
	}

	return l
}

// Keystrokes returns the key presses that type r, and false when this layout
// cannot type it. A character on its own key takes one press. An accented
// letter takes its dead key and then the letter, and an accent that exists
// only as a dead key takes the dead key and then Space.
func (l *Layout) Keystrokes(r rune) ([]Char, bool) {
	if key, ok := l.direct[r]; ok {
		return []Char{key}, true
	}
	if keys, ok := l.composed[r]; ok {
		return keys, true
	}
	if key, ok := l.deadKeys[r]; ok {
		return []Char{key, {0, usageSpace}}, true
	}
	return nil, false
}

// The key tables follow the Windows layouts of the same names. Where Linux
// differs, it is only in whether a few AltGr accents are dead.

var usLetters = letters("abcdefghijklmnopqrstuvwxyz")

var usKeys = with(usLetters,
	k(53, '`', '~'), k(30, '1', '!'), k(31, '2', '@'), k(32, '3', '#'), k(33, '4', '$'),
	k(34, '5', '%'), k(35, '6', '^'), k(36, '7', '&'), k(37, '8', '*'), k(38, '9', '('),
	k(39, '0', ')'), k(45, '-', '_'), k(46, '=', '+'),
	k(47, '[', '{'), k(48, ']', '}'), k(49, '\\', '|'),
	k(51, ';', ':'), k(52, '\'', '"'),
	k(54, ',', '<'), k(55, '.', '>'), k(56, '/', '?'),
)

// ukKeys is the UK layout. The key beside Enter is sent as usage 49, which
// hosts read the same as the ISO usage 50.
var ukKeys = with(usLetters,
	k(53, '`', '¬'), k(30, '1', '!'), k(31, '2', '"'), k(32, '3', '£'), k(33, '4', '$', '€'),
	k(34, '5', '%'), k(35, '6', '^'), k(36, '7', '&'), k(37, '8', '*'), k(38, '9', '('),
	k(39, '0', ')'), k(45, '-', '_'), k(46, '=', '+'),
	k(47, '[', '{'), k(48, ']', '}'), k(49, '#', '~'),
	k(51, ';', ':'), k(52, '\'', '@'),
	k(100, '\\', '|'), k(54, ',', '<'), k(55, '.', '>'), k(56, '/', '?'),
)

// deKeys is German QWERTZ: Y and Z swap places, and ^, ´ and ` are dead.
var deKeys = with(letters("abcdefghijklmnopqrstuvwxzy"),
	k(20, 'q', 'Q', '@'), k(8, 'e', 'E', '€'), k(16, 'm', 'M', 'µ'),
	k(53, dead('^'), '°'), k(30, '1', '!'), k(31, '2', '"', '²'), k(32, '3', '§', '³'),
	k(33, '4', '$'), k(34, '5', '%'), k(35, '6', '&'), k(36, '7', '/', '{'),
	k(37, '8', '(', '['), k(38, '9', ')', ']'), k(39, '0', '=', '}'),
	k(45, 'ß', '?', '\\'), k(46, dead('´'), dead('`')),
	k(47, 'ü', 'Ü'), k(48, '+', '*', '~'), k(49, '#', '\''),
	k(51, 'ö', 'Ö'), k(52, 'ä', 'Ä'),
	k(100, '<', '>', '|'), k(54, ',', ';'), k(55, '.', ':'), k(56, '-', '_'),
)

// frKeys is French AZERTY: A and Q swap, Z and W swap, M sits right of L,
// and the digits need Shift. ^ and ¨ are dead, and so are AltGr ~ and `.
var frKeys = with(letters("qbcdefghijkl_noparstuvzxyw"),
	k(8, 'e', 'E', '€'), k(16, ',', '?'), k(51, 'm', 'M'),
	k(53, '²'), k(30, '&', '1'), k(31, 'é', '2', dead('~')), k(32, '"', '3', '#'),
	k(33, '\'', '4', '{'), k(34, '(', '5', '['), k(35, '-', '6', '|'),
	k(36, 'è', '7', dead('`')), k(37, '_', '8', '\\'), k(38, 'ç', '9', '^'),
	k(39, 'à', '0', '@'), k(45, ')', '°', ']'), k(46, '=', '+', '}'),
	k(47, dead('^'), dead('¨')), k(48, '$', '£', '¤'), k(49, '*', 'µ'),
	k(52, 'ù', '%'),
	k(100, '<', '>'), k(54, ';', '.'), k(55, ':', '/'), k(56, '!', '§'),
)

// esKeys is Spanish (Spain). `, ^, ´ and ¨ are dead, and so is AltGr ~.
var esKeys = with(usLetters,
	k(8, 'e', 'E', '€'),
	k(53, 'º', 'ª', '\\'), k(30, '1', '!', '|'), k(31, '2', '"', '@'), k(32, '3', '·', '#'),
	k(33, '4', '$', dead('~')), k(34, '5', '%'), k(35, '6', '&', '¬'), k(36, '7', '/'),
	k(37, '8', '('), k(38, '9', ')'), k(39, '0', '='), k(45, '\'', '?'), k(46, '¡', '¿'),
	k(47, dead('`'), dead('^'), '['), k(48, '+', '*', ']'), k(49, 'ç', 'Ç', '}'),
	k(51, 'ñ', 'Ñ'), k(52, dead('´'), dead('¨'), '{'),
	k(100, '<', '>'), k(54, ',', ';'), k(55, '.', ':'), k(56, '-', '_'),
)

// itKeys is Italian. It has no dead keys, and no key for ` or ~.
var itKeys = with(usLetters,
	k(8, 'e', 'E', '€'),
	k(53, '\\', '|'), k(30, '1', '!'), k(31, '2', '"'), k(32, '3', '£'), k(33, '4', '$'),
	k(34, '5', '%'), k(35, '6', '&'), k(36, '7', '/'), k(37, '8', '('), k(38, '9', ')'),
	k(39, '0', '='), k(45, '\'', '?'), k(46, 'ì', '^'),
	k(47, 'è', 'é', '[', '{'), k(48, '+', '*', ']', '}'), k(49, 'ù', '§'),
	k(51, 'ò', 'ç', '@'), k(52, 'à', '°', '#'),
	k(100, '<', '>'), k(54, ',', ';'), k(55, '.', ':'), k(56, '-', '_'),
)

// ptBRKeys is Brazilian ABNT2. ´, `, ~, ^ and ¨ are dead. / and ? sit on the
// extra ABNT2 key left of Right Shift, usage 135.
var ptBRKeys = with(usLetters,
	k(53, '\'', '"'), k(30, '1', '!', '¹'), k(31, '2', '@', '²'), k(32, '3', '#', '³'),
	k(33, '4', '$', '£'), k(34, '5', '%', '¢'), k(35, '6', dead('¨'), '¬'), k(36, '7', '&'),
	k(37, '8', '*'), k(38, '9', '('), k(39, '0', ')'), k(45, '-', '_'), k(46, '=', '+', '§'),
	k(47, dead('´'), dead('`')), k(48, '[', '{', 'ª'), k(49, ']', '}', 'º'),
	k(51, 'ç', 'Ç'), k(52, dead('~'), dead('^')),
	k(100, '\\', '|'), k(54, ',', '<'), k(55, '.', '>'), k(56, ';', ':'), k(135, '/', '?'),
)

// seKeys is Swedish, which Finnish shares. ´, `, ¨, ^ and ~ are dead.
var seKeys = with(usLetters,
	k(8, 'e', 'E', '€'), k(16, 'm', 'M', 'µ'),
	k(53, '§', '½'), k(30, '1', '!'), k(31, '2', '"', '@'), k(32, '3', '#', '£'),
	k(33, '4', '¤', '$'), k(34, '5', '%'), k(35, '6', '&'), k(36, '7', '/', '{'),
	k(37, '8', '(', '['), k(38, '9', ')', ']'), k(39, '0', '=', '}'),
	k(45, '+', '?', '\\'), k(46, dead('´'), dead('`')),
	k(47, 'å', 'Å'), k(48, dead('¨'), dead('^'), dead('~')), k(49, '\'', '*'),
	k(51, 'ö', 'Ö'), k(52, 'ä', 'Ä'),
	k(100, '<', '>', '|'), k(54, ',', ';'), k(55, '.', ':'), k(56, '-', '_'),
)

// ruKeys is Russian ЙЦУКЕН. It types no Latin letters.
var ruKeys = with(letters("фисвуапршолдьтщзйкыегмцчня"),
	k(53, 'ё', 'Ё'), k(30, '1', '!'), k(31, '2', '"'), k(32, '3', '№'), k(33, '4', ';'),
	k(34, '5', '%'), k(35, '6', ':'), k(36, '7', '?'), k(37, '8', '*'), k(38, '9', '('),
	k(39, '0', ')'), k(45, '-', '_'), k(46, '=', '+'),
	k(47, 'х', 'Х'), k(48, 'ъ', 'Ъ'), k(49, '\\', '/'),
	k(51, 'ж', 'Ж'), k(52, 'э', 'Э'),
	k(54, 'б', 'Б'), k(55, 'ю', 'Ю'), k(56, '.', ','),
)

// jaKeys is Japanese JIS 106/109 with the IME off. The Yen key (usage 137)
// and the Ro key (usage 135) both type a backslash; the table takes it from
// Ro. Kana need the IME and are not typed.
var jaKeys = with(usLetters,
	k(30, '1', '!'), k(31, '2', '"'), k(32, '3', '#'), k(33, '4', '$'), k(34, '5', '%'),
	k(35, '6', '&'), k(36, '7', '\''), k(37, '8', '('), k(38, '9', ')'), k(39, '0'),
	k(45, '-', '='), k(46, '^', '~'), k(137, 0, '|'),
	k(47, '@', '`'), k(48, '[', '{'),
	k(51, ';', '+'), k(52, ':', '*'), k(49, ']', '}'),
	k(54, ',', '<'), k(55, '.', '>'), k(56, '/', '?'), k(135, '\\', '_'),
)

// layouts holds every layout by ID. Korean keyboards carry the US layout for
// Latin text; Hangul needs the IME and is not typed.
var layouts = func() map[string]*Layout {
	m := make(map[string]*Layout)
	for _, def := range []struct {
		id   string
		keys []keyDef
	}{
		{"us", usKeys},
		{"uk", ukKeys},
		{"de", deKeys},
		{"fr", frKeys},
		{"es", esKeys},
		{"it", itKeys},
		{"pt-br", ptBRKeys},
		{"se", seKeys},
		{"ru", ruKeys},
		{"ja", jaKeys},
		{"ko", usKeys},
	} {
		m[def.id] = buildLayout(def.id, def.keys)
	}
	return m
}()

// LayoutIDs lists the IDs GetLayout accepts, besides the aliases.
func LayoutIDs() []string {
	return []string{"us", "uk", "de", "fr", "es", "it", "pt-br", "se", "ru", "ja", "ko"}
}

// GetLayout returns the layout with the given ID. An empty ID and "en", the
// name an older web client sends, both mean US.
func GetLayout(id string) (*Layout, error) {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" || id == "en" {
		id = "us"
	}
	layout, ok := layouts[id]
	if !ok {
		return nil, fmt.Errorf("unknown keyboard layout %q", id)
	}
	return layout, nil
}

// GetCharMap returns the characters one key press types on the layout, or on
// US for an unknown one.
func GetCharMap(lang string) map[rune]Char {
	layout, err := GetLayout(lang)
	if err != nil {
		layout = layouts["us"]
	}
	m := make(map[rune]Char, len(layout.direct))
	for r, key := range layout.direct {
		m[r] = key
	}
	return m
}
