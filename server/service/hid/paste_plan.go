package hid

import (
	"unicode"

	"golang.org/x/text/unicode/norm"

	"NanoKVM-Server/proto"
)

// maxUntypeableReported bounds the list of characters a layout cannot type.
// The count beside it stays exact, so a text in the wrong script reports how
// much it holds without a response the size of the text.
const maxUntypeableReported = 100

// pasteStep is one character of the text and the key presses that type it.
type pasteStep struct {
	index   int // offset of the character in the text, in code points
	strokes []Char
}

// pastePlan is a text turned into key presses for one layout.
type pastePlan struct {
	steps           []pasteStep
	keystrokes      int
	untypeable      []proto.PasteUntypeable
	untypeableCount int
}

// planPaste turns text into the key presses that type it on layout.
//
// A line ending of CRLF, CR or LF becomes one Enter, and a tab becomes Tab. A
// letter followed by combining marks, as a macOS file name carries é, is
// composed first, so it is typed when the layout can type the composed
// letter. Every character the layout cannot type is reported with its offset,
// line and column instead of being dropped silently.
func planPaste(text string, layout *Layout) pastePlan {
	runes := []rune(text)
	plan := pastePlan{steps: make([]pasteStep, 0, len(runes))}

	line, column := 1, 1
	for i := 0; i < len(runes); {
		r := runes[i]
		width := 1

		switch r {
		case '\r', '\n':
			if r == '\r' && i+1 < len(runes) && runes[i+1] == '\n' {
				width = 2
			}
			plan.add(i, []Char{{0, usageEnter}})
			i += width
			line++
			column = 1
			continue
		}

		// Compose a letter with the combining marks after it.
		end := i + 1
		for end < len(runes) && unicode.Is(unicode.Mn, runes[end]) {
			end++
		}
		if end > i+1 {
			if composed := []rune(norm.NFC.String(string(runes[i:end]))); len(composed) == 1 {
				r = composed[0]
				width = end - i
			}
		}

		if strokes, ok := layout.Keystrokes(r); ok {
			plan.add(i, strokes)
		} else {
			plan.untypeableCount++
			if len(plan.untypeable) < maxUntypeableReported {
				plan.untypeable = append(plan.untypeable, proto.PasteUntypeable{
					Index:  i,
					Line:   line,
					Column: column,
					Char:   string(runes[i : i+width]),
				})
			}
		}
		i += width
		column += width
	}

	return plan
}

func (p *pastePlan) add(index int, strokes []Char) {
	p.steps = append(p.steps, pasteStep{index: index, strokes: strokes})
	p.keystrokes += len(strokes)
}
